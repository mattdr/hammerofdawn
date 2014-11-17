package app

import (
  "fmt"
  "net/http"
)

func root(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Welcome")
}

func init() {
  http.HandleFunc("/", root)
}