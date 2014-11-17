package app

import (
  "fmt"
  "net/http"
)

func root(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello")
}

func init() {
  http.HandleFunc("/", root)
}
