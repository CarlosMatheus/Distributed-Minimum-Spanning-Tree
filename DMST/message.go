package DMST

// MessageArgs is invoked by leader to replicate log entries; also used as
// heartbeat.
// Term  	- leader’s term
// leaderId - so follower can redirect clients
type MessageArgs struct {
	Type         string
	NodeLevel    int
	NodeStatus   int
	NodeFragment int
	EdgeWeight   int

	// internal
	replyChan chan *MessageReply
}

// MessageReply contains data to be returned to leader
// Term - currentTerm, for leader to update itself
// Success - true if follower contained entry matching PrevLogIndex and PrevLogTerm
type MessageReply struct {
	Term    int
	Success bool

	// internal
	peerIndex int
}

// Message is called by other instances of Node. It'll write the args received
// in the appendEntryChan.
func (rpc *RPC) Message(args *MessageArgs, reply *MessageReply) error {
	args.replyChan = make(chan *MessageReply)
	rpc.node.msgChan <- args
	*reply = *<-args.replyChan
	return nil
}

// broadcastAppendEntries will send Message to all peers
func (node *Node) broadcastAppendEntries(replyChan chan<- *MessageReply) {
	args := &MessageArgs{
	}

	for peerIndex := range node.peers {
		if peerIndex != node.me {
			go func(peer int) {
				reply := &MessageReply{}
				ok := node.sendMessage(peer, args, reply)
				if ok {
					reply.peerIndex = peer
					replyChan <- reply
				}
			}(peerIndex)
		}
	}
}

// sendMessage will send Message to a peer
func (node *Node) sendMessage(peerIndex int, args *MessageArgs, reply *MessageReply) bool {
	err := node.CallHost(peerIndex, "Message", args, reply)
	if err != nil {
		return false
	}
	return true
}
