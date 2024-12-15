package main

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

func main() {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   "c2ac1187-3faa-4734-bb66-ac45d270da92.us-west-1-0.aws.cloud.qdrant.io",
		Port:   6334,
		APIKey: "R5uK-uIpgGMGgQOSnV7a3BQFkrZcSEmyj_diGOymITeu9mQV5w7Q9A",
		UseTLS: true,
	})
	if err != nil {
		panic(err)
	}

	collections, err := client.ListCollections(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(collections)
}