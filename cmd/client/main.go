package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
	"github.com/hashicorp/vault/api"
)

var name, action string

func init() {
	flag.StringVar(&name, "name", "", "Set a name to controller")
	flag.StringVar(&action, "action", "create", "Action to perform: action or read")
}

func query(url, name string) (string, error) {
	client := &http.Client{}

	// Create the request body with the name argument
	requestBody := bytes.NewBufferString(fmt.Sprintf("name=%s", name))

	// Create the POST request with the request body
	req, err := http.NewRequest("POST", url, requestBody)
	if err != nil {
		return "", err
	}

	// Set the Content-Type header to "application/x-www-form-urlencoded"
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// Handle error
	}

	return string(body), nil
}

func main() {
	flag.Parse()

	if name == "" {
		name = uuid.New().String()
	}

	switch action {
	case "delete":
		fmt.Println("name", name)
		_, err := query("http://localhost:8080/delete", name)
		if err != nil {
			panic(err)
		}

	case "create":
		fmt.Println("name", name)
		_, err := query("http://localhost:8080/create", name)
		if err != nil {
			panic(err)
		}

		tempToken, err := query("http://localhost:8080/login", name)
		if err != nil {
			panic(err)
		}

		fmt.Println("temp token:", tempToken)

		client, err := api.NewClient(api.DefaultConfig())
		if err != nil {
			panic(err)
		}
		client.SetToken(tempToken)

		// Read cubbyhole to retrieve wrapped token
		wrappedToken, err := ReadCubbyhole(client, name)
		if err != nil {
			panic(err)
		}

		// Unwrap wrapped token
		definitiveToken, err := Unwrap(client, wrappedToken)
		if err != nil {
			panic(err)
		}

		fmt.Println("definitive token", definitiveToken)

		// Read secret
		someSecret, err := ReadSecret(client, definitiveToken, name)
		if err != nil {
			panic(err)
		}

		fmt.Println("identity:", someSecret)
	}
}
