package main

import (
	"fmt"
	"time"

	"github.com/GAMINpanda/somas/internal/server"
)

func main() {
	s := server.Initialise(10, 5, 10, 10, time.Millisecond) // Initlise with 10 agents, 5 iterations, 10 turns per iteration a message bw of 10 and a timeout of 1ms (can hardcode certain vals and not param it)
	s.Start()
	fmt.Printf("SOMAS Platform Finished: %d cycles %d agent registered\n", s.Iterations, len(s.GetAgentMap()))
}
