package main

import (
	"net/http"

	william "github.com/topfreegames/Will.IAM/pkg/client"
)

func main() {
	client := william.New("http://localhost:4040", "DemoService")

	mux := http.NewServeMux()
	mux.HandleFunc("/am", client.AmHandler)
	mux.HandleFunc("/action", client.HandlerFunc(
		client.Generate("RL", "Action"),
		func(w http.ResponseWriter, r *http.Request) {
			// User require a permission of:
			// DemoService::RL::Action::*
		},
	))

	http.ListenAndServe(":8000", mux)
}
