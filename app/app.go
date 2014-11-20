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
	http.HandleFunc("config", config)
}

type Config struct {
	Line []string
}

func config(w http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	key := r.FormValue("key")
	if key == "" {
		http.Error(w, "No KEY specified", http.StatusInternalServerError)
		return
	}
	k := datastore.NewKey(c, "LoadConfig", key, 0, nil)
	config := new(Config)
	err := datastore.Get(c, k, config)
	if err != nil {
		http.Error(w, "Read failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, strings.Join(Config.Line, "\n"))
}
