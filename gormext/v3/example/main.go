// Example: gormext/v3 — traffic + metrics + errorLog + slowLog via GORM callbacks.
//
// Opens an in-memory sqlite DB, applies a Tracker, then does a create + query
// to demonstrate that every GORM operation transparently flows through the
// registered callbacks. Point it at any real DSN by swapping the gorm.Open call.
package main

import (
	"context"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/tenz-io/gokit/gormext/v3"
	"github.com/tenz-io/gokit/logger/v3"
	"github.com/tenz-io/gokit/monitor/v3"
)

func init() {
	logger.ConfigureWithOpts(
		logger.WithLevel(logger.DebugLevel),
		logger.WithConsole(true),
		logger.WithFilePath("log"),
		logger.WithCaller(true),
		logger.WithCallerSkip(1),
		logger.WithTraffic(true),
	)
}

func main() {
	// In-memory sqlite for a self-contained demo. Swap for mysql.Open(dsn)
	// (see the v2 example-mysql) to run against a real database.
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		log.Fatal("open database error: ", err)
	}

	tracker := gormext.NewTrackerWithOpts(
		gormext.WithEnableTraffic(true),
		gormext.WithEnableMetrics(true),
		gormext.WithEnableErrorLog(true),
		// 开启参数记录,但配 Redactor 把字符串参数 (含密码) 脱敏为 ***,
		// 这样 traffic/慢日志里既能看到"绑了几个参数"又不泄露密码。
		gormext.WithEnableVars(true),
		gormext.WithRedactor(gormext.Redactor(redactPassword)),
		// 1ms floor makes every query here "slow" so the slow-log path fires.
		gormext.WithSlowLogFloor(1*time.Millisecond),
	)
	if err = tracker.Apply(db); err != nil {
		log.Fatal("apply tracker error: ", err)
	}

	// Inject a monitor Exporter so WithEnableMetrics records something. At the
	// request edge this is the single-flight injection; downstream calls reuse it.
	ctx := monitor.Init(context.Background(), "gormext-example")

	if err = db.AutoMigrate(&User{}); err != nil {
		log.Fatal("auto migrate error: ", err)
	}

	if err = Save(ctx, db, &User{Username: "sky", Password: "sky123"}); err != nil {
		log.Printf("save user error: %+v", err)
	}

	user, err := Find(ctx, db, "sky")
	log.Printf("find user: %+v, err: %v", user, err)

	// A miss to demonstrate the error-log + traffic paths on ErrRecordNotFound.
	_, err = Find(ctx, db, "missing")
	log.Printf("find missing error: %+v", err)

	time.Sleep(200 * time.Millisecond) // let traffic.log flush
}

func Find(ctx context.Context, db *gorm.DB, username string) (*User, error) {
	var u User
	if err := db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func Save(ctx context.Context, db *gorm.DB, user *User) error {
	return db.WithContext(ctx).Create(user).Error
}

type User struct {
	ID        int64     `gorm:"primaryKey"`
	Username  string    `gorm:"column:username;unique"`
	Password  string    `gorm:"column:password"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (User) TableName() string { return "user_tab" }

// redactPassword 把 SQL 绑定参数中的字符串值脱敏为 ***,其余类型原样保留。
// 它保护 traffic/错误/慢日志中的密码、token 等敏感字符串不被明文记录。
func redactPassword(sql string, vars []any) (string, []any) {
	out := make([]any, len(vars))
	for i, v := range vars {
		switch v.(type) {
		case string:
			out[i] = "***"
		default:
			out[i] = v
		}
	}
	return sql, out
}
