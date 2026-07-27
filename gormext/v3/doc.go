// Package gormext 为 GORM 注入一组可组合的回调层:per-query 流量日志、
// Prometheus 指标、数据库错误日志,以及慢查询检测。
//
// v3 是一次干净的重写,不带任何向后兼容 shim,构建于 logger/v3、
// monitor/v3 与 tracer/v3 之上。它与未改动的 gormext/v2 并存;使用方不会
// 被自动迁移。
//
// 用法:用 NewTrackerWithOpts 构建 Tracker,通过 Tracker.Apply 一次性把回调
// 注册到 *gorm.DB,此后正常使用 gorm (db.WithContext(ctx).First/...) 即可,
// 所有 Query/Create/Update/Delete/Row/Raw 操作都会透明地经过回调。
//
// 快速开始:
//
//	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
//	tracker := gormext.NewTrackerWithOpts(
//	    gormext.WithEnableTraffic(true),
//	    gormext.WithEnableMetrics(true),
//	    gormext.WithEnableErrorLog(true),
//	    // 安全:开启参数记录但配 Redactor 脱敏密码/token 等字符串参数。
//	    gormext.WithEnableVars(true),
//	    gormext.WithRedactor(gormext.Redactor(redactSecrets)),
//	    gormext.WithSlowLogFloor(500*time.Millisecond),
//	)
//	if err := tracker.Apply(db); err != nil { log.Fatal(err) }
//
//	var user User
//	_ = db.WithContext(ctx).First(&user, 1).Error
//
// 行为说明 (与 v2 不同):
//
//   - 安全默认:SQL 绑定参数 (vars) 默认**不**记入 traffic/errorLog/slowLog
//     (含密码、token、手机号等)。需要时 WithEnableVars(true),建议配合
//     WithRedactor 脱敏。SQL 语句本身始终记录。
//   - 错误日志带实际错误:非 not_found 的错误输出 Error 时附 WithError
//     (db.Error),因此线上能看到是超时/断连/唯一键冲突还是语法错误,
//     而非只一句 "db error"。
//   - 统一错误分类:metrics/traffic/errorLog 共用 classify(err) →
//     ok/not_found/err,口径一致。尤其 gorm.ErrRecordNotFound 在三层都
//     视为"正常" (metrics code="0" 不拉高失败率;traffic code="not_found";
//     errorLog 降级 Debug),避免大量正常未命中污染监控与告警。
//   - 选项改名为 WithEnableTraffic / WithEnableMetrics / WithEnableErrorLog /
//     WithEnableVars / WithSlowLogFloor / WithRedactor,与 httpext/v3 的
//     WithEnable* 命名一致。nil option 被安全跳过;负 SlowLogFloor 归一为 0。
//   - traffic 日志改用 logger/v3 的 span API:
//     logger.FromContext(ctx).StartTraffic(cmd).WithTyp(logger.TrafficTypSend),
//     以 rec.End(code, fields...) / rec.EndWithError(err, fields...) 完成。
//     不再使用 v2 的 ReqEntity/RespEntity 表面;不记 result body。
//   - metrics 改用 monitor/v3 的 monitor.Begin / rec.EndWithCode(code)。
//   - 慢查询日志 (SlowLogFloor) 保留:SQL 场景下慢日志会打印具体 SQL
//     (参数受 EnableVars/Redactor 控制),比 monitor 的 latency 直方图更
//     能定位是哪条 SQL 慢。
//   - 覆盖 GORM 六类操作:Query/Create/Update/Delete/Row/Raw。Raw 覆盖
//     db.Exec / db.Raw / 部分迁移 DDL,否则它们不产生任何指标或日志。
//     回调按 ops 表循环注册,回调名带 "gormext:" 前缀避免冲突。
//
// 语义边界 (务必了解):
//
//   - 回调衡量的是 GORM operation 粒度,不一定是"一条物理 SQL":
//     Before("*")/After("*") 会含事务、hook、association、preload;
//     DryRun 等情况下可能根本不执行 SQL。指标语义是 GORM operation
//     latency,而非 database execution latency。
//   - Row/Rows 的语义是"查询派发耗时/派发是否成功",不承诺结果集读取的
//     完整耗时与错误:QueryRowContext 的错误可能延迟到 row.Scan() 才
//     出现,Rows() 的迭代错误不会进入当前 db.Error。
//   - Apply 设计为在 DB 初始化阶段调用恰好一次。gorm 对重名回调只发
//     warn 并保留两份 handler (非替换),因此重复 Apply 会让每个回调
//     被调用两次 (指标/日志翻倍);如需更换配置,重建 db 或先 Remove。
//   - Apply 中途失败会留下半注册状态;实践中失败即 fatal,不到达业务路径。
package gormext
