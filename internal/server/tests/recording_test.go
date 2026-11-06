package server_test

import (
	"math"
	"testing"

	"github.com/GAMINpanda/somas/internal/recorder"
	"github.com/GAMINpanda/somas/internal/server"
)

func TestRecordsEveryStep(t *testing.T) {
	const rounds, turns = 3, 4
	s := server.Initialise(5, rounds, turns, 10, 0)
	rec := s.EnableRecording(rounds, turns)
	s.Start()

	frames := rec.Run().Frames
	if len(frames) != rounds*(turns+2) {
		t.Fatal(len(frames), "frames, expected", rounds*(turns+2))
	}
	i := 0
	for round := range rounds {
		want := []struct {
			phase string
			turn  int
		}{{recorder.PhaseStart, -1}}
		for turn := range turns {
			want = append(want, struct {
				phase string
				turn  int
			}{recorder.PhaseTurn, turn})
		}
		want = append(want, struct {
			phase string
			turn  int
		}{recorder.PhaseEnd, -1})
		for _, w := range want {
			f := frames[i]
			if f.Round != round || f.Turn != w.turn || f.Phase != w.phase {
				t.Errorf("frame %d is round %d turn %d %s, expected round %d turn %d %s", i, f.Round, f.Turn, f.Phase, round, w.turn, w.phase)
			}
			i++
		}
	}
	if s := rec.Run().Settings; s.Agents != 5 || s.Rounds != rounds || s.Turns != turns || s.GridWidth != server.GridWidth || s.Threshold != 30 || s.Capacity != 1000 {
		t.Errorf("settings are %+v", s)
	}
}

func TestFramesMatchGameState(t *testing.T) {
	s := server.Initialise(10, 20, 10, 10, 0)
	rec := s.EnableRecording(20, 10)
	s.Start()

	for i, f := range rec.Run().Frames {
		total := 0.0
		for _, row := range f.Tiles {
			total += sum(row)
		}
		if math.Abs(total-f.CPR) > 1e-6 {
			t.Errorf("frame %d tiles hold %v but CPR is %v", i, total, f.CPR)
		}
		if len(f.Tiles) != server.GridHeight || len(f.Tiles[0]) != server.GridWidth || len(f.Clouds) == 0 {
			t.Errorf("frame %d has a %dx%d grid and %d clouds", i, len(f.Tiles[0]), len(f.Tiles), len(f.Clouds))
		}
		if len(f.Agents) != 10 {
			t.Errorf("frame %d lists %d agents, expected all 10 (dead ones included)", i, len(f.Agents))
		}
		if f.Phase == recorder.PhaseEnd {
			provided := 0.0
			for _, e := range f.Events {
				if e.Kind == recorder.EventProvision {
					provided += e.Amount
				}
			}
			if math.Abs(provided-f.CTK) > 1e-9 {
				t.Errorf("frame %d CTK is %v but provisions add up to %v", i, f.CTK, provided)
			}
		}
	}
}

func sum(row []float64) float64 {
	total := 0.0
	for _, v := range row {
		total += v
	}
	return total
}

func TestAppropriationAndSunEvents(t *testing.T) {
	s := server.Initialise(0, 1, 1, 10, 0)
	a := addScriptedAgent(s, 12, 0, 100, 100)
	rec := s.EnableRecording(1, 1)

	s.RunStartOfIteration(0)
	s.RunTurn(0, 0)

	frames := rec.Run().Frames
	start, turn := frames[0], frames[1]
	if len(start.Events) != 1 || start.Events[0].Kind != recorder.EventSun || start.Events[0].Agent != a.GetID().String() {
		t.Error("start events are", start.Events, "expected one sun event for the agent")
	}
	if start.Agents[0].Brain != 100+start.Events[0].Amount {
		t.Error("recorded brain is", start.Agents[0].Brain, "expected 100 plus the sunlight")
	}
	if len(turn.Events) != 1 || turn.Events[0].Kind != recorder.EventAppropriate || turn.Events[0].Amount != 12 {
		t.Error("turn events are", turn.Events, "expected one appropriation of 12")
	}
	if turn.CPR != 488 || turn.Agents[0].Body != 112 {
		t.Error("turn frame has CPR", turn.CPR, "and body", turn.Agents[0].Body, "expected 488 and 112")
	}
	pos, _ := s.GetPosition(a.GetID())
	if ag := turn.Agents[0]; ag.X != pos.X || ag.Y != pos.Y || ag.Name != "A1" || !ag.Alive {
		t.Errorf("recorded agent is %+v, expected A1 alive at %v", ag, pos)
	}
}

func TestRegrowEventOnlyWhenThresholdMet(t *testing.T) {
	for _, tt := range []struct {
		provision float64
		regrows   bool
	}{{29, false}, {30, true}} {
		s := server.Initialise(0, 1, 1, 10, 0)
		addScriptedAgent(s, 0, tt.provision, 200, 200)
		rec := s.EnableRecording(1, 1)

		s.RunEndOfIteration(0)

		end := rec.Run().Frames[0]
		var regrow *recorder.Event
		for i, e := range end.Events {
			if e.Kind == recorder.EventRegrow {
				regrow = &end.Events[i]
			}
		}
		if tt.regrows && (regrow == nil || regrow.Amount != 125 || regrow.Agent != "") {
			t.Error("provision", tt.provision, "regrow event is", regrow, "expected 125 with no agent")
		}
		if !tt.regrows && regrow != nil {
			t.Error("provision", tt.provision, "recorded regrowth below the threshold")
		}
		if end.CTK != tt.provision {
			t.Error("end frame CTK is", end.CTK, "expected", tt.provision, "(before the kitty resets)")
		}
	}
}

func TestDeadAgentsStayInLaterFrames(t *testing.T) {
	s := server.Initialise(0, 2, 1, 10, 0)
	starving := addScriptedAgent(s, 0, 0, 3, 200)
	addScriptedAgent(s, 0, 0, 200, 200)
	rec := s.EnableRecording(2, 1)

	s.RunEndOfIteration(0)
	s.RunStartOfIteration(1)
	s.RunTurn(1, 0)

	frames := rec.Run().Frames
	deaths := 0
	for _, e := range frames[0].Events {
		if e.Kind == recorder.EventDeath && e.Agent == starving.GetID().String() {
			deaths++
		}
	}
	if deaths != 1 {
		t.Error("end frame has", deaths, "death events for the starving agent, expected 1")
	}
	for i, f := range frames {
		found := false
		for _, ag := range f.Agents {
			if ag.ID == starving.GetID().String() {
				found = true
				if ag.Alive || ag.Body != -2 {
					t.Errorf("frame %d shows the dead agent as %+v, expected alive=false with its final body of -2", i, ag)
				}
			}
		}
		if !found {
			t.Error("frame", i, "is missing the dead agent")
		}
	}
	for _, e := range frames[1].Events {
		if e.Agent == starving.GetID().String() {
			t.Error("dead agent got an event after dying:", e)
		}
	}
}

func TestNoRecordingByDefault(t *testing.T) {
	s := server.Initialise(5, 2, 2, 10, 0)
	s.Start()
	if s.Recorder != nil {
		t.Error("server recorded without EnableRecording")
	}
}
