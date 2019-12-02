package DMST

import (
	"math/rand"
	"time"
)

// BroadcastInterval << ElectionTimeout << MTBF
// MTBF = Mean Time Between Failures

func (node *Node) electionTimeout() time.Duration {
	timeout := minElectionTimeoutMilli + rand.Intn(maxElectionTimeoutMilli-minElectionTimeoutMilli)
	return time.Duration(timeout) * time.Millisecond
}

func (node *Node) broadcastInterval() time.Duration {
	timeout := minElectionTimeoutMilli / 10
	return time.Duration(timeout) * time.Millisecond
}

func (node *Node) resetElectionTimeout() {
	node.electionTick = time.NewTimer(node.electionTimeout()).C
}
