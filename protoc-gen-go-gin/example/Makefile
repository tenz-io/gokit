.PHONY: protoc-compile
protoc-compile:
	# go install
	protoc -I ./api \
			--openapiv2_out ./api --openapiv2_opt logtostderr=true \
			--openapiv2_opt json_names_for_fields=false \
			--go_out ./api --go_opt=paths=source_relative \
			--go-gin_out ./api --go-gin_opt=paths=source_relative \
			api/product/app/v1/v1.proto

	protoc-go-inject-tag -input=./api/product/app/v1/v1.pb.go


.PHONY: dep
dep:
	go mod tidy -v


.PHONY: run
run:
	go run main.go


.PHONY: login
login:
	curl -v -X POST localhost:8888/login \
	-H "Content-Type: application/json" \
	-d '{"username":"admin","password":"admin"}'


.PHONY: get-articles
get-articles:
	curl -v -X GET "localhost:8888/v1/author/123/articles?title=hello&page=1" \
	-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzE4NjQzNjE3fQ.hIo3MSoU2yCQPQOoRSYaEf4OCp79K7R7yhUDCYZC0kc"


.PHONY: create-article
create-article:
	curl -v -X POST localhost:8888/v1/author/123/articles \
	-H "Content-Type: application/json" \
	-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzE4NjQzNjE3fQ.hIo3MSoU2yCQPQOoRSYaEf4OCp79K7R7yhUDCYZC0kc" \
	-d '{"title":"Hello World","content":"This is a test article"}'


.PHONY: get-image
get-image:
	curl -v -X GET "localhost:8888/v1/images/xyz123?region=uk" \
	-H "Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzE4NjQzNjE3fQ.hIo3MSoU2yCQPQOoRSYaEf4OCp79K7R7yhUDCYZC0kc"


.PHONY: upload-image
upload-image:
	curl -v -X POST --location 'localhost:8888/v1/images/xyz123?region=uk' \
    --header 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwiZXhwIjoxNzE4NjQzNjE3fQ.hIo3MSoU2yCQPQOoRSYaEf4OCp79K7R7yhUDCYZC0kc' \
    --form 'image=@"/path/to/file"' \
    --form 'filename="cat.jpg"'
