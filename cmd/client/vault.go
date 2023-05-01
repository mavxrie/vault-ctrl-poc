package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/vault/api"
)

func ReadCubbyhole(client *api.Client, name string) (string, error) {
	path := fmt.Sprintf("cubbyhole/%s/token", name)
	secret, err := client.Logical().Read(path)
	if err != nil {
		return "", err
	}

	return secret.Data["token"].(string), nil
}

func Unwrap(client *api.Client, wrappedToken string) (string, error) {
	// Send the request and get the response
	path := "/sys/wrapping/unwrap"

	client.SetToken(wrappedToken)
	resp, err := client.Logical().Write(path, map[string]interface{}{})
	if err != nil {
		return "", err
	}

	return resp.Data["data"].(string), nil
}

func ReadSecret(client *api.Client, token, name string) (string, error) {
	path := fmt.Sprintf("controllers/%s/identity", name)
	tempClient, err := client.Clone()
	if err != nil {
		return "", err
	}
	tempClient.SetToken(token)

	secret, err := tempClient.KVv2("secret-controllers").Get(context.TODO(), path)
	if err != nil {
		return "", err
	}

	return secret.Data["identity"].(string), nil
}
