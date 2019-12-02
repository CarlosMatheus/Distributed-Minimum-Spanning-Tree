package DMST

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"os"
	"path"
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
	edgeMap map[int]*Edge // todo initialize this variable on new Nodes
}

type Edge struct {
	weight       int
	state        string // SE
	targetNodeID int
}

func initializeEdgeMap(nIDs []string, nWTs []string) map[int]*Edge{
	var edge *Edge
	var intID int
	var intWT int
	edgeMap := make(map[int]*Edge)

	for i, _ := range nIDs {
		intWT, _ = strconv.Atoi(nWTs[i])
		intID, _ = strconv.Atoi(nIDs[i])
		edge = &Edge{
			weight: intWT,
			state: BasicState,
			targetNodeID: intID,
		}
		edgeMap[edge.weight] = edge
	}

	return edgeMap
}

// NewNode create a new node object and return a pointer to it.
func NewNode(peers map[int]string, nIDs []string, nWTs []string, me int) *Node {
	var err error

	// 0 is reserved to represent undefined vote/leader
	if me == 0 {
		panic(errors.New("Reserved instanceID('0')"))
	}

	edgeMap := initializeEdgeMap(nIDs, nWTs)

	node := &Node{
		done: make(chan struct{}),
		peers: peers,
		me:    me,
		state: SleepingState,
		edgeMap: edgeMap,
		logChan: make(chan string, 20),
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

	if node.me == 1 {
		node.awakeningResponse()
	}

	for {
			node.handler()
	}
}

func (node *Node) writeLog(){
	fmt.Print("START LOG GO ROUTINE\n")

	f, err := os.Create(path.Join("logs", strconv.Itoa(node.me) + ".txt"))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer f.Close()

	node.logNode()
	node.logEdges()

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
				node.responseToConnect(msg)
			case InitiateType:
				node.responseToInitiate(msg)
			case TestType:
				node.responseToTest(msg)
			case AcceptType:
				node.responseToAccept(msg)
			case RejectType:
				node.responseToReject(msg)
			case ReportType:
				node.responseToReport(msg)
			case ChangeCoreType:
				node.responseToChangeCore(msg)
		}
		node.logNode()
		node.logEdges()
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

func (node *Node) connect(EdgeWeight int, targetNodeID int) {
	args := &MessageArgs{
		FromID: node.me,
		Type: ConnectType,
		NodeLevel: node.level,
		EdgeWeight: EdgeWeight,
	} 
	go func(peer int) {
		reply := &MessageReply{}
		node.sendMessage(peer, args, reply)
	}(targetNodeID)
}

func (node *Node) sendInitiate(level int, fragment int, state string, edgeWeight int) {
	args := &MessageArgs{
		FromID: node.me,
		Type: InitiateType,
		NodeLevel: level,
		NodeFragment: fragment,
		NodeState: state,
		EdgeWeight: edgeWeight,
	}
	go func(peer int) {
		reply := &MessageReply{}
		node.sendMessage(peer, args, reply)
	}(node.edgeMap[edgeWeight].targetNodeID)
}

func (node *Node) sendTest(level int, fragment int, edgeWeight int) {
	args := &MessageArgs{
		FromID: node.me,
		Type: TestType,
		NodeLevel: level,
		NodeFragment: fragment,
		EdgeWeight: edgeWeight,
	} 
	go func(peer int) {
		reply := &MessageReply{}
		node.sendMessage(peer, args, reply)
	}(node.edgeMap[edgeWeight].targetNodeID)
}

func (node *Node) wakeupProcedure() {
	minEdge := node.getMinEdge()
	minEdge.state = BranchState
	node.level = 0
	node.state = FoundState
	node.findCount = 0
	node.connect(minEdge.weight, minEdge.targetNodeID)
}

func (node *Node) placeMessageEndOfQueue(msg *MessageArgs) {
	time.Sleep(100 * time.Millisecond)
	node.msgChan <- msg
}

func (node *Node) responseToConnect(msg *MessageArgs) {
	if node.state == SleepingState {
		node.wakeupProcedure()
	}
	if msg.NodeLevel < node.level {
		node.edgeMap[msg.EdgeWeight].state = BranchState
		node.sendInitiate(node.level, node.fragment, node.state, msg.EdgeWeight)
		if node.state == FindState {
			node.findCount++
		}
	} else {
		if node.edgeMap[msg.EdgeWeight].state == BasicState {
			node.placeMessageEndOfQueue(msg)
		} else {
			node.sendInitiate(node.level +1, node.fragment, FindState, msg.EdgeWeight)
		}
	}
}

func (node *Node) responseToInitiate(msg *MessageArgs) {

	node.level = msg.NodeLevel
	node.fragment = msg.NodeFragment
	node.state = msg.NodeState
	node.inBranch = msg.EdgeWeight
	node.bestEdgeWeight = Infinite

	for _, edge := range node.edgeMap {
		if edge.state == BranchState && edge.weight != msg.EdgeWeight {
			node.sendInitiate(node.level, node.fragment, node.state, edge.weight)
			if msg.NodeState == FindState {
				node.findCount ++
			}
		}
	}
	if msg.NodeState == FindState {
		node.testProcedure()
	}
}

func (node *Node) testProcedure() {
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
		node.sendTest(node.level, node.fragment, minWeightedEdge)
	} else {
		node.testEdge = nil
		node.reportProcedure()
	}
}

func (node *Node) sendAccept(EdgeWeight int) {
	args := &MessageArgs{
		FromID: node.me,
		Type: AcceptType,
		EdgeWeight: EdgeWeight,
	} 
	go func(peer int) {
		reply := &MessageReply{}
		node.sendMessage(peer, args, reply)
	}(node.edgeMap[EdgeWeight].targetNodeID)
}

func (node *Node) sendReject(EdgeWeight int) {
	args := &MessageArgs{
		FromID: node.me,
		Type: RejectType,
		EdgeWeight: EdgeWeight,
	} 
	go func(peer int) {
		reply := &MessageReply{}
		node.sendMessage(peer, args, reply)
	}(node.edgeMap[EdgeWeight].targetNodeID)
}

func (node *Node) responseToTest(msg *MessageArgs) {
	if node.state == SleepingState {
		node.wakeupProcedure()
	}
	if msg.NodeLevel > node.level {
		node.placeMessageEndOfQueue(msg)
	} else{
		if msg.NodeFragment != node.fragment {
			node.sendAccept(msg.EdgeWeight)
		} else {
			if node.edgeMap[msg.EdgeWeight].state == BasicState {
				node.edgeMap[msg.EdgeWeight].state = RejectedState
			}
			if node.testEdge != nil && node.testEdge.weight == msg.EdgeWeight {
				node.sendReject(msg.EdgeWeight)
			} else {
				node.testProcedure()
			}
		} 
	}
}


func (node *Node) responseToAccept(msg *MessageArgs){
	node.testEdge = nil
	if msg.EdgeWeight < node.bestEdgeWeight{
		node.bestEdgeWeight = msg.EdgeWeight
	}
	node.reportProcedure()
}

func (node *Node) responseToReject(msg *MessageArgs){
	if node.edgeMap[msg.EdgeWeight].state == BasicState {
		node.edgeMap[msg.EdgeWeight].state = RejectedState
	}
	node.testProcedure()
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
		node.connect(node.bestEdgeWeight, node.edgeMap[node.bestEdgeWeight].targetNodeID)
		node.edgeMap[node.bestEdgeWeight].state = BranchState
	}
}

func (node *Node) logNode() {
	node.logChan <- fmt.Sprintf("TIME >> %v >> NODE >> %d %s\n", time.Now().UnixNano() , node.me, node.state)
}

func (node *Node) logEdges() {
	for _, v := range node.edgeMap {
		node.logChan <- fmt.Sprintf("TIME >> %v >> EDGE >> %d %d %d %s\n", time.Now().UnixNano() , node.me, v.targetNodeID, v.weight, v.state)
	}
}

func (node *Node) responseToChangeCore(msg *MessageArgs) {
	node.changeCoreProcedure()
}

func (node *Node) halt() {
	node.logNode()
	node.logEdges()
}
