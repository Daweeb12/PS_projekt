package raft

import (
	"context"
	"errors"
	"testing"
	"time"
)

// InProcRPC routes RPC calls directly between in-process Raft instances.
type InProcRPC struct {
	nodes map[int]*Raft
}

func (r *InProcRPC) RequestVote(ctx context.Context, peer int, args *RequestVoteArgs, reply *RequestVoteReply) error {
	node, ok := r.nodes[peer]
	if !ok {
		return errors.New("peer not found")
	}
	return node.RequestVote(ctx, args, reply)
}

func (r *InProcRPC) AppendEntries(ctx context.Context, peer int, args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	node, ok := r.nodes[peer]
	if !ok {
		return errors.New("peer not found")
	}
	return node.AppendEntries(ctx, args, reply)
}

// TestLeaderElection verifies that among 3 in-process Raft nodes, one becomes leader.
func TestLeaderElection(t *testing.T) {
	peers := []int{1, 2, 3}
	nodes := make(map[int]*Raft)
	rpc := &InProcRPC{nodes: nodes}

	// create apply channels
	applyChs := make(map[int]chan ApplyMsg)
	for _, id := range peers {
		applyChs[id] = make(chan ApplyMsg, 64)
	}

	// start nodes
	for _, id := range peers {
		rf := NewRaft(id, peers, rpc, applyChs[id])
		// shorten election timers for faster test
		rf.mu.Lock()
		rf.electionTimeout = 120 * time.Millisecond
		rf.mu.Unlock()
		nodes[id] = rf
	}

	// wait for a leader to appear
	deadline := time.After(3 * time.Second)
	ticker := time.NewTicker(150 * time.Millisecond)
	defer ticker.Stop()

	var leaderCount int
	var leaderID int
	for {
		select {
		case <-deadline:
			// attempt to stop nodes
			for _, n := range nodes {
				n.Stop()
			}
			t.Fatalf("timed out waiting for a single leader; leaderCount=%d", leaderCount)
		case <-ticker.C:
			leaderCount = 0
			for _, id := range peers {
				_, isLeader := nodes[id].GetState()
				if isLeader {
					leaderCount++
					leaderID = id
				}
			}
			if leaderCount == 1 {
				// success
				for _, n := range nodes {
					n.Stop()
				}
				t.Logf("leader elected: %d", leaderID)
				return
			}
		}
	}
}
