package solver

import (
	"fmt"
	"net/http"
)

var matchInfo string

//RunWeb run a webserver
func RunWeb(port string) {
	// The "HandleFunc" method accepts a path and a function as arguments
	// (Yes, we can pass functions as arguments, and even trat them like variables in Go)
	// However, the handler function has to have the appropriate signature (as described by the "handler" function below)
	http.HandleFunc("/match", handler)

	// After defining our server, we finally "listen and serve" on port 8080
	// The second argument is the handler, which we will come to later on, but for now it is left as nil,
	// and the handler defined above (in "HandleFunc") is used
	http.ListenAndServe(":"+port, nil)
}

// "handler" is our handler function. It has to follow the function signature of a ResponseWriter and Request type
// as the arguments.
func handler(w http.ResponseWriter, r *http.Request) {
	// For this case, we will always pipe "Hello World" into the response writer

	fmt.Fprintf(w, "Hello solver!\n")
	fmt.Fprintf(w, getMatch())
}

func getMatch() string {
	return matchInfo
}

func setMatch(json string) {
	matchInfo = json
}
