package app

import (
	"appengine"
	"appengine/urlfetch"
	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/compute/v1"
	"fmt"
	"net/http"
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
	}

	project := "g-hammerofdawn"
	res, err := computeApi.Images.List(project).Do()
	fmt.Fprintf(responseWriter, "%v %v\n", res, err)

	var _ = computeApi
	fmt.Fprintf(responseWriter, "Complete")
}

func init() {
	http.HandleFunc("/", root)
}
