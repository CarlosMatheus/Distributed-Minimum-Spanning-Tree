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
	msgChan chan *MessageArgs

	// GHS variables
	nodeLevel int
	nodeStatus string
	nodeFragment int
	findCount int
	inBranch int
	bestEdge Edge
	testEdge Edge
	edgeList [] Edge // todo initialize this variable on new Nodes



	currentState *util.ProtectedString
	currentTerm  int
	votedFor     int

	// Goroutine communication channels
	electionTick    <-chan time.Time
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

		// Communication channel
		msgChan: make(chan *MessageArgs, 20*len(peers)),
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

func (node *Node) connect(targetNodeID int) {
	// todo create connect function
}

func (node *Node) wakeupProcedure() {
	minEdge := node.getMinEdge()
	minEdge.edgeStatus = BRANCH_STATE
	node.nodeLevel = 0
	node.nodeStatus = FOUND_STATE
	node.findCount = 0
	node.connect(minEdge.targetNodeID)
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

	for {
		if node.me == 2{
			args := &MessageArgs{
				Type: "Teste",
				NodeLevel: 1,
				NodeStatus: 2,
				NodeFragement: 3,
				EdgeWeight: 4,
			} 
			// go func(peer int) {
				reply := &MessageReply{}
				node.sendMessage(1, args, reply)
			// }(1)
		}

		if node.me == 1{
			node.handler()
		}
	}
}



// followerSelect implements the logic to handle messages from distinct
// events when in follower state.
func (node *Node) handler() {
	log.Println("Starting Handler")
	for {
		msg := <-node.msgChan
		switch msg.Type{
		case "Teste":
			log.Println("Message received")
			log.Println(msg)
			return
		}
	}
}
