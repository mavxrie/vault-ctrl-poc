package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/vault/api"
)

func main() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}

	server := http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/create", createHandler(client))
	http.HandleFunc("/login", loginHandler(client))
	http.HandleFunc("/secret", secretHandler(client))
	http.HandleFunc("/delete", deleteHandler(client))

	// start the server
	fmt.Println("Listening on :8080...")
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
