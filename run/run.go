package main

import (
	"flag"
	"hash/fnv"
	// "labRaft/raft"
	// "log"
	"math/rand"
	"time"
	"strings"
	"fmt"
	"strconv"
)

// Command line parameters
var (
	ID = flag.Int("id", 0, "ID of the node")
	neighborIDs = flag.String("nid", "0", "IDs of the neighbors")
	seed       = flag.String("seed", "", "Seed for RNG")
	//RNG = Random Number Generator
)

func main() {
	flag.Parse()

	if *seed != "" {
		h := fnv.New32a()
		h.Write([]byte(*seed))
		rand.Seed(int64(h.Sum32()))
	} else {
		rand.Seed(time.Now().UnixNano())
	}

	nIDs := strings.Split(*neighborIDs, ",")


	peers := makePeers(*ID, nIDs)

	// fmt.Print(peers)

	// if _, ok := peers[*nodeID]; !ok {
	// 	log.Fatalf("[MAIN] Invalid instance id.\n")
	// }

	node := node.NewNode(peers, *nodeID)

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
