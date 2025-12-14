package main

import (
	"flag"
	"fmt"
)

func main() {

	sPtr := flag.String("s", "", "server URL")
	pPtr := flag.Int("p", 9876, "port number")
	flag.Parse()

	// Start server/client
	url := fmt.Sprintf("%v:%v", *sPtr, *pPtr)
	if *sPtr == "" {
		Server(url)
	} else {
		Client(url)
	}
}
