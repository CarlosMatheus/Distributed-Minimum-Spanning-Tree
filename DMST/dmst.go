package DMST

import (
	_ "Distributed-Minimum-Spanning-Tree/util"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

const SleepingState = "Sleeping"
const FindState = "Find"
const FoundState = "Found"
const RejectedState = "Rejected"
const BranchState = "Branch"
const BasicState = "Basic"

// Node is the struct that hold all information that is used by this instance of node
type Node struct {
	sync.Mutex

	serv *server
	done chan struct{}

	peers map[int]string
	me    int

	// Communication channels
	sendMsgChan chan *sendMessageArgs
	receiveMsgChan chan *receiveMsgArgs

	// GHS variables
	nodeLevel int
	nodeStatus string
	nodeFragment int
	findCount int
	inBranch int
	bestEdge Edge
	testEdge Edge

	edgeList [] Edge // todo initialize this variable on new Nodes
}

type Edge struct {
	weight int
	edgeStatus string  // SE
	targetNodeID int
}

// NewNode create a new node object and return a pointer to it.
func NewNode(peers map[int]string, me int) *Node {
	var err error

	// 0 is reserved to represent undefined vote/leader
	if me == 0 {
		panic(errors.New("Reserved instanceID('0')"))
	}

	node := &Node{
		done: make(chan struct{}),

		peers: peers,
		me:    me,

		// Communication channels
		sendMsgChan: make(chan *sendMessageArgs, 20*len(peers)),
		receiveMsgChan: make(chan *receiveMsgArgs, 20*len(peers)),
	}

	node.serv, err = newServer(node, peers[me])
	if err != nil {
		panic(err)
	}

	go node.loop()

	return node
}

// Done returns a channel that will be used when the instance is done.
func (node *Node) Done() <-chan struct{} {
	return node.done
}

func debugPrint(s string){
	fmt.Print(s)
}

func (node *Node) awakeningResponse() {
	// Is a reponse to a awake call, this can only occur to sleeping node
	if node.nodeStatus == SleepingState {
		// ok
		wakeupProcedure(node)
	} else {
		// problem
		debugPrint("Error: awakeningResponse called when node not in sleeping state")
	}
}

func (node *Node) getMinEdge() Edge {
	var minEdge Edge
	var minEdgeVal = (1<<31) - 1
	for _, edge := range node.edgeList {
		if edge.weight < minEdgeVal {
			minEdgeVal = edge.weight
			minEdge = edge
		}
	}
	return minEdge
}


func (node *Node) wakeupProcedure() {
	//
	//minEdge := getMinEdge(node)

}

func (node *Node) onTest(int level) {

}

func (node *Node) onAccept(edge Edge){
	node.testEdge = nil
	if edge.weight < node.bestEdge.weight {
		node.bestEdge = edge
	}
	// execute report
}

func (node *Node) onReject(edge Edge){
	if edge.edgeStatus == BasicState {
		edge.edgeStatus = RejectedState
	}
	// execute test
}

// All changes to Node structure should occur in the context of this routine.
// This way it's not necessary to use synchronizers to protect shared data.
// To send data to each of the states, use the channels provided.
func (node *Node) loop() {

	err := node.serv.startListening()
	if err != nil {
		panic(err)
	}

	node.currentState.Set(follower)
	for {
		switch node.currentState.Get() {
		case follower:
			node.followerSelect()
		case candidate:
			node.candidateSelect()
		case leader:
			node.leaderSelect()
		}
	}
}

// followerSelect implements the logic to handle messages from distinct
// events when in follower state.
func (node *Node) followerSelect() {
	log.Println("[FOLLOWER] Run Logic.")
	node.resetElectionTimeout()
	for {
		select {
		case <-node.electionTick:
			log.Println("[FOLLOWER] Election timeout.")
			node.currentState.Set(candidate)
			return

		case rv := <-node.requestVoteChan:
			///////////////////
			//  MODIFY HERE  //
			reply := &RequestVoteReply{
				Term: node.currentTerm,
			}

			log.Printf("[FOLLOWER] Vote denied to '%v' for term '%v'.\n", node.peers[rv.CandidateID], node.currentTerm)

			reply.VoteGranted = false
			rv.replyChan <- reply
			break
			// END OF MODIFY //
			///////////////////

		case ae := <-node.appendEntryChan:
			///////////////////
			//  MODIFY HERE  //
			reply := &AppendEntryReply{
				Term: node.currentTerm,
			}

			log.Printf("[FOLLOWER] Accept AppendEntry from '%v'.\n", node.peers[ae.LeaderID])
			reply.Success = true
			ae.replyChan <- reply
			break
			// END OF MODIFY //
			///////////////////
		}
	}
}

// candidateSelect implements the logic to handle messages from distinct
// events when in candidate state.
func (node *Node) candidateSelect() {
	log.Println("[CANDIDATE] Run Logic.")
	// Candidates (§5.2):
	// Increment currentTerm, vote for self
	node.currentTerm++
	node.votedFor = node.me
	voteCount := 1

	log.Printf("[CANDIDATE] Running for term '%v'.\n", node.currentTerm)
	// Reset election timeout
	node.resetElectionTimeout()
	// Send RequestVote RPCs to all other servers
	replyChan := make(chan *RequestVoteReply, 10*len(node.peers))
	node.broadcastRequestVote(replyChan)

	for {
		select {
		case <-node.electionTick:
			// If election timeout elapses: start new election
			log.Println("[CANDIDATE] Election timeout.")
			node.currentState.Set(candidate)
			return
		case rvr := <-replyChan:
			///////////////////
			//  MODIFY HERE  //

			if rvr.VoteGranted {
				log.Printf("[CANDIDATE] Vote granted by '%v'.\n", node.peers[rvr.peerIndex])
				voteCount++
				log.Println("[CANDIDATE] VoteCount: ", voteCount)
				break
			}
			log.Printf("[CANDIDATE] Vote denied by '%v'.\n", node.peers[rvr.peerIndex])

			// END OF MODIFY //
			///////////////////

		case rv := <-node.requestVoteChan:
			///////////////////
			//  MODIFY HERE  //
			reply := &RequestVoteReply{
				Term: node.currentTerm,
			}

			log.Printf("[CANDIDATE] Vote denied to '%v' for term '%v'.\n", node.peers[rv.CandidateID], node.currentTerm)
			reply.VoteGranted = false
			rv.replyChan <- reply
			break
			// END OF MODIFY //
			///////////////////

		case ae := <-node.appendEntryChan:
			///////////////////
			//  MODIFY HERE  //
			reply := &AppendEntryReply{
				Term: node.currentTerm,
			}

			log.Printf("[CANDIDATE] Accept AppendEntry from '%v'.\n", node.peers[ae.LeaderID])
			reply.Success = true
			ae.replyChan <- reply
			break
			// END OF MODIFY //
			///////////////////
		}
	}
}

// leaderSelect implements the logic to handle messages from distinct
// events when in leader state.
func (node *Node) leaderSelect() {
	log.Println("[LEADER] Run Logic.")
	replyChan := make(chan *AppendEntryReply, 10*len(node.peers))
	node.broadcastAppendEntries(replyChan)

	heartbeat := time.NewTicker(node.broadcastInterval())
	defer heartbeat.Stop()

	broadcastTick := make(chan time.Time)
	defer close(broadcastTick)

	go func() {
		for t := range heartbeat.C {
			broadcastTick <- t
		}
	}()

	for {
		select {
		case <-broadcastTick:
			node.broadcastAppendEntries(replyChan)
		case aet := <-replyChan:
			///////////////////
			//  MODIFY HERE  //
			_ = aet
			// END OF MODIFY //
			///////////////////
		case rv := <-node.requestVoteChan:
			///////////////////
			//  MODIFY HERE  //

			reply := &RequestVoteReply{
				Term: node.currentTerm,
			}

			log.Printf("[LEADER] Vote denied to '%v' for term '%v'.\n", node.peers[rv.CandidateID], node.currentTerm)
			reply.VoteGranted = false
			rv.replyChan <- reply
			break

			// END OF MODIFY //
			///////////////////

		case ae := <-node.appendEntryChan:
			///////////////////
			//  MODIFY HERE  //
			reply := &AppendEntryReply{
				Term: node.currentTerm,
			}

			log.Printf("[LEADER] Accept AppendEntry from '%v'.\n", node.peers[ae.LeaderID])
			reply.Success = true
			ae.replyChan <- reply
			break
			// END OF MODIFY //
			///////////////////
		}
	}
}
