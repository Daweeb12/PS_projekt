package internal

import (
	"fmt"
	"testing"
)

func TestAddNode(t *testing.T) {
	fmt.Println("started add test")
	masterNode := NewMasterNode(0, "a")
	bNode := &Node{Id: 1, Url: "b"}
	masterNode.AddNode(bNode)
	cNode := &Node{Id: 2, Url: "c"}
	masterNode.AddNode(cNode)
	dNode := &Node{Id: 3, Url: "d"}
	dNode.Url = "d"
	masterNode.AddNode(dNode)
	for i := masterNode.Head; i != nil; i = i.Next {
		fmt.Println("node.Url: ", i.Url)
	}
	fmt.Println("from tail")
	for i := masterNode.Tail; i != nil; i = i.Prev {
		fmt.Println("node.Url: ", i.Url)
	}
	fmt.Println("ended add test")
}

func TestRemoveNode(t *testing.T) {
	fmt.Println("started delete test")
	masterNode := NewMasterNode(0, "a")
	bNode := &Node{Id: 1, Url: "b"}
	masterNode.AddNode(bNode)
	cNode := &Node{Id: 2, Url: "c"}
	masterNode.AddNode(cNode)
	dNode := &Node{Id: 3, Url: "d"}
	dNode.Url = "d"
	masterNode.AddNode(dNode)
	masterNode.RemoveNode(dNode)
	fmt.Println()
	for node := masterNode.Head; node != nil; node = node.Next {
		fmt.Println("node.Url ", node.Url)
	}
	fmt.Printf("from tail \n\n")
	for i := masterNode.Tail; i != nil; i = i.Prev {
		fmt.Println("node.Url: ", i.Url)
	}
	fmt.Println("ended delete test")

}

func TestRemoveTail(t *testing.T) {
	fmt.Println("start remove tail test")
	masterNode := NewMasterNode(0, "a")
	bNode := &Node{Id: 1, Url: "b"}
	masterNode.AddNode(bNode)
	cNode := &Node{Id: 2, Url: "c"}
	masterNode.AddNode(cNode)
	dNode := &Node{Id: 3, Url: "d"}
	dNode.Url = "d"
	masterNode.AddNode(dNode)
	fmt.Println()
	masterNode.RemoveNode(bNode)
	for node := masterNode.Head; node != nil; node = node.Next {
		fmt.Println("node.Url ", node.Url)
	}
	fmt.Printf("from tail \n\n")
	for i := masterNode.Tail; i != nil; i = i.Prev {
		fmt.Println("node.Url: ", i.Url)
	}

	fmt.Println("ended remove tail test")

}
