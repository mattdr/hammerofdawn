package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
)

const (
	URL    = "https://g-hammerofdawn.appspot.com/config?key="
	Binary = "cat"
	// "/usr/bin/curl-loader"
)

var key = flag.String("key", "", "Config key")

func main() {
	flag.Parse()
	if *key == "" {
		panic("Key required.")
	}

	url := fmt.Sprintf("%s%s", URL, *key)
	req, err := http.NewRequest("GET", url, nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	f, err := ioutil.TempFile("", *key)
	if err != nil {
		panic(err)
	}
	f.Write(body)
	fmt.Printf("file: %s\n", f.Name())
	defer f.Close()

	cmd := exec.Command(Binary, f.Name())
	err = cmd.Run()
	if err != nil {
		panic(err)
	}
}
