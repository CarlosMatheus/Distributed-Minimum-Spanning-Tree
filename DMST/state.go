package DMST

const (
	minElectionTimeoutMilli int = 500
	maxElectionTimeoutMilli int = 5000
)

const (
	follower  string = "Follower"
	candidate string = "Candidate"
	leader    string = "Leader"
)

// resetState will reset the state of a node when new term is discovered.
func (node *Node) resetState(term int) {
	node.currentTerm = term
	node.votedFor = 0
	node.currentState.Set(follower)
}
