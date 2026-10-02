package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	root := flag.String("root", "dist", "directory to serve")
	addr := flag.String("addr", "127.0.0.1:8765", "listen address")
	flag.Parse()
	log.Fatal(http.ListenAndServe(*addr, http.FileServer(http.Dir(*root))))
}
