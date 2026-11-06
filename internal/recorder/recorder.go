// Package recorder collects a frame of game state at each step of a run and writes the whole run as JSON for the UI to replay.
package recorder

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/GAMINpanda/somas/internal/weather"
)

// Phases a frame can be recorded in
const (
	PhaseStart = "start" // after sunlight is given out at the start of a round
	PhaseTurn  = "turn"  // after one agent has appropriated
	PhaseEnd   = "end"   // after provision, regrowth and deaths at the end of a round
)

// Event kinds
const (
	EventSun         = "sun"         // agent received sunlight
	EventAppropriate = "appropriate" // agent took vegetables from the CPR
	EventProvision   = "provision"   // agent gave sunlight to the CTK
	EventRegrow      = "regrow"      // the CPR regrew (no agent)
	EventDeath       = "death"       // agent died and was removed
)

// Settings describes the run, so the UI can size the grid and label the threshold
type Settings struct {
	Agents       int     `json:"agents"`
	Rounds       int     `json:"rounds"`
	Turns        int     `json:"turns"`
	GridWidth    int     `json:"gridWidth"`
	GridHeight   int     `json:"gridHeight"`
	TileCapacity float64 `json:"tileCapacity"`
	Capacity     float64 `json:"capacity"`
	Threshold    float64 `json:"threshold"`
}

// AgentState is one agent's state in a frame
type AgentState struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"` // short label such as "A3", numbered in the order agents were added
	X     int     `json:"x"`
	Y     int     `json:"y"`
	Body  float64 `json:"body"`
	Brain float64 `json:"brain"`
	Alive bool    `json:"alive"`
}

// Event is something that happened during a frame's step. Agent is empty for events with no agent, such as regrowth.
type Event struct {
	Kind   string  `json:"kind"`
	Agent  string  `json:"agent,omitempty"`
	Amount float64 `json:"amount"`
}

// Frame is the game state after one step, with the events that happened in that step
type Frame struct {
	Round     int             `json:"round"`
	Turn      int             `json:"turn"` // -1 for start and end frames
	Phase     string          `json:"phase"`
	CPR       float64         `json:"cpr"`
	CTK       float64         `json:"ctk"` // in an end frame, the total provisioned this round, before the kitty resets
	Threshold float64         `json:"threshold"`
	Tiles     [][]float64     `json:"tiles"` // tiles[y][x]
	Clouds    []weather.Cloud `json:"clouds"`
	Agents    []AgentState    `json:"agents"`
	Events    []Event         `json:"events"`
}

// Run is everything written to the JSON file
type Run struct {
	Settings Settings `json:"settings"`
	Frames   []Frame  `json:"frames"`
}

// Recorder builds up a Run. Events are collected with Event and attached to the next frame passed to Frame.
type Recorder struct {
	run     Run
	pending []Event
}

// New returns an empty recorder for a run with the given settings
func New(settings Settings) *Recorder {
	return &Recorder{run: Run{Settings: settings, Frames: []Frame{}}}
}

// Event notes something that happened in the current step
func (r *Recorder) Event(kind, agent string, amount float64) {
	r.pending = append(r.pending, Event{Kind: kind, Agent: agent, Amount: amount})
}

// Frame stores the frame together with every event noted since the last frame
func (r *Recorder) Frame(f Frame) {
	f.Events = r.pending
	if f.Events == nil {
		f.Events = []Event{}
	}
	r.pending = nil
	r.run.Frames = append(r.run.Frames, f)
}

// Run returns the run recorded so far
func (r *Recorder) Run() Run {
	return r.run
}

// JSON returns the run encoded as JSON, as written by Write and served to the UI
func (r *Recorder) JSON() ([]byte, error) {
	return json.Marshal(r.run)
}

// Write saves the run as JSON, creating the file's folder if needed
func (r *Recorder) Write(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := r.JSON()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
