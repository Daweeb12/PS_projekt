package raft

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"
)

// State represents the role of a Raft node
type State int

const (
	Follower State = iota
	Candidate
	Leader
)

// LogEntry is a single log entry stored by Raft
type LogEntry struct {
	Term    int64
	Command []byte
}

// ApplyMsg is sent on apply channel when a log entry is committed
type ApplyMsg struct {
	Index   int64
	Command []byte
}

// RequestVoteArgs is the arguments for RequestVote RPC
type RequestVoteArgs struct {
	Term         int64
	CandidateId  int
	LastLogIndex int64
	LastLogTerm  int64
}

// RequestVoteReply is the reply for RequestVote RPC
type RequestVoteReply struct {
	Term        int64
	VoteGranted bool
}

// AppendEntriesArgs is the arguments for AppendEntries (heartbeat / replication)
type AppendEntriesArgs struct {
	Term         int64
	LeaderId     int
	PrevLogIndex int64
	PrevLogTerm  int64
	Entries      []LogEntry
	LeaderCommit int64
}

// AppendEntriesReply is the reply for AppendEntries RPC
type AppendEntriesReply struct {
	Term    int64
	Success bool
}

// RPCSender defines how the Raft core sends RPCs to peers. Networking layer should
// implement this interface and be provided to Raft when constructing it.
type RPCSender interface {
	RequestVote(ctx context.Context, peer int, args *RequestVoteArgs, reply *RequestVoteReply) error
	AppendEntries(ctx context.Context, peer int, args *AppendEntriesArgs, reply *AppendEntriesReply) error
}

// Raft is the core Raft instance managing state and log
type Raft struct {
	mu sync.Mutex

	id    int
	peers []int // peer ids

	// persistent state
	currentTerm int64
	votedFor    int
	log         []LogEntry

	// volatile state
	commitIndex int64
	lastApplied int64

	// leader state
	nextIndex  map[int]int64
	matchIndex map[int]int64

	state State

	applyCh chan ApplyMsg

	rpc RPCSender

	electionTimeout  time.Duration
	heartbeatTimeout time.Duration

	heartbeatCh chan struct{}
	stopCh      chan struct{}
}

// NewRaft constructs a new Raft instance. The caller must provide an implementation
// of RPCSender to enable actual networked RPCs.
func NewRaft(id int, peers []int, rpc RPCSender, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		id:               id,
		peers:            peers,
		currentTerm:      0,
		votedFor:         -1,
		log:              make([]LogEntry, 0, 128),
		commitIndex:      0,
		lastApplied:      0,
		nextIndex:        make(map[int]int64),
		matchIndex:       make(map[int]int64),
		state:            Follower,
		applyCh:          applyCh,
		rpc:              rpc,
		electionTimeout:  randomElectionTimeout(),
		heartbeatTimeout: 150 * time.Millisecond,
		heartbeatCh:      make(chan struct{}, 1),
		stopCh:           make(chan struct{}),
	}

	// initialize leader bookkeeping
	for _, p := range peers {
		rf.nextIndex[p] = 1
		rf.matchIndex[p] = 0
	}

	go rf.run()
	return rf
}

func randomElectionTimeout() time.Duration {
	// randomized between 300-500ms
	return time.Duration(300+rand.Intn(200)) * time.Millisecond
}

// Stop stops Raft's background goroutines.
func (rf *Raft) Stop() {
	close(rf.stopCh)
}

// GetState returns current term and whether this node believes it's leader.
func (rf *Raft) GetState() (int64, bool) {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.currentTerm, rf.state == Leader
}

// Start appends a new command to the log (if leader) and begins replication.
// It returns the log index, current term and whether this node is leader.
func (rf *Raft) Start(command []byte) (int64, int64, bool) {
	rf.mu.Lock()

	defer rf.mu.Unlock()
	if rf.state != Leader {
		return -1, rf.currentTerm, false
	}

	entry := LogEntry{Term: rf.currentTerm, Command: command}
	rf.log = append(rf.log, entry)
	index := int64(len(rf.log))

	// trigger replication to followers
	term := rf.currentTerm
	peers := append([]int{}, rf.peers...)
	nextIndex := make(map[int]int64)
	for k, v := range rf.nextIndex {
		nextIndex[k] = v
	}
	leaderCommit := rf.commitIndex
	// replicate asynchronously
	for _, p := range peers {
		if p == rf.id {
			continue
		}
		go func(peer int) {
			for {
				rf.mu.Lock()
				if rf.state != Leader {
					rf.mu.Unlock()
					return
				}

				prevIndex := nextIndex[peer] - 1
				var prevTerm int64
				if prevIndex > 0 && prevIndex <= int64(len(rf.log)) {
					prevTerm = rf.log[prevIndex-1].Term
				}

				entries := make([]LogEntry, 0)
				if nextIndex[peer] <= int64(len(rf.log)) {
					entries = append(entries, rf.log[nextIndex[peer]-1:]...)
				}
				args := &AppendEntriesArgs{Term: term, LeaderId: rf.id, PrevLogIndex: prevIndex, PrevLogTerm: prevTerm, Entries: entries, LeaderCommit: leaderCommit}
				rf.mu.Unlock()

				var reply AppendEntriesReply
				ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
				err := rf.rpc.AppendEntries(ctx, peer, args, &reply)
				cancel()
				if err != nil {
					// retry shortly after
					time.Sleep(50 * time.Millisecond)
					continue
				}

				rf.mu.Lock()
				if reply.Term > rf.currentTerm { // basically restart election
					rf.currentTerm = reply.Term
					rf.state = Follower
					rf.votedFor = -1
					rf.mu.Unlock()
					return
				}
				if reply.Success {
					// update nextIndex/matchIndex
					sentCount := int64(0) // ugly workaround but whatever
					if entries != nil {
						sentCount = int64(len(entries))
					}
					if sentCount > 0 {
						nextIndex[peer] = args.PrevLogIndex + 1 + sentCount
						rf.nextIndex[peer] = nextIndex[peer]
						rf.matchIndex[peer] = rf.nextIndex[peer] - 1
					}

					logCount := int64(len(rf.log))
					for n := rf.commitIndex + 1; n <= logCount; n++ {
						if rf.log[n-1].Term != rf.currentTerm {
							continue
						}

						count := 1 // count replicas with matchIndex >= n
						for _, peer2 := range rf.peers {
							if peer2 == rf.id {
								continue
							}
							if rf.matchIndex[peer2] >= n {
								count++
							}
						}
						if count >= (len(rf.peers)/2 + 1) {
							rf.commitIndex = n
						}
					}

					for rf.lastApplied < rf.commitIndex {
						rf.lastApplied++
						cmd := rf.log[rf.lastApplied-1].Command
						rf.mu.Unlock()
						rf.applyCh <- ApplyMsg{Index: rf.lastApplied, Command: cmd}
						rf.mu.Lock()
					}
					rf.mu.Unlock()

					return
				}

				// failed because of log mismatch. decrement nextIndex and retry
				if nextIndex[peer] > 1 {
					nextIndex[peer]--
					rf.nextIndex[peer] = nextIndex[peer]
				} else {
					rf.mu.Unlock()
					time.Sleep(50 * time.Millisecond)
				}
				rf.mu.Unlock()
			}
		}(p)
	}

	return index, rf.currentTerm, true
}

// run is the main background loop handling elections and heartbeat sending.
func (rf *Raft) run() {
	// election timer
	electionTimer := time.NewTimer(rf.electionTimeout)
	defer electionTimer.Stop()

	for {
		select {
		case <-rf.stopCh:
			return
		case <-rf.heartbeatCh: // reset election timer on heartbeat or vote
			if !electionTimer.Stop() {
				select {
				case <-electionTimer.C:
				default:
				}
			}
			electionTimer.Reset(rf.electionTimeout)
		case <-electionTimer.C: // reset timer for next election
			rf.startElection()
			rf.mu.Lock()
			rf.electionTimeout = randomElectionTimeout()
			rf.mu.Unlock()
			electionTimer.Reset(rf.electionTimeout)
		}
	}
}

func (rf *Raft) resetElectionTimer(t *time.Timer) {
	rf.mu.Lock()
	rf.electionTimeout = randomElectionTimeout()
	rf.mu.Unlock()

	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	t.Reset(rf.electionTimeout)
}

func (rf *Raft) startElection() {
	rf.mu.Lock()
	rf.state = Candidate
	rf.currentTerm++
	rf.votedFor = rf.id
	term := rf.currentTerm
	lastLogIndex := int64(len(rf.log))
	var lastLogTerm int64
	if lastLogIndex > 0 {
		lastLogTerm = rf.log[lastLogIndex-1].Term
	}
	rf.mu.Unlock()

	votesC := make(chan bool, len(rf.peers))
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	for _, peer := range rf.peers {
		if peer == rf.id {
			continue
		}

		go func(p int) {
			args := &RequestVoteArgs{Term: term, CandidateId: rf.id, LastLogIndex: lastLogIndex, LastLogTerm: lastLogTerm}
			var reply RequestVoteReply
			err := rf.rpc.RequestVote(ctx, p, args, &reply)
			if err != nil {
				votesC <- false
				return
			}
			rf.mu.Lock()
			if reply.Term > rf.currentTerm {
				rf.currentTerm = reply.Term
				rf.state = Follower
				rf.votedFor = -1
				// signal reset election timer on discovering higher term
				select {
				case rf.heartbeatCh <- struct{}{}:
				default:
				}
			}
			rf.mu.Unlock()
			// signal reset election timer on receiving reply
			select {
			case rf.heartbeatCh <- struct{}{}:
			default:
			}
			votesC <- reply.VoteGranted
		}(peer)
	}

	// count votes (including self)
	granted := 1
	needed := (len(rf.peers)/2 + 1)
	timeout := time.After(250 * time.Millisecond)

	for i := 0; i < len(rf.peers)-1; i++ {
		select {
		case v := <-votesC:
			if v {
				granted++
			}
			if granted >= needed {
				rf.becomeLeader()
				return
			}
		case <-timeout:
			return
		}
	}
}

func (rf *Raft) becomeLeader() {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	rf.state = Leader

	// initialize leader state
	next := int64(len(rf.log) + 1)
	for _, i := range rf.peers {
		rf.nextIndex[i] = next
		rf.matchIndex[i] = 0
	}

	// create heartbeat (empty entries)
	go func() {
		ticker := time.NewTicker(rf.heartbeatTimeout)
		defer ticker.Stop()
		for {
			select {
			case <-rf.stopCh:
				return
			case <-ticker.C:
				rf.sendHeartbeats()
			}
		}
	}()
}

func (rf *Raft) sendHeartbeats() {
	rf.mu.Lock()
	if rf.state != Leader {
		rf.mu.Unlock()
		return
	}

	term := rf.currentTerm
	peers := append([]int{}, rf.peers...)
	nextIndex := make(map[int]int64)
	for k, v := range rf.nextIndex {
		nextIndex[k] = v
	}
	leaderCommit := rf.commitIndex
	rf.mu.Unlock()

	for _, p := range peers {
		if p == rf.id {
			continue
		}

		go func(peer int) { // send minimal or empty entries
			args := &AppendEntriesArgs{Term: term, LeaderId: rf.id, PrevLogIndex: nextIndex[peer] - 1, PrevLogTerm: 0, Entries: nil, LeaderCommit: leaderCommit}
			var reply AppendEntriesReply
			ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
			defer cancel()
			_ = rf.rpc.AppendEntries(ctx, peer, args, &reply)
		}(p)
	}
}

// Handle RequestVote RPC call from a candidate.
func (rf *Raft) RequestVote(ctx context.Context, args *RequestVoteArgs, reply *RequestVoteReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.VoteGranted = false
	reply.Term = rf.currentTerm

	if args.Term < rf.currentTerm {
		return nil
	}
	if args.Term > rf.currentTerm {
		rf.currentTerm = args.Term
		rf.votedFor = -1
		rf.state = Follower
	}

	// check if we can grant vote
	if rf.votedFor == -1 || rf.votedFor == args.CandidateId {
		// candidate's log up-to-date check (simplified)
		lastIndex := int64(len(rf.log))
		var lastTerm int64
		if lastIndex > 0 {
			lastTerm = rf.log[lastIndex-1].Term
		}

		upToDate := (args.LastLogTerm > lastTerm) || (args.LastLogTerm == lastTerm && args.LastLogIndex >= lastIndex)
		if upToDate {
			rf.votedFor = args.CandidateId
			reply.VoteGranted = true
			reply.Term = rf.currentTerm

			// reset election tiemr cause of vote granted
			select {
			case rf.heartbeatCh <- struct{}{}:
			default:
			}
		}
	}

	return nil
}

// Handle AppendEntries RPC call from leader.
func (rf *Raft) AppendEntries(ctx context.Context, args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	reply.Success = false
	reply.Term = rf.currentTerm

	if args.Term < rf.currentTerm {
		return nil
	}
	rf.state = Follower
	rf.currentTerm = args.Term

	// reset election timer on heartbeat
	select {
	case rf.heartbeatCh <- struct{}{}:
	default:
	}

	// SIMPLIFIED: accept entries by appending them if PrevLogIndex matches
	if args.PrevLogIndex > int64(len(rf.log)) {
		// missing entries
		return nil
	}

	// basic append (garbage ahh implementation, improve when possible)
	for i, e := range args.Entries {
		pos := args.PrevLogIndex + 1 + int64(i)
		if pos-1 < int64(len(rf.log)) {
			// overwrite
			rf.log[pos-1] = e
		} else {
			rf.log = append(rf.log, e)
		}
	}

	if args.LeaderCommit > rf.commitIndex {
		rf.commitIndex = minInt64(args.LeaderCommit, int64(len(rf.log))) // i had to define this separately. I'm tired.
		// apply to state machine up to commitIndex
		for rf.lastApplied < rf.commitIndex {
			rf.lastApplied++
			cmd := rf.log[rf.lastApplied-1].Command
			// send apply without holding lock (unlock temporarily), won't work any other way. i'm still tired...
			rf.mu.Unlock()
			rf.applyCh <- ApplyMsg{Index: rf.lastApplied, Command: cmd}
			rf.mu.Lock()
		}
	}

	reply.Success = true
	return nil // here just to make everything work... i don't even have anything to return bruh
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// ErrNotLeader is returned when API is invoked on non-leader.
var ErrNotLeader = errors.New("Invoking API on non-leader node")
