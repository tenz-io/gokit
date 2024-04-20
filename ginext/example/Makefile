
.PHONY run:
run:
	go run main.go


.PHONY test:
test:
	@echo "Running tests"
	curl localhost:8080/ping -vvv
	@echo "\n\nRunning panic test"
	curl localhost:8080/panic -vvv