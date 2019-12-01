package raft

import (
	"math/rand"
	"time"
)

// BroadcastInterval << ElectionTimeout << MTBF
// MTBF = Mean Time Between Failures

func (raft *Raft) electionTimeout() time.Duration {
	timeout := minElectionTimeoutMilli + rand.Intn(maxElectionTimeoutMilli-minElectionTimeoutMilli)
	return time.Duration(timeout) * time.Millisecond
}

func (raft *Raft) broadcastInterval() time.Duration {
	timeout := minElectionTimeoutMilli / 10
	return time.Duration(timeout) * time.Millisecond
}

func (raft *Raft) resetElectionTimeout() {
	raft.electionTick = time.NewTimer(raft.electionTimeout()).C
}
