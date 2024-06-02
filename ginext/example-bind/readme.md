
## test restful api

```shell
curl --location 'http://localhost:8080/user/123/?offset=1' \
--header 'Authorization: abc123' \
--header 'Content-Type: application/json' \
--data '{
    "name": "eddie",
    "profile": "I am eddie"
}'
```

## test file upload

```shell
curl --location --request PUT 'http://localhost:8080/search' \
--header 'Authorization: abc123' \
--form 'image=@"/path/to/image.jpg"' \
--form 'bbox="12345678"'
```