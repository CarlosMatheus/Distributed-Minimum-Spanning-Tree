package DMST

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"os"
)

const Infinite = (1<<31) - 1

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
	// Log channel
	logChan chan string

	// GHS variables
	level int	   // LN
	state string  // SN
	fragment int   // FN
	findCount int
	inBranch int
	bestEdgeWeight int

	testEdge *Edge
	edgeList [] *Edge // todo initialize this variable on new Nodes
	edgeMap map[int]*Edge // todo initialize this variable on new Nodes
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

	go node.writeLog()

	for {
		if node.me == 2{
			args := &MessageArgs{
				Type: InitiateType,
				NodeLevel: 1,
				NodeState: SleepingState,
				NodeFragment: 3,
				EdgeWeight: 4,
			} 
			go func(peer int) {
				reply := &MessageReply{}
				node.sendMessage(1, args, reply)
			}(1)
		}

		if node.me == 1{
			node.handler()
		}
	}
}

func (node *Node) writeLog(){
	fmt.Print("START LOG GO ROUTINE")

	f, err := os.Create("log/" + strconv.Itoa(node.me) + ".txt")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	for{
		logEntry := <- node.logChan
		_, err := f.WriteString(logEntry)
		if err != nil {
			fmt.Println(err)
			f.Close()
			return
		}
	}
}

func (node *Node) handler() {
	log.Println("Starting Handler")
	for {
		msg := <-node.msgChan
		node.messageLog(msg)
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

func (node *Node) messageLog(msg *MessageArgs){
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

func (node *Node) getMinEdge() *Edge {
	var minEdge *Edge
	var minEdgeVal = Infinite
	for _, edge := range node.edgeMap {
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

func (node *Node) sendInitiate(level int, fragment int, state string, edgeWeight int) {
	args := &MessageArgs{
		FromID: node.me,
		Type: InitiateType,
		NodeLevel: level,
		NodeFragment: fragment,
		NodeState: state,
	} 
	go func(peer int) {
		reply := &MessageReply{}
		node.sendMessage(peer, args, reply)
	}(node.edgeMap[edgeWeight].targetNodeID)
}

func (node *Node) sendTest(message *MessageArgs, edgeID int) {
	// todo create send send test
}

func (node *Node) wakeupProcedure() {
	minEdge := node.getMinEdge()
	minEdge.state = BranchState
	node.level = 0
	node.state = FoundState
	node.findCount = 0
	node.connect(minEdge.targetNodeID)
}

func (node *Node) placeMessageEndOfQueue(msg *MessageArgs) {
	// todo verify this
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
			node.placeMessageEndOfQueue(msg)
		} else {
			node.sendInitiate(node.level +1, msg.EdgeWeight, FindState, node.edgeMap[msg.EdgeWeight].targetNodeID)
		}
	}
}

func (node *Node) responseToInitiate(msg *MessageArgs) {

	node.level = msg.NodeLevel
	node.fragment = msg.NodeFragment
	node.state = msg.NodeState
	node.inBranch = msg.FromID
	node.bestEdgeWeight = Infinite

	for edgeWeight, edge := range node.edgeMap {
		if edge.state == BranchState && edge.weight != msg.FromID{
			node.sendInitiate(node.level, node.fragment, node.state, node.edgeMap[msg.EdgeWeight].targetNodeID)
			if msg.NodeState == FindState {
				node.findCount ++
			}
		}
	}
	if msg.NodeState == FindState {
		node.procedureTest()
	}
}

func (node *Node) procedureTest(msg *MessageArgs) {
	minWeightedEdge := Infinite
	for _, edge := range node.edgeMap {
		if edge.state == BasicState {
			if edge.weight < minWeightedEdge {
				minWeightedEdge = edge.weight
			}
		}
	}
	if minWeightedEdge != Infinite {
		node.testEdge = node.edgeMap[minWeightedEdge]
		node.sendTest(msg, minWeightedEdge)
	} else {
		node.testEdge = nil
		node.reportProcedure()
	}
}

func (node *Node) responseToTest(msg *MessageArgs) {
	if node.state == SleepingState {
		node.wakeupProcedure()
	}
	if msg.NodeLevel > node.level {
		node.placeMessageEndOfQueue(msg)
	} else {
		if msg.NodeFragment != node.fragment {
			node.responseToAccept(msg.EdgeWeight)
		} else {
			if node.edgeMap[msg.EdgeWeight].state == BasicState {
				node.edgeMap[msg.EdgeWeight].state = RejectedState
				if node.testEdge.weight != msg.EdgeWeight {
					node.responseToReject(msg.EdgeWeight)
				} else {
					node.procedureTest()
				}
			}
		}
	}
}

func (node *Node) responseToAccept(weight int){
	node.testEdge = nil
	if weight < node.bestEdgeWeight{
		node.bestEdgeWeight = weight
	}
	node.reportProcedure()
}

func (node *Node) responseToReject(weight int){
	if node.edgeMap[weight].state == BasicState {
		node.edgeMap[weight].state = RejectedState
	}
	node.procedureTest()
}

func (node *Node) reportProcedure() {
	if node.findCount == 0 && node.testEdge == nil {
		node.state = FoundState
		args := &MessageArgs{
			FromID: node.me,
			Type: ReportType,
			BestEdgeWeight: node.bestEdgeWeight,
		} 
		go func(peer int) {
			reply := &MessageReply{}
			node.sendMessage(peer, args, reply)
		}(node.inBranch)
	}
}

func (node *Node) responseToReport(msg *MessageArgs) {
	if msg.FromID != node.inBranch {
		node.findCount -= 1
		if msg.BestEdgeWeight < node.bestEdgeWeight {
			node.bestEdgeWeight = msg.BestEdgeWeight
		}
		node.reportProcedure()
	} else if node.state == FindState {
		node.placeMessageEndOfQueue(msg)
	} else if msg.BestEdgeWeight > node.bestEdgeWeight {
		node.changeCoreProcedure()
	} else if msg.BestEdgeWeight == node.bestEdgeWeight && node.bestEdgeWeight == Infinite {
		node.halt()
	}
}

func (node *Node) changeCoreProcedure() {
	if node.edgeMap[node.bestEdgeWeight].state == BranchState {
		args := &MessageArgs{
			FromID: node.me,
			Type: ChangeCoreType,
		} 
		go func(peer int) {
			reply := &MessageReply{}
			node.sendMessage(peer, args, reply)
		}(node.edgeMap[node.bestEdgeWeight].targetNodeID)
	} else {
		args := &MessageArgs{
			FromID: node.me,
			Type: ConnectType,
			NodeLevel: node.level,
		} 
		go func(peer int) {
			reply := &MessageReply{}
			node.sendMessage(peer, args, reply)
		}(node.edgeMap[node.bestEdgeWeight].targetNodeID)
		node.edgeMap[node.bestEdgeWeight].state = BranchState
	}
}

func (node *Node) logNode() {
	node.logChan <- fmt.Sprintf("TIME >> %v >> NODE >> %d %s\n", time.Now(), node.me, node.state)
}

func (node *Node) logEdges() {
	for k, v := range node.edgeMap {
		node.logChan <-fmt.Sprintf("TIME >> %v >> EDGE >> %d %d %d %s\n", time.Now(), node.me, k, v.weight, v.state)
	}
}

func (node *Node) halt() {
	// todo, stop function

	node.logNode()
	node.logEdges()
}
