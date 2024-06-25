package main

import (
	"context"
	"fmt"
	v1pb "validator-example/api/example/v1"
)

func main() {

	req := &v1pb.LoginRequest{
		Username: "alice",
		Password: "123456",
	}

	if err := req.Validate(context.Background()); err != nil {
		fmt.Println(err)
		return
	}

}
