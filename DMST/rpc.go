package DMST

import "net/rpc"

// RPC is the struct that implements the methods exposed to RPC calls
type RPC struct {
	node *Node
}

// CallHost will communicate to another host through it's RPC public API.
func (node *Node) CallHost(index int, method string, args interface{}, reply interface{}) error {
	var (
		err    error
		client *rpc.Client
	)

	client, err = rpc.Dial("tcp", node.peers[index])
	if err != nil {
		return err
	}

	defer client.Close()

	err = client.Call("RPC."+method, args, reply)

	if err != nil {
		return err
	}

	return nil
}
