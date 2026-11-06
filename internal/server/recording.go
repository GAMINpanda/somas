package server

import (
	"fmt"
	"slices"

	"github.com/GAMINpanda/somas/internal/recorder"
	"github.com/google/uuid"
)

// EnableRecording makes the server record a frame at every step from now on, for the UI to replay.
// rounds and turns are only stored in the run's settings; they don't change how long the game runs.
func (s *Server) EnableRecording(rounds, turns int) *recorder.Recorder {
	s.Recorder = recorder.New(recorder.Settings{
		Agents:       len(s.GetAgentMap()),
		Rounds:       rounds,
		Turns:        turns,
		GridWidth:    GridWidth,
		GridHeight:   GridHeight,
		TileCapacity: s.Patch.TileCapacity,
		Capacity:     capacityCPR,
		Threshold:    thresholdCTK,
	})
	return s.Recorder
}

// nameAgent gives a newly added agent a short label (A1, A2, ...) in the order agents are added
func (s *Server) nameAgent(id uuid.UUID) {
	if _, named := s.names[id]; named {
		return
	}
	s.order = append(s.order, id)
	s.names[id] = fmt.Sprintf("A%d", len(s.order))
}

// recordEvent notes an event for the next frame. It does nothing when recording is off.
func (s *Server) recordEvent(kind string, id uuid.UUID, amount float64) {
	if s.Recorder == nil {
		return
	}
	agent := ""
	if id != uuid.Nil {
		agent = id.String()
	}
	s.Recorder.Event(kind, agent, amount)
}

// recordDeath keeps the agent's final state so it stays in later frames marked as dead
func (s *Server) recordDeath(id uuid.UUID, body, brain float64) {
	if s.Recorder == nil {
		return
	}
	state := s.agentState(id)
	state.Body, state.Brain, state.Alive = body, brain, false
	s.dead[id] = state
	s.recordEvent(recorder.EventDeath, id, 0)
}

// recordFrame stores the current state. ctk is passed in because the end of a round resets s.CTK before the frame is taken.
func (s *Server) recordFrame(phase string, round, turn int, ctk float64) {
	if s.Recorder == nil {
		return
	}
	agents := make([]recorder.AgentState, 0, len(s.order))
	for _, id := range s.order {
		if state, dead := s.dead[id]; dead {
			agents = append(agents, state)
		} else if _, alive := s.GetAgentMap()[id]; alive {
			agents = append(agents, s.agentState(id))
		}
	}
	s.Recorder.Frame(recorder.Frame{
		Round:     round,
		Turn:      turn,
		Phase:     phase,
		CPR:       s.CPR,
		CTK:       ctk,
		Threshold: thresholdCTK,
		Tiles:     s.Patch.Grid(),
		Clouds:    slices.Clone(s.Weather.Clouds),
		Agents:    agents,
	})
}

// agentState reads a living agent's current state
func (s *Server) agentState(id uuid.UUID) recorder.AgentState {
	pos, _ := s.GetPosition(id)
	state := recorder.AgentState{ID: id.String(), Name: s.names[id], X: pos.X, Y: pos.Y, Alive: true}
	if ag, ok := s.GetAgentMap()[id]; ok {
		state.Body, state.Brain = ag.GetBody(), ag.GetBrain()
	}
	return state
}
