package notiongo

type Client struct {
	CreateDatabase func(request CreateDatabaseRequest) (CreateDatabaseResponse, error)
}
