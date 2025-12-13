package main

import (
	"flag"
)

func main() {
	p := flag.Int("p", 9000, "port where the server runs")
	port := *p
	if port != 9000 {
		//start client
	} else {
		//start server

	}

}
