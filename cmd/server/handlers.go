package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/vault/api"
)

func createHandler(client *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
		}

		name := r.FormValue("name")
		if name == "" {
			log.Println("Missing or invalid 'name' field.")
			http.Error(w, "Missing or invalid 'name' field.", http.StatusInternalServerError)
			return
		}

		_, err = CreateController(client, name)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func deleteHandler(client *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
		}

		name := r.FormValue("name")
		if name == "" {
			log.Println("Missing or invalid 'name' field.")
			http.Error(w, "Missing or invalid 'name' field.", http.StatusInternalServerError)
			return
		}

		// Need to delete controller secrets
		// Need to delete metadata secret
		err = DeleteSecrets(client, name)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Need to delete approle
		err = DeleteAppRole(client, name)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func loginHandler(client *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
		}

		name := r.FormValue("name")
		if name == "" {
			log.Println("Missing or invalid 'name' field.")
			http.Error(w, "Missing or invalid 'name' field.", http.StatusInternalServerError)
			return
		}

		log.Printf("Login with name %s", name)

		token, err := LoginController(client, name)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		wrappedToken, err := Wrap(client, token)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cubbyToken, err := CreateTempTokenCubbyhole(client, name, wrappedToken)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, cubbyToken)
	}
}

func secretHandler(client *api.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TOOD: This part must be reworked
		secret, err := client.Logical().Read("secret/myapp")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if secret == nil {
			http.Error(w, "Secret not found", http.StatusNotFound)
			return
		}

		// write the secret data to the response
		fmt.Fprintf(w, "Secret data: %v", secret.Data)
	}
}
