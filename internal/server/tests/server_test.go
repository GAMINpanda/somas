package server_test

import (
	"testing"
	"time"

	"github.com/GAMINpanda/somas/internal/server"
)

func TestInitialise(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, time.Millisecond) // 10 agents, 5 iterations, 5 turns, 10 second timeout and a message bw of 10

	if got := len(s.GetAgentMap()); got != 10 {
		t.Errorf("agents = %d, want 10", got)
	}
}

func TestRunGame(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, time.Microsecond) //test with a us turn timeout
	s.Start()

	if s.Iterations != 5 {
		t.Errorf("cycles completed = %d, want 5", s.Iterations)
	}
}
