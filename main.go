package main

import (
	"fmt"
	"github.com/cbrgm/githubevents/githubevents"
	"github.com/tjololo/mope/pkg/handlers"
	"github.com/tjololo/mope/pkg/utils"
	"net/http"
	"os"
)

func main() {
	utils.InitializeLogger()
	defer utils.Logger.Sync()
	secret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	handle := githubevents.New(secret)

	handle.OnIssueCommentCreated(
		handlers.HandleIssueCommentCreated,
	)
	// add a http handleFunc
	http.HandleFunc("/hook", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Add("Content-Type", "application/json")
		err := handle.HandleEventRequest(r)
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	})
	// start the server listening on port 8080
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}
