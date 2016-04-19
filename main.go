package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/distribution/notifications"
	"github.com/duyanghao/registry-notification-server/config"
	"github.com/duyanghao/registry-notification-server/pkg/handler"
	"net/http"
	"strings"
)

var con_fig *config.Config

//map handler
var DIS_RULE = map[string]func(http.ResponseWriter, *http.Request, notifications.Event, *config.Config) error{
	"pull": handler.ProcessPullEvent,
	"push": handler.ProcessPushEvent,
}

type Dispatcher struct {
	disRule map[string]func(http.ResponseWriter, *http.Request, notifications.Event, *config.Config) error
}
type Server struct {
	dispatcher Dispatcher
}

func newServer() Server {
	dis := Dispatcher{DIS_RULE}
	server := Server{dis}
	return server
}

func (server Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Bypass
	uri := r.RequestURI
	uri_string := strings.Split(uri, "/")
	if uri == "/favicon.ico" {
		http.NotFound(w, r)
		return
	}
	if len(uri_string) > 2 {
		if uri_string[1] == "search" { //for search handler
			handler.ProcessSearch(w, r, con_fig)
			return
		} else if uri_string[1] == "analysis" { //for analysis handler
			handler.ProcessAnalysis(w, r, con_fig)
			return
		} else { //ignore unexpected http request
			http.NotFound(w, r)
			return
		}
	}

	// A request needs to be made via POST
	if r.Method != "POST" {
		fmt.Printf("ERROR: Ignoring request. Required method is \"POST\" but got \"%s\".\n", r.Method)
		http.Error(w, fmt.Sprintf("Ignoring request. Required method is \"POST\" but got \"%s\".\n", r.Method), http.StatusOK)
		return
	}
	// A request must have a body.
	if r.Body == nil {
		fmt.Printf("ERROR: Ignoring request. Required non-empty request body.\n")
		http.Error(w, "Ignoring request. Required non-empty request body.\n", http.StatusOK)
		return
	}

	// Test for correct mimetype and reject all others
	// Even the documentation on docker notfications says that we shouldn't be to
	// picky about the mimetype. But we are and let the caller know this.
	contentType := r.Header.Get("Content-Type")
	if contentType != notifications.EventsMediaType {
		fmt.Printf("ERROR: Ignoring request. Required mimetype is \"%s\" but got \"%s\"\n", notifications.EventsMediaType, contentType)
		http.Error(w, fmt.Sprintf("Ignoring request. Required mimetype is \"%s\" but got \"%s\"\n", notifications.EventsMediaType, contentType), http.StatusOK)
		return
	}

	all_event := notifications.Envelope{}
	json_decoder := json.NewDecoder(r.Body)
	err := json_decoder.Decode(&all_event)
	if err != nil {
		fmt.Printf("ERROR:Failed to decode Envelope: %s\n", err)
		http.Error(w, fmt.Sprintf("Failed to decode envelope: %s\n", err), http.StatusBadRequest)
		return

	}
	// process events
	for _, event := range all_event.Events {
		switch event.Action {
		case notifications.EventActionPull:
			err := server.dispatcher.disRule["pull"](w, r, event, con_fig)
			if err != nil {
				fmt.Printf("UNEXPECT Error: %s\n", err)
				http.Error(w, fmt.Sprintf("UNEXPECT ERR: %s\n", err), http.StatusBadRequest)
				return
			}
		case notifications.EventActionPush:
			err := server.dispatcher.disRule["push"](w, r, event, con_fig)
			if err != nil {
				fmt.Printf("UNEXPECT Error: %s\n", err)
				http.Error(w, fmt.Sprintf("UNEXPECT ERR: %s\n", err), http.StatusBadRequest)
				return
			}
		case notifications.EventActionDelete:
			fmt.Printf("ERROR: Manifest event type is: %s\n", event.Action)
			http.Error(w, fmt.Sprintf("Manifest event type is: %s\n", event.Action), http.StatusOK)
			return
		default:
			fmt.Printf("ERROR: Manifest event type is: %s\n", event.Action)
			http.Error(w, fmt.Sprintf("Manifest event type is: %s\n", event.Action), http.StatusOK)
			return
		}
	}
	http.Error(w, fmt.Sprintf("Done\n"), http.StatusOK)
}

func main() {
	flag.Parse()
	// Load config file given by first argument
	configFilePath := flag.Arg(0)
	if configFilePath == "" {
		fmt.Println("Error: Config file not specified")
		return
	}

	fmt.Println("INFO: Starting Loading config!")
	_, err := config.LoadConfig(configFilePath)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	con_fig, _ = config.LoadConfig(configFilePath)
	fmt.Println("Loading config done!")

	// Setup HTTP endpoint
	var httpConnectionString = con_fig.GetEndpointConnectionString()
	fmt.Printf("INFO: Listening on %s\n", httpConnectionString)
	//start the server
	var s Server
	s = newServer()
	// http server
	err = http.ListenAndServe(httpConnectionString, s)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Printf("Error: Exiting\n")
}
