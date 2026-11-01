package server

import (
	"math/rand/v2"
	"time"

	"github.com/GAMINpanda/somas/internal/common/objects"
	baseserver "github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"
)

// Server embeds the platform's BaseServer and we complete the inheretted funcs (RunStartOfIteration,RunTurn and RunEndOfIteration)
type Server struct {
	*baseserver.BaseServer[objects.IAmoeBoid]     // the platform's BaseServer
	Iterations                                int // number of iterations completed
}

// Initialise returns a Server which adds the numAgents of BaseAmoeBoids, ready to Start.
func Initialise(numAgents, iterations, turns, messageBandwidth int, timeout time.Duration) *Server {
	s := &Server{BaseServer: baseserver.CreateBaseServer[objects.IAmoeBoid](iterations, turns, timeout, messageBandwidth)}
	s.SetGameRunner(s) // BaseServer can't see our methods through embedding, so pass them in
	for range numAgents {
		s.AddAgent(objects.CreateBaseAmoeBoid(s))
	}
	return s
}

// RunStartOfIteration will do the voluntary association and deciding of rules
func (s *Server) RunStartOfIteration(int) {}

// Every agent randomly provisions and appropriates for now later add modfying rules and replenishing
func (s *Server) RunTurn(int, int) {
	for _, ag := range s.GetAgentMap() {
		ag.SendProvisionMessage(rand.Float64() * 100)
	}

	for _, ag := range s.GetAgentMap() {
		ag.SendAppropriationMessage(rand.Float64() * 100)
	}
}

// RunEndOfIteration will increment coutner and disband
func (s *Server) RunEndOfIteration(int) { s.Iterations++ }
