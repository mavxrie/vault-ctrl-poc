package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/vault/api"
)

// CreateController creates a controller approle in Vault and returns the role_id
func CreateController(client *api.Client, name string) (string, error) {
	config := map[string]interface{}{
		"bind_secret_id":    "false",
		"token_ttl":         "24h",
		"token_max_ttl":     "768h",
		"token_bound_cidrs": "0.0.0.0/0",
		"token_policies":    "ctrl-policy",
	}

	log.Printf("Creating ctrl with name %s", name)

	_, err := client.Logical().Write(
		fmt.Sprintf("auth/approle-controllers/role/%s", name),
		config,
	)
	if err != nil {
		return "", err
	}

	// Read role-id
	role, err := client.Logical().Read(
		fmt.Sprintf("auth/approle-controllers/role/%s/role-id", name),
	)
	if err != nil {
		return "", err
	}

	role_id := role.Data["role_id"].(string)

	path := fmt.Sprintf("controllers/%s", name)
	data := map[string]interface{}{
		"role_id": role_id,
	}

	// Save name <-> role_id into secrets.
	_, err = client.KVv2("secret").Put(context.TODO(), path, data)
	if err != nil {
		return "", err
	}

	// Save somedata in controllers's space
	path = fmt.Sprintf("controllers/%s/identity", name)
	data = map[string]interface{}{
		"identity": name,
	}
	_, err = client.KVv2("secret-controllers").Put(context.TODO(), path, data)
	if err != nil {
		return "", err
	}

	return role_id, nil
}

// CheckControllerUniqness checks given name exists
// TODO: This needs to be done.
func CheckControllerUniqness(client *api.Client, name string) bool {
	return false
}

// Login retrieves controller role_id per name, then retrieves a token.
//
// > vault write auth/approle-controllers/login role_id=2aadc38e-59e8-f1ba-2302-565fe4480cb9
func LoginController(client *api.Client, name string) (string, error) {
	// Retrieve role_id for controller
	path := fmt.Sprintf("controllers/%s", name)
	secret, err := client.KVv2("secret").Get(context.TODO(), path)
	if err != nil {
		return "", err
	}

	role_id := secret.Data["role_id"].(string)

	// Login
	loginData, err := client.Logical().Write(
		"auth/approle-controllers/login",
		map[string]interface{}{
			"role_id": role_id,
		},
	)
	if err != nil {
		return "", err
	}

	token, err := loginData.TokenID()
	if err != nil {
		return "", err
	}

	return token, nil
}

func Wrap(client *api.Client, plaintext string) (string, error) {
	wrapData := map[string]interface{}{
		"data": plaintext,
	}

	wrappedSecret, err := client.Logical().Write("sys/wrapping/wrap", wrapData)
	if err != nil {
		return "", err
	}

	return wrappedSecret.WrapInfo.Token, nil
}

// CreateTempTokenCubbyhole creates a temporary token to write/read cubbyhole and stores
// the "data" in it.
func CreateTempTokenCubbyhole(client *api.Client, name, data string) (string, error) {
	tempClient, err := client.Clone()
	if err != nil {
		panic(err)
	}

	renewable := false
	params := &api.TokenCreateRequest{
		Policies:  []string{"cubbyhole-limited"},
		TTL:       "5m",
		Renewable: &renewable,
	}

	// Create the token
	token, err := client.Auth().Token().Create(params)
	if err != nil {
		return "", err
	}

	// Store data in cubbyhole
	path := fmt.Sprintf("cubbyhole/%s/token", name)
	tempClient.SetToken(token.Auth.ClientToken)
	_, err = tempClient.Logical().Write(path, map[string]interface{}{
		"token": data,
	})
	if err != nil {
		return "", err
	}

	return token.Auth.ClientToken, nil
}

func DeleteSecrets(client *api.Client, name string) error {
	path := fmt.Sprintf("controllers/%s", name)
	err := client.KVv2("secret").Delete(context.TODO(), path)
	if err != nil {
		return err
	}
	log.Printf("secret %s deleted", path)

	path = fmt.Sprintf("controllers/%s", name)
	err = client.KVv2("secret-controllers").Delete(context.TODO(), path)
	if err != nil {
		return err
	}
	log.Printf("secret %s deleted", path)

	return nil
}

func DeleteAppRole(client *api.Client, name string) error {
	path := fmt.Sprintf("auth/approle-controllers/role/%s", name)
	_, err := client.Logical().Delete(path)
	if err != nil {
		return err
	}
	log.Printf("approle %s deleted", path)

	return nil
}
