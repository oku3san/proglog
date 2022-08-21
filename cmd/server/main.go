package main

import (
  "github.com/oku3san/proglog/internal/server"
  "log"
)

func main() {
  srv := server.NewHTTPServer(":8080")
  log.Fatal(srv.ListenAndServe())
}
