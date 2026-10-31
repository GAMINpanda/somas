package server

import (
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
		s.AddAgent(objects.NewBaseAmoeBoid(s))
	}
	return s
}

// RunStartOfIteration will replenish resources and run the political game.
func (s *Server) RunStartOfIteration(int) {}

// RunTurn will run the operational game: agents decide, environment updates.
func (s *Server) RunTurn(int, int) {}

// RunEndOfIteration will apply the lifecycle (deaths) and record metrics.
func (s *Server) RunEndOfIteration(int) { s.Iterations++ }
