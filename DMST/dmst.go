package DMST

import (
	"errors"
	"Distributed-Minimum-Spanning-Tree/util"
	"log"
	"sync"
	"time"
)

var nodeStatusList = map[int]string{
	1: "Sleeping",
	2: "Find",
	3: "Found",
}

var edgeStatusList = map[int]string{
	1: "Rejected",
	2: "Branch",
	3: "Basic",
}

// Node is the struct that hold all information that is used by this instance
// of node.
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
	nodeStatus int
	nodeFragement int
	findCount int
	inBranch int
	bestEdge int
	bestWt int
	testEdge int



	currentState *util.ProtectedString
	currentTerm  int
	votedFor     int

	// Goroutine communication channels
	electionTick    <-chan time.Time

}

type Edge struct {
	weight int
	edgeStatus int
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
