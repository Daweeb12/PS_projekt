package master_node

import (
	"fmt"
	"sync"
	"testing"
)

func TestAddNode(t *testing.T) {
	masterNode := NewMasterNode(0, "localhost:9000")
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func(masterNode *MasterNode) {
		defer wg.Done()
		for i := range 100 {
			node := NewNode(nil, nil, int64(i), fmt.Sprintf("addr%d", i))
			if err := masterNode.AddNode(node); err != nil {
				t.Error(err)
			}
		}
	}(masterNode)

	go func(masterNode *MasterNode) {
		defer wg.Done()
		for i := range 100 {
			node := NewNode(nil, nil, int64(i), fmt.Sprintf("addr%d", i))
			if err := masterNode.AddNode(node); err != nil {
				t.Error(err)
			}
		}
	}(masterNode)
	wg.Wait()

	fmt.Println("CHAIN LENGTH", masterNode.ChainLen.Load())

}
