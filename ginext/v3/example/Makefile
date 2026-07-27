
.PHONY run:
run:
	go run main.go

.PHONY test:
test:
	@echo "ping"
	curl -s localhost:8080/ping -vvv
	@echo "\n\ncreate user"
	curl -s -X POST 'localhost:8080/user/1?limit=5&offset=0' -H 'Authorization: Bearer abc' -H 'Content-Type: application/json' -d '{"name":"bob"}' -vvv
