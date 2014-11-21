package app

import (
	"appengine"
	// "appengine/channel"
	"appengine/datastore"
	"appengine/urlfetch"
	"fmt"
	// "html/template"
	"io"
	"net/http"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/compute/v1"
)

var (
	PROJECT string = "g-hammerofdawn"
)

func createComputeApi(request *http.Request) (service *compute.Service, err error) {
	context := appengine.NewContext(request)
	accessToken, expiryTime, err := appengine.AccessToken(context, compute.ComputeScope)
	var _ = accessToken
	var _ = expiryTime
	if err != nil {
		return nil, err
	}

	// https://code.google.com/p/google-api-go-client/wiki/GettingStarted
	transport := &oauth.Transport{
		Token:     &oauth.Token{AccessToken: accessToken},
		Transport: &urlfetch.Transport{Context: context},
	}

	// https://code.google.com/p/google-api-go-client/source/browse/compute/v1/compute-gen.go
	computeApi, err := compute.New(transport.Client())
	if err != nil {
		return nil, err
	}

	return computeApi, nil
}

func root(responseWriter http.ResponseWriter, request *http.Request) {
	computeApi, err := createComputeApi(request)
	if err != nil {
		http.Error(responseWriter, "Couldn't use Compute API", 500)
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

func startonevm(responseWriter http.ResponseWriter, request *http.Request) {
	projectPrefix := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%v", PROJECT)
	zone := "us-central1-b"
	pdStandardDiskType := fmt.Sprintf("%v/zones/%v/diskTypes/pd-standard", projectPrefix, zone)
	n1Standard1MachineType := fmt.Sprintf("%v/zones/%v/machineTypes/n1-standard-1", projectPrefix, zone)
	defaultNetwork := fmt.Sprintf("%v/global/networks/default", projectPrefix)

	instance := &compute.Instance{
		Description: "test9df",
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				// DeviceName (use default)
				InitializeParams: &compute.AttachedDiskInitializeParams{
					// DiskName (use default)
					DiskSizeGb: 30,
					DiskType:   pdStandardDiskType,
					// Unlike gcloud, the API doesn't do alias matching. :(
					SourceImage: "https://www.googleapis.com/compute/v1/projects/debian-cloud/global/images/backports-debian-7-wheezy-v20141108",
				},
				Type: "PERSISTENT",
			},
		},
		MachineType: n1Standard1MachineType,
		// Metadata
		Name: "test7c8",
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Name: "test42f",
						Type: "ONE_TO_ONE_NAT",
					},
				},
				Network: defaultNetwork,
			},
		},
		// ServiceAccounts
		// Tags
	}

	computeApi, err := createComputeApi(request)
	if err != nil {
		http.Error(responseWriter, "Couldn't use Compute API", 500)
		return
	}

	op, err := computeApi.Instances.Insert(PROJECT, zone, instance).Do()
	if err != nil {
		http.Error(responseWriter, fmt.Sprintf("Couldn't insert instance: %v", err), 500)
		return
	}
	fmt.Fprint(responseWriter, op)
}

/*
BATCH_NAME= 10K

CLIENTS_NUM_MAX=10000
CLIENTS_NUM_START=100
CLIENTS_RAMPUP_INC=50

INTERFACE   =eth0
NETMASK=16
IP_ADDR_MIN= 192.168.1.1
IP_ADDR_MAX= 192.168.53.255

*/
// Config describes the basic configuration of a loadtest.
// TODO: Figure out the ip config, as above.
//
// Generally use ClientsMax to determine the load,
type Config struct {
	Name             string
	Cycles           int
	ClientsMax       int
	ClientsNumStart  int
	ClientsRampupInc int
	Line             []string
	URL              []CurlURL
}

func (c *Config) write(w io.Writer) {
	fmt.Fprintf(w, "BATCH_NAME=%s\n", c.Name)

	// Describe the client behavior
	fmt.Fprintf(w, "CLIENTS_NUM_MAX=%d", c.ClientsMax)
	fmt.Fprintf(w, "CLIENTS_NUM_START=%d", c.ClientsNumStart)
	fmt.Fprintf(w, "CLIENTS_RAMPUP_INC=%d", c.ClientsRampupInc)

	// Describe the network config space.

	fmt.Fprintf(w, strings.Join(c.Line, "\n"))
	fmt.Fprintf(w, "CYCLES_NUM=%d", c.Cycles)
	fmt.Fprintf(w, "URLS_NUM=%d", len(c.URL))
	for i := range c.URL {
		c.URL[i].write(w)
	}
}

type CurlURL struct {
	URL                  string // http://www.google.com
	URLShortName         string // "google-com"
	RequestType          string // GET
	TimerURLCompletionMs int    // enforced by cancelling url fetch on timeout
	TimerAfterURLSleep   int    //
	RandomMin            int
	RandomMax            int
	RandomToken          string
	Header               []string
}

func (c *CurlURL) write(w io.Writer) {
	fmt.Fprintln(w, "")
	fmt.Fprintf(w, "URL=%s\n", c.URL)
	fmt.Fprintf(w, "URL_SHORT_NAME=%s\n", c.URLShortName)
	fmt.Fprintf(w, "REQUEST_TYPE=%s\n", c.RequestType)
	if len(c.Header) > 0 {
		for _, v := range c.Header {
			fmt.Fprintf(w, "HEADER=%q", v)
		}
	}
	if c.RandomMin != 0 || c.RandomMax != 0 {
		fmt.Fprintf(w, "URL_RANDOM_RANGE=%d-%d\n", c.RandomMin, c.RandomMax)
		fmt.Fprintf(w, "URL_RANDOM_TOKEN=%s\n", c.RandomToken)
	}
	fmt.Fprintf(w, "TIMER_URL_COMPLETION=%d\n", c.TimerURLCompletionMs)
	if c.TimerAfterURLSleep > 0 {
		fmt.Fprintf(w, "TIMER_AFTER_URL_SLEEP=%s\n", c.TimerAfterURLSleep)
	}
}

func config(w http.ResponseWriter, request *http.Request) {
	context := appengine.NewContext(request)
	key := request.FormValue("key")
	if key == "" {
		http.Error(w, "No KEY specified", http.StatusInternalServerError)
		return
	}
	k := datastore.NewKey(context, "LoadConfig", key, 0, nil)
	config := new(Config)
	err := datastore.Get(context, k, config)
	if err != nil {
		http.Error(w, "Read failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
}

func init() {
	http.HandleFunc("/", root)
	http.HandleFunc("/startonevm", startonevm)
	http.HandleFunc("config", config)
}
