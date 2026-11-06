package server_test

import (
	"math"
	"testing"
	"time"

	"github.com/GAMINpanda/somas/internal/common/objects"
	"github.com/GAMINpanda/somas/internal/server"
	"github.com/GAMINpanda/somas/internal/weather"
)

func TestInitialise(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, time.Millisecond) //10 agents, 5 iterations, 10 turns, a message bw of 10 and a 1ms timeout
	if len(s.GetAgentMap()) != 10 {
		t.Error(len(s.GetAgentMap()), "agents initialised, expected 10")
	}
}

func TestRunGame(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, time.Microsecond) //test with a us turn timeout
	s.Start()
	if s.Iterations != 5 {
		t.Error(s.Iterations, "cycles completed, expected 5")
	}
}

func TestMessaging(t *testing.T) { // Followed how the library tests sync msg broadcasts
	s := server.Initialise(10, 5, 10, 10, 0)
	sender := objects.CreateBaseAmoeBoid(s)
	s.AddAgent(sender)
	senderID := sender.GetID()

	sender.SendProvisionMessage(2.5)
	sender.SendAppropriationMessage(1.5)
	for _, agent := range s.GetAgentMap() {
		provision := agent.GetProvisionMap()[senderID]
		appropriation := agent.GetAppropriationMap()[senderID]

		if agent.GetID() == senderID && (provision != 0 || appropriation != 0) {
			t.Error(agent.GetID(), "is sender and received its own message")
		} else if agent.GetID() != senderID && (provision != 2.5 || appropriation != 1.5) {
			t.Error(agent.GetID(), "didn't receive both prov and approp messages")
		}
	}
}

// scriptedAmoeBoid replaces the random strategy with fixed choices so results can be checked exactly
type scriptedAmoeBoid struct {
	*objects.BaseAmoeBoid
	appropriation float64
	provision     float64
}

func (a *scriptedAmoeBoid) DecideAppropriation(float64) float64 { return a.appropriation }
func (a *scriptedAmoeBoid) DecideProvision() float64            { return a.provision }

func addScriptedAgent(s *server.Server, appropriation, provision, body, brain float64) *scriptedAmoeBoid {
	a := &scriptedAmoeBoid{objects.CreateBaseAmoeBoid(s), appropriation, provision}
	a.Eat(body)
	a.AddBrain(brain)
	s.AddAgent(a)
	return a
}

func TestInitialPools(t *testing.T) {
	s := server.Initialise(10, 5, 10, 10, 0)
	if s.CPR != 500 || s.CTK != 0 {
		t.Error("CPR is", s.CPR, "and CTK is", s.CTK, "expected 500 and 0")
	}
}

func TestAppropriationLimitedToCPR(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := addScriptedAgent(s, 50, 0, 0, 1000)
	b := addScriptedAgent(s, 50, 0, 0, 1000)
	s.CPR = 10

	s.RunTurn(0, 0)

	if s.CPR != 0 {
		t.Error("CPR is", s.CPR, "expected 0")
	}
	taker, observer := a, b
	if b.GetUtility() > 0 {
		taker, observer = b, a
	}
	if taker.GetUtility() != 10 {
		t.Error("agent ate", taker.GetUtility(), "expected 10")
	}
	if observer.GetUtility() != 0 {
		t.Error("two agents appropriated in one turn")
	}
	if observer.GetAppropriationMap()[taker.GetID()] != 10 {
		t.Error("broadcast", observer.GetAppropriationMap()[taker.GetID()], "expected the 10 actually taken")
	}
}

func TestNegativeAppropriationIgnored(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := addScriptedAgent(s, -20, 0, 0, 1000)

	s.RunTurn(0, 0)

	if s.CPR != 500 || a.GetUtility() != 0 {
		t.Error("CPR is", s.CPR, "and agent ate", a.GetUtility(), "expected 500 and 0")
	}
}

func TestNoAppropriationFromEmptyCPR(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := addScriptedAgent(s, 50, 0, 0, 1000)
	b := addScriptedAgent(s, 50, 0, 0, 1000)
	s.CPR = 0

	s.RunTurn(0, 0)

	if a.GetUtility() != 0 || b.GetUtility() != 0 {
		t.Error("agent ate from an empty CPR")
	}
	if len(a.GetAppropriationMap()) != 0 || len(b.GetAppropriationMap()) != 0 {
		t.Error("appropriation was broadcast from an empty CPR")
	}
}

func TestRunTurnWithNoAgents(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	s.RunTurn(0, 0) // would panic if it tried to select an agent
	if s.CPR != 500 {
		t.Error("CPR is", s.CPR, "expected 500")
	}
}

func TestSunlightPerRound(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	agents := make([]*scriptedAmoeBoid, 50)
	for i := range agents {
		agents[i] = addScriptedAgent(s, 0, 0, 1000, 0) // lots of body so utility equals brain
	}

	s.RunStartOfIteration(0)

	for _, a := range agents {
		if sun := a.GetUtility(); sun < 8 || sun > 12 {
			t.Error("agent received", sun, "sunlight, expected 8 to 12")
		}
	}
}

func TestSpendBrainLimits(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := objects.CreateBaseAmoeBoid(s)
	a.Eat(1000)
	a.AddBrain(10)

	if spent := a.SpendBrain(-5); spent != 0 {
		t.Error("spent", spent, "expected 0 for a negative amount")
	}
	if spent := a.SpendBrain(4); spent != 4 {
		t.Error("spent", spent, "expected 4")
	}
	if spent := a.SpendBrain(100); spent != 6 {
		t.Error("spent", spent, "expected only the remaining 6")
	}
	if a.GetUtility() != 0 {
		t.Error("brain is", a.GetUtility(), "expected 0")
	}
}

func TestMetabolise(t *testing.T) {
	tests := []struct {
		name      string
		body      float64
		brain     float64
		wantAlive bool
	}{
		{"enough of both", 10, 10, true},
		{"body reaches zero", 5, 10, false},
		{"brain reaches zero", 10, 5, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.Initialise(0, 1, 1, 10, 0)
			a := objects.CreateBaseAmoeBoid(s)
			a.Eat(tt.body)
			a.AddBrain(tt.brain)
			if alive := a.Metabolise(5); alive != tt.wantAlive {
				t.Error("alive is", alive, "expected", tt.wantAlive)
			}
		})
	}
}

func TestReplenishment(t *testing.T) { // Threshold T is 30, K is 1000 and r is 0.5
	tests := []struct {
		name      string
		cpr       float64
		provision float64
		wantCPR   float64
	}{
		{"below threshold", 500, 29, 500},
		{"at threshold", 500, 30, 625},
		{"above threshold gives no extra growth", 500, 100, 625},
		{"empty patch stays empty", 0, 100, 0},
		{"full patch stays full", 1000, 100, 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.Initialise(0, 1, 1, 10, 0)
			addScriptedAgent(s, 0, tt.provision, 200, 200)
			s.CPR = tt.cpr

			s.RunEndOfIteration(0)

			if s.CPR != tt.wantCPR {
				t.Error("CPR is", s.CPR, "expected", tt.wantCPR)
			}
		})
	}
}

func TestProvisionLimitedToBrain(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	giver := addScriptedAgent(s, 0, 100, 200, 20)
	observer := addScriptedAgent(s, 0, 0, 200, 200)

	s.RunEndOfIteration(0)

	if s.CPR != 500 {
		t.Error("CPR is", s.CPR, "expected no regrowth because only 20 could be provisioned")
	}
	if observer.GetProvisionMap()[giver.GetID()] != 20 {
		t.Error("broadcast", observer.GetProvisionMap()[giver.GetID()], "expected the 20 actually provisioned")
	}
}

func TestCTKResetsEachRound(t *testing.T) {
	s := server.Initialise(0, 2, 1, 10, 0)
	addScriptedAgent(s, 0, 20, 200, 200)

	s.RunEndOfIteration(0)
	if s.CTK != 0 {
		t.Error("CTK is", s.CTK, "expected 0 after the round")
	}

	s.RunEndOfIteration(1)
	if s.CPR != 500 {
		t.Error("CPR is", s.CPR, "expected no regrowth because 20 per round must not accumulate")
	}
}

func TestDeadAgentsRemoved(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	starving := addScriptedAgent(s, 0, 0, 3, 200)
	healthy := addScriptedAgent(s, 0, 0, 200, 200)

	s.RunEndOfIteration(0)

	if _, found := s.GetAgentMap()[starving.GetID()]; found {
		t.Error("starving agent is still in the agent map")
	}
	if _, found := s.ViewAgentIdSet()[starving.GetID()]; found {
		t.Error("starving agent is still in the agent ID set")
	}
	if _, found := s.GetAgentMap()[healthy.GetID()]; !found {
		t.Error("healthy agent was removed")
	}
	if s.Iterations != 1 {
		t.Error(s.Iterations, "cycles completed, expected 1")
	}
}

func TestCPROutcomeDependsOnAppropriation(t *testing.T) { // The core dilemma: restraint sustains the patch and greed destroys it
	tests := []struct {
		name          string
		appropriation float64
		sustained     bool
	}{
		{"restrained", 5, true},
		{"greedy", 50, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := server.Initialise(0, 20, 10, 10, 0)
			for range 10 {
				addScriptedAgent(s, tt.appropriation, 10, 100000, 100000)
			}

			s.Start()

			if tt.sustained && s.CPR <= 500 {
				t.Error("CPR is", s.CPR, "expected it to grow above 500")
			}
			if !tt.sustained && s.CPR != 0 {
				t.Error("CPR is", s.CPR, "expected the patch to be destroyed")
			}
		})
	}
}

func TestFullGameKeepsPoolsValid(t *testing.T) { // Random agents, including the case where all of them die
	s := server.Initialise(10, 200, 10, 10, 0)
	s.Start()
	if s.Iterations != 200 {
		t.Error(s.Iterations, "cycles completed, expected 200")
	}
	if s.CPR < 0 || s.CPR > 1000 {
		t.Error("CPR is", s.CPR, "expected 0 to 1000")
	}
	if s.CTK != 0 {
		t.Error("CTK is", s.CTK, "expected 0 after the final round")
	}
}

func TestGetBodyAndBrain(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := objects.CreateBaseAmoeBoid(s)
	a.Eat(7)
	a.AddBrain(3)
	if a.GetBody() != 7 || a.GetBrain() != 3 {
		t.Error("body is", a.GetBody(), "and brain is", a.GetBrain(), "expected 7 and 3")
	}
}

func TestPositionsDistinctAndInBounds(t *testing.T) {
	numAgents := server.GridWidth * server.GridHeight // fill every cell so a clash would be certain to show up
	s := server.Initialise(numAgents, 1, 1, 10, 0)
	seen := make(map[server.Position]bool)
	for id := range s.GetAgentMap() {
		p, ok := s.GetPosition(id)
		if !ok {
			t.Fatal(id, "has no position")
		}
		if p.X < 0 || p.X >= server.GridWidth || p.Y < 0 || p.Y >= server.GridHeight {
			t.Error(id, "is at", p, "which is outside the grid")
		}
		if seen[p] {
			t.Error("two agents share cell", p)
		}
		seen[p] = true
	}
}

func TestPositionGivenToAgentsAddedLater(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := addScriptedAgent(s, 0, 0, 0, 0) // added directly, not through Initialise
	if _, ok := s.GetPosition(a.GetID()); !ok {
		t.Error("agent added with AddAgent has no position")
	}
}

func TestPositionFixedAfterReAdd(t *testing.T) {
	s := server.Initialise(1, 1, 1, 10, 0)
	for id, ag := range s.GetAgentMap() {
		before, _ := s.GetPosition(id)
		s.AddAgent(ag)
		if after, _ := s.GetPosition(id); after != before {
			t.Error("agent moved from", before, "to", after, "when added again")
		}
	}
}

func TestPositionKeptAfterDeath(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	starving := addScriptedAgent(s, 0, 0, 3, 200)
	before, _ := s.GetPosition(starving.GetID())

	s.RunEndOfIteration(0)

	if after, ok := s.GetPosition(starving.GetID()); !ok || after != before {
		t.Error("dead agent's position is", after, ok, "expected", before)
	}
}

func TestMoreAgentsThanCells(t *testing.T) {
	numAgents := server.GridWidth*server.GridHeight + 5
	s := server.Initialise(numAgents, 1, 1, 10, 0)
	for id := range s.GetAgentMap() {
		if p, ok := s.GetPosition(id); !ok || p.X < 0 || p.X >= server.GridWidth || p.Y < 0 || p.Y >= server.GridHeight {
			t.Error(id, "has position", p, ok, "expected a cell inside the grid")
		}
	}
}

func patchMatchesCPR(t *testing.T, s *server.Server) {
	t.Helper()
	if diff := s.Patch.Total() - s.CPR; diff > 1e-6 || diff < -1e-6 {
		t.Error("patch holds", s.Patch.Total(), "but CPR is", s.CPR)
	}
}

func TestPatchStartsAtCPR(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	patchMatchesCPR(t, s)
	if s.Patch.Width != server.GridWidth || s.Patch.Height != server.GridHeight || s.Patch.Capacity() != 1000 {
		t.Error("patch is", s.Patch.Width, "x", s.Patch.Height, "holding", s.Patch.Capacity(), "expected 10 x 10 holding 1000")
	}
}

func TestTurnEatsFromAgentsOwnTile(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := addScriptedAgent(s, 3, 0, 0, 1000) // a bite smaller than one tile's 5 vegetables
	pos, _ := s.GetPosition(a.GetID())

	s.RunTurn(0, 0)

	if got := s.Patch.Tile(pos.X, pos.Y); got < 1.99 || got > 2.01 {
		t.Error("agent's tile holds", got, "expected 5 - 3 = 2")
	}
	patchMatchesCPR(t, s)
}

func TestPatchFollowsDirectCPRChanges(t *testing.T) { // tests and future rules may set s.CPR directly
	s := server.Initialise(0, 1, 1, 10, 0)
	addScriptedAgent(s, 5, 0, 0, 1000)
	s.CPR = 10

	s.RunTurn(0, 0)

	patchMatchesCPR(t, s)
}

func TestRegrowthReachesPatch(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	addScriptedAgent(s, 0, 30, 200, 200)

	s.RunEndOfIteration(0)

	if s.CPR != 625 {
		t.Error("CPR is", s.CPR, "expected 625")
	}
	patchMatchesCPR(t, s)
}

func TestPatchMatchesCPRThroughGame(t *testing.T) {
	s := server.Initialise(10, 50, 10, 10, 0)
	s.Start()
	patchMatchesCPR(t, s)
}

func TestSunlightFollowsCloudCover(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	shaded := addScriptedAgent(s, 0, 0, 1000, 0)
	sunny := addScriptedAgent(s, 0, 0, 1000, 0)
	p, _ := s.GetPosition(shaded.GetID())
	// one still cloud centred on the shaded agent, small enough to miss every other cell
	s.Weather.Clouds = []weather.Cloud{{X: float64(p.X) + 0.5, Y: float64(p.Y) + 0.5, Radius: 0.5}}

	s.RunStartOfIteration(0)

	if got := shaded.GetBrain(); math.Abs(got-8) > 1e-9 {
		t.Error("agent under the cloud got", got, "sunlight, expected 8")
	}
	if got := sunny.GetBrain(); math.Abs(got-12) > 1e-9 {
		t.Error("agent in clear sky got", got, "sunlight, expected 12")
	}
}

func TestCloudsDriftEachRound(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	s.Weather.Clouds = []weather.Cloud{{X: 2, Y: 2, Radius: 2, VX: 1, VY: 0}}

	s.RunStartOfIteration(0)
	s.RunStartOfIteration(1)

	if c := s.Weather.Clouds[0]; math.Abs(c.X-4) > 1e-9 || c.Y != 2 {
		t.Error("cloud is at", c.X, c.Y, "after two rounds, expected 4, 2")
	}
}
