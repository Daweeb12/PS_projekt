package main

import (
	"flag"
	"fmt"
)

func main() {
	idPtr := flag.Int64("id", 0, "id of the node")
	portPrt := flag.Int64("p", 9000, "port where the node will run")
	flag.Parse()
	id := *idPtr
	port := *portPrt
	fmt.Println("port", port)
	fmt.Println("id", id)
	url := fmt.Sprintf("localhost:%d", port)
	masterUrl := fmt.Sprintf("localhost:%d", 9000)
	if id == 0 {
		//add node to chain
		StartMasterServer(masterUrl, 0)
	} else if id < 5 {
		AddMsgBoardServer(url, masterUrl, id)
	} else {
		//start master node server
		UpdateClient(masterUrl)

	}

}
