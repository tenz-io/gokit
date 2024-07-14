package main

import (
	"fmt"

	v1pb "validator-example/api/example/v1"
)

func main() {

	var loginReq *v1pb.LoginRequest

	loginReq = &v1pb.LoginRequest{
		Username: "alice",
		Password: "123456",
	}

	if err := loginReq.Validate(); err != nil {
		fmt.Println(err)
		return
	}

	loginReq = &v1pb.LoginRequest{}
	if err := loginReq.Validate(); err != nil {
		panic(err)
	}

}
