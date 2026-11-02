package server

import (
	"maps"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/GAMINpanda/somas/internal/common/objects"
	baseserver "github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"
)

const (
	initialCPR   = 500.0
	capacityCPR  = 1000.0 // K: the most the patch can hold
	growthRate   = 0.5    // r
	thresholdCTK = 30.0   // T
	sunPerRound  = 10.0
	metabolism   = 5.0
)

// Server embeds the platform's BaseServer and we complete the inheretted funcs (RunStartOfIteration,RunTurn and RunEndOfIteration)
// One platform iteration is one round (M) and one platform turn is one of the N turns
type Server struct {
	*baseserver.BaseServer[objects.IAmoeBoid]         // the platform's BaseServer
	Iterations                                int     // number of iterations completed
	CPR                                       float64 // vegetables left in the common-pool resource
	CTK                                       float64 // sunlight provisioned to the collective threshold kitty this round
}

// Initialise returns a Server which adds the numAgents of BaseAmoeBoids, ready to Start.
func Initialise(numAgents, iterations, turns, messageBandwidth int, timeout time.Duration) *Server {
	s := &Server{
		BaseServer: baseserver.CreateBaseServer[objects.IAmoeBoid](iterations, turns, timeout, messageBandwidth),
		CPR:        initialCPR,
	}
	s.SetGameRunner(s) 
	for range numAgents {
		s.AddAgent(objects.CreateBaseAmoeBoid(s))
	}
	return s
}

// Each round starts with sunlight, varied by cloud cover
// RunStartOfIteration will also do the voluntary association and deciding of rules
func (s *Server) RunStartOfIteration(int) {
	for _, ag := range s.GetAgentMap() {
		ag.AddBrain(sunPerRound * (0.8 + rand.Float64()*0.4))
	}
}

// One randomly selected agent appropriates each turn
func (s *Server) RunTurn(int, int) {
	agents := slices.Collect(maps.Values(s.GetAgentMap()))
	if s.CPR <= 0 || len(agents) == 0 {
		return
	}
	ag := agents[rand.IntN(len(agents))]
	taken := max(0, min(ag.DecideAppropriation(s.CPR), s.CPR))
	s.CPR -= taken
	ag.Eat(taken)
	ag.SendAppropriationMessage(taken)
}

// Provision, replenish, then remove any agents that died
// RunEndOfIteration will also disband
func (s *Server) RunEndOfIteration(int) {
	for _, ag := range s.GetAgentMap() {
		given := ag.SpendBrain(ag.DecideProvision())
		s.CTK += given
		ag.SendProvisionMessage(given)
	}
	if s.CTK >= thresholdCTK {
		s.CPR += growthRate * s.CPR * (1 - s.CPR/capacityCPR)
	}
	s.CTK = 0
	for _, ag := range s.GetAgentMap() {
		if !ag.Metabolise(metabolism) {
			s.RemoveAgent(ag)
		}
	}
	s.Iterations++
}
