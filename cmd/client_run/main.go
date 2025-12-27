package main

import (
	"PS_projekt/clientlib"
	"flag"
)

// here basically for quicker access to Client
func main() {
	url := flag.String("url", "localhost:9001", "gRPC server address")
	flag.Parse()
	clientlib.Client(*url)
}
