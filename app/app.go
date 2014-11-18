package app

import (
	"appengine"
	"appengine/channel"
	"appengine/datastore"
	"appengine/urlfetch"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/compute/v1"
)

func root(responseWriter http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	accessToken, expiryTime, err := appengine.AccessToken(context, compute.ComputeScope)
	var _ = accessToken
	var _ = expiryTime
	if err != nil {
		http.Error(responseWriter, "Couldn't get access token", 500)
	}

	// https://code.google.com/p/google-api-go-client/wiki/GettingStarted
	transport := &oauth.Transport{
		Token:     &oauth.Token{AccessToken: accessToken},
		Transport: &urlfetch.Transport{Context: context},
	}

	// https://code.google.com/p/google-api-go-client/source/browse/compute/v1/compute-gen.go
	computeApi, err := compute.New(transport.Client())
	if err != nil {
		http.Error(responseWriter, "Couldn't activate Compute API", 500)
		return
	}

	project := "g-hammerofdawn"
	zone := "us-central1-b"
	list, err := computeApi.Instances.List(project, zone).Do()
	if err != nil {
		http.Error(responseWriter, "Couldn't retrieve instances", 500)
		return
	}

	for _, instance := range list.Items {
		fmt.Fprintf(responseWriter, "%#v\n", *instance)
	}

	var _ = computeApi
	fmt.Fprintf(responseWriter, "Complete")
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/brb", brb)
	http.HandleFunc("/brbtrigger", brbTrigger)
	// http.HandleFunc("/_ah/channel/connected", brbConnected)
	// http.HandleFunc("/_ah/channel/disconnected", brbDisconnected)
}

var brbTemplate = template.Must(template.ParseFiles("brb.html"))

type BigRedButton struct {
	Done     bool
	Listener []string
}

type BigRedButtonData struct {
	Done bool
}

// stop implements the Big Red Button
//
// Based on the channel example:
// https://cloud.google.com/appengine/docs/go/channel/
func brb(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	key := r.FormValue("key")
	id := r.FormValue("id")
	if key == "" {
		http.Error(w, "No KEY specified", http.StatusInternalServerError)
		return
	}
	if id == "" {
		http.Error(w, "No ID specified", http.StatusInternalServerError)
		return
	}

	// Persist state in the datastore.
	running := "true"
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		k := datastore.NewKey(c, "BigRedButton", key, 0, nil)
		brb := new(BigRedButton)
		_ = datastore.Get(c, k, brb)
		// Ignore the error.
		if brb.Done {
			running = "false"
		}
		found := false
		for _, lid := range brb.Listener {
			if lid == id {
				found = true
			}
		}
		if found {
			return nil
		}
		// Not found. Store it instead.
		brb.Listener = append(brb.Listener, id)
		_, err := datastore.Put(c, k, brb)
		return err
	}, nil)
	if err != nil {
		http.Error(w, "Couldn't load State", http.StatusInternalServerError)
		c.Errorf("channel.Create: %v", err)
		return
	}

	tok, err := channel.Create(c, id+key)
	if err != nil {
		http.Error(w, "Couldn't create Channel", http.StatusInternalServerError)
		c.Errorf("channel.Create: %v", err)
		return
	}

	err = brbTemplate.Execute(w, map[string]string{
		"token":   tok,
		"id":      id,
		"key":     key,
		"running": running,
	})
	if err != nil {
		c.Errorf("brbTemplate: %v", err)
	}
}

// The trigger method
func brbTrigger(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	key := r.FormValue("key")
	id := r.FormValue("id")
	if key == "" {
		http.Error(w, "No KEY specified", http.StatusInternalServerError)
		return
	}
	if id == "" {
		http.Error(w, "No ID specified", http.StatusInternalServerError)
		return
	}

	brb := new(BigRedButton)
	err := datastore.RunInTransaction(c, func(c appengine.Context) error {
		k := datastore.NewKey(c, "BigRedButton", key, 0, nil)
		if err := datastore.Get(c, k, brb); err != nil {
			return err
		}
		if brb.Done {
			return errors.New("Already done")
		}
		brb.Done = true

		// Update the Datastore.
		_, err := datastore.Put(c, k, brb)
		return err
	}, nil)
	if err != nil {
		http.Error(w, "Couldn't trigger", http.StatusInternalServerError)
		c.Errorf("trigger: %v", err)
		return
	}

	// Send the state to both clients.
	data := BigRedButtonData{Done: brb.Done}
	for _, id := range brb.Listener {
		err := channel.SendJSON(c, id+key, data)
		if err != nil {
			c.Errorf("sending trigger: %v", err)
		}
	}
}
