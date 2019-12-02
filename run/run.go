package main

import (
	"flag"
	"Distributed-Minimum-Spanning-Tree/DMST"
	// "log"
	"fmt"
	"strconv"
	"strings"
)

// Command line parameters
var (
	ID = flag.Int("id", 0, "ID of the node")
	neighborIDs = flag.String("nid", "0", "IDs of the neighbors")
	//RNG = Random Number Generator
)

func main() {
	flag.Parse()

	nIDs := strings.Split(*neighborIDs, ",")

	peers := makePeers(*ID, nIDs)

	fmt.Print(peers)

	// if _, ok := peers[*nodeID]; !ok {
	// 	log.Fatalf("[MAIN] Invalid instance id.\n")
	// }

	node := DMST.NewNode(peers, *ID)

	 <-node.Done()
}

func makePeers(ID int, nIDs []string) map[int]string{
	peers := make(map[int]string)
	
	base := 3000

	port := base + ID

	peers[ID] = "localhost:" + strconv.Itoa(port)

	for _, s := range nIDs {
		intID, _ := strconv.Atoi(s)
		port = base + intID

		peers[intID] = "localhost:" + strconv.Itoa(port) 
	}

	return peers
}
