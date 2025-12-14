package main

import (
	"PS_projekt/api/grpc/protobufInternal"
	"PS_projekt/internal"
	"flag"
	"fmt"
)

func main() {
	p := flag.Int("p", 9000, "port where the server runs")
	idPtr := flag.Int("id", 0, "id of the chain node")
	port := *p
	id := *idPtr

	if id != 0 {
		//start control server
		if err := ChainNodeServer(url, id); err != nil {
			panic(err)
		}
	} else {
		//start node server

	}

}

func InitChainNodes(nodesNum, id, port int) {
	chainNodes := []*internal.ChainNode{}
	for i := range nodesNum {
		portNumber := port + i
		url := fmt.Sprintf("http://localhost:%d", portNumber)
		chainNode := internal.NewChainNode(int64(i), url)
		chainNodes = append(chainNodes, chainNode)
	}
}
