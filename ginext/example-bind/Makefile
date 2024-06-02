
.PHONY run:
run:
	go run main.go


.PHONY test:
test:
	@echo "Running tests"
	curl --location 'http://localhost:8080/user/123/?offset=1' \
    --header 'Authorization: abc123' \
    --header 'Content-Type: application/json' \
    --data '{ "name": "eddie",  "profile": "I am m eddie" }'