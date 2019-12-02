package DMST

import (
	"Distributed-Minimum-Spanning-Tree/util"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// Node possible states
const SleepingState = "Sleeping"
const FindState = "Find"
const FoundState = "Found"

// Edge possible states
const RejectedState = "Rejected"
const BranchState = "Branch"
const BasicState = "Basic"

// Connection possible types
const ConnectType = "Connect"
const InitiateType = "Initiate"
const TestType = "Test"
const AcceptType = "Accept"
const RejectType = "Reject"
const ReportType = "Report"
const ChangeCoreType = "Change-core"

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
	level int	   // LN
	state string  // SN
	fragment int   // FN
	findCount int
	inBranch int
	bestEdge Edge
	testEdge Edge
	edgeList [] Edge // todo initialize this variable on new Nodes

	edgeMap map[int]*Edge // todo initialize this variable on new Nodes

	currentState *util.ProtectedString
	currentTerm  int
	votedFor     int

	// Goroutine communication channels
	electionTick    <-chan time.Time
}

type Edge struct {
	weight       int
	state        string // SE
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

func (node *Node) loop() {

	err := node.serv.startListening()
	if err != nil {
		panic(err)
	}

	for {
		if node.me == 2{
			args := &MessageArgs{
				Type: InitiateType,
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

func (node *Node) handler() {
	log.Println("Starting Handler")
	for {
		msg := <-node.msgChan
		node.messageLog(*msg)
		switch msg.Type{
			case ConnectType:
				return
			case InitiateType:
				return
			case TestType:
				return
			case AcceptType:
				return
			case RejectType:
				return
			case ReportType:
				return
			case ChangeCoreType:
				return
		}
	}
}

func (node *Node) messageLog(msg MessageArgs){
	log.Printf("[NODE %d] %s message received from node %d", node.me, msg.Type, msg.FromID)
}

func debugPrint(s string){
	fmt.Print(s)
}

func (node *Node) awakeningResponse() {
	// Is a reponse to a awake call, this can only occur to sleeping node
	if node.state == SleepingState {
		// ok
		node.wakeupProcedure()
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

func (node *Node) sendInitiate() {
	// todo create send initiate function
}

func (node *Node) wakeupProcedure() {
	minEdge := node.getMinEdge()
	minEdge.state = BranchState
	node.level = 0
	node.state = FoundState
	node.findCount = 0
	node.connect(minEdge.targetNodeID)
}

func (node *Node) placeReceivedMessageOnEndOfQueue(msg *MessageArgs) {
	node.msgChan <- msg
}

func (node *Node) responseToConnect(msg *MessageArgs) {
	if node.state == SleepingState {
		node.wakeupProcedure()
	}
	if msg.NodeLevel < node.level {
		node.edgeMap[msg.EdgeWeight].state = BranchState
		if node.state == FindState {
			node.findCount++
		}
	} else {
		if node.edgeMap[msg.EdgeWeight].state == BasicState {
			node.placeReceivedMessageOnEndOfQueue(msg)
		} else {
			node.sendInitiate() // todo
		}
	}
}

func (node *Node) responseToInitiate(msg *MessageArgs) {
	node.level = msg.NodeLevel
	node.fragment = msg.NodeFragement
	node.state = msg.NodeStatus
	node.inBranch = msg.FromID
}

func (node *Node) onTest(level int, fragment int, edge Edge) {
	if node.state == SleepingState {
		node.wakeupProcedure()
	}
	if level > node.level {
		// place msg to the queue
	} else {
		if fragment != node.fragment {
			node.onAccept(edge)
		} else {
			if edge.state == BasicState {
				edge.state = RejectedState
				if node.testEdge.weight != edge.weight {
					node.onReject(edge)
				} else {
					// execute test
				}
			}
		}
	}
}

func (node *Node) onAccept(edge Edge){
	//node.testEdge = nil
	if edge.weight < node.bestEdge.weight {
		node.bestEdge = edge
	}
	// execute report
}

func (node *Node) onReject(edge Edge){
	if edge.state == BasicState {
		edge.state = RejectedState
	}
	// execute test
}
