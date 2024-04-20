
.PHONY docker-up:
docker-up:
	docker-compose up -d


.PHONY docker-down:
docker-down:
	docker-compose down

cmd := etcdctl get --prefix /services/echo

.PHONY docker-etcdctl:
docker-etcdctl:
	@echo "cmd: $(cmd)"
	docker-compose exec etcd $(cmd)