package server

import (
	"maps"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/GAMINpanda/somas/internal/common/objects"
	"github.com/GAMINpanda/somas/internal/patch"
	"github.com/GAMINpanda/somas/internal/recorder"
	"github.com/GAMINpanda/somas/internal/weather"
	baseserver "github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"
	"github.com/google/uuid"
)

const (
	initialCPR   = 500.0
	capacityCPR  = 1000.0 // K: the most the patch can hold
	growthRate   = 0.5    // r
	thresholdCTK = 30.0   // T
	sunPerRound  = 10.0
	metabolism   = 5.0
)

// The vegetable patch is a GridWidth x GridHeight grid of cells, used for placing agents and for the UI
const (
	GridWidth  = 10
	GridHeight = 10
	numClouds  = 4
)

// Position is a cell on the patch grid, with (0, 0) in the top left
type Position struct {
	X, Y int
}

// Server embeds the platform's BaseServer and we complete the inheretted funcs (RunStartOfIteration,RunTurn and RunEndOfIteration)
// One platform iteration is one round (M) and one platform turn is one of the N turns
type Server struct {
	*baseserver.BaseServer[objects.IAmoeBoid]                                   // the platform's BaseServer
	Iterations                                int                               // number of iterations completed
	CPR                                       float64                           // vegetables left in the common-pool resource
	CTK                                       float64                           // sunlight provisioned to the collective threshold kitty this round
	Patch                                     *patch.Patch                      // where the CPR's vegetables are on the grid; always sums to CPR after each step
	Weather                                   *weather.Weather                  // clouds over the patch, which shade agents from the sun
	positions                                 map[uuid.UUID]Position            // each agent's fixed cell, kept after it dies so the UI can still show it
	occupied                                  map[Position]bool                 // cells already given to an agent
	Recorder                                  *recorder.Recorder                // records every step for the UI; nil means recording is off
	names                                     map[uuid.UUID]string              // short agent labels (A1, A2, ...) used in recordings
	order                                     []uuid.UUID                       // agents in the order they were added, so recordings list them consistently
	dead                                      map[uuid.UUID]recorder.AgentState // final state of agents that died, kept in later frames
}

// Initialise returns a Server which adds the numAgents of BaseAmoeBoids, ready to Start.
func Initialise(numAgents, iterations, turns, messageBandwidth int, timeout time.Duration) *Server {
	s := &Server{
		BaseServer: baseserver.CreateBaseServer[objects.IAmoeBoid](iterations, turns, timeout, messageBandwidth),
		CPR:        initialCPR,
		Patch:      patch.New(GridWidth, GridHeight, capacityCPR/(GridWidth*GridHeight), initialCPR),
		Weather:    weather.NewRandom(GridWidth, GridHeight, numClouds),
		positions:  make(map[uuid.UUID]Position),
		occupied:   make(map[Position]bool),
		names:      make(map[uuid.UUID]string),
		dead:       make(map[uuid.UUID]recorder.AgentState),
	}
	s.SetGameRunner(s)
	for range numAgents {
		s.AddAgent(objects.CreateBaseAmoeBoid(s))
	}
	return s
}

// AddAgent registers the agent with the platform and gives it a fixed random cell no other agent has.
// If every cell is taken (more than GridWidth*GridHeight agents), it shares a random cell instead.
func (s *Server) AddAgent(ag objects.IAmoeBoid) {
	s.BaseServer.AddAgent(ag)
	s.nameAgent(ag.GetID())
	if _, placed := s.positions[ag.GetID()]; placed {
		return
	}
	var free []Position
	for y := range GridHeight {
		for x := range GridWidth {
			if p := (Position{x, y}); !s.occupied[p] {
				free = append(free, p)
			}
		}
	}
	p := Position{rand.IntN(GridWidth), rand.IntN(GridHeight)}
	if len(free) > 0 {
		p = free[rand.IntN(len(free))]
	}
	s.positions[ag.GetID()] = p
	s.occupied[p] = true
}

// GetPosition returns the agent's cell, and false if the agent was never added
func (s *Server) GetPosition(id uuid.UUID) (Position, bool) {
	p, ok := s.positions[id]
	return p, ok
}

// Each round the clouds drift, then every agent gets sunlight: 1.2x sunPerRound in clear sky down to 0.8x under the middle of a cloud
// RunStartOfIteration will also do the voluntary association and deciding of rules
func (s *Server) RunStartOfIteration(round int) {
	s.Weather.Step()
	for _, ag := range s.GetAgentMap() {
		pos, _ := s.GetPosition(ag.GetID())
		sun := sunPerRound * (1.2 - 0.4*s.Weather.Coverage(pos.X, pos.Y))
		ag.AddBrain(sun)
		s.recordEvent(recorder.EventSun, ag.GetID(), sun)
	}
	s.recordFrame(recorder.PhaseStart, round, -1, 0)
}

// One randomly selected agent appropriates each turn
func (s *Server) RunTurn(round, turn int) {
	s.appropriate()
	s.recordFrame(recorder.PhaseTurn, round, turn, 0) // recorded even when nobody could eat, so every turn has a frame
}

func (s *Server) appropriate() {
	agents := slices.Collect(maps.Values(s.GetAgentMap()))
	if s.CPR <= 0 || len(agents) == 0 {
		return
	}
	ag := agents[rand.IntN(len(agents))]
	taken := max(0, min(ag.DecideAppropriation(s.CPR), s.CPR))
	s.CPR -= taken
	s.Patch.SetTotal(s.CPR + taken) // catch up with any direct change to s.CPR so the bite comes out of the right tiles
	pos, _ := s.GetPosition(ag.GetID())
	s.Patch.Take(pos.X, pos.Y, taken)
	s.Patch.SetTotal(s.CPR)
	ag.Eat(taken)
	ag.SendAppropriationMessage(taken)
	s.recordEvent(recorder.EventAppropriate, ag.GetID(), taken)
}

// Provision, replenish, then remove any agents that died
// RunEndOfIteration will also disband
func (s *Server) RunEndOfIteration(round int) {
	for _, ag := range s.GetAgentMap() {
		given := ag.SpendBrain(ag.DecideProvision())
		s.CTK += given
		ag.SendProvisionMessage(given)
		s.recordEvent(recorder.EventProvision, ag.GetID(), given)
	}
	s.Patch.SetTotal(s.CPR)
	if s.CTK >= thresholdCTK {
		growth := growthRate * s.CPR * (1 - s.CPR/capacityCPR)
		s.CPR += growth
		s.Patch.Regrow(growth)
		s.Patch.SetTotal(s.CPR)
		s.recordEvent(recorder.EventRegrow, uuid.Nil, growth)
	}
	roundCTK := s.CTK
	s.CTK = 0
	for _, ag := range s.GetAgentMap() {
		if !ag.Metabolise(metabolism) {
			s.recordDeath(ag.GetID(), ag.GetBody(), ag.GetBrain())
			s.RemoveAgent(ag)
		}
	}
	s.Iterations++
	s.recordFrame(recorder.PhaseEnd, round, -1, roundCTK)
}
