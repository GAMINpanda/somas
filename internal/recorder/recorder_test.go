package recorder_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/GAMINpanda/somas/internal/recorder"
)

func TestEventsAttachToNextFrameOnly(t *testing.T) {
	r := recorder.New(recorder.Settings{})
	r.Event(recorder.EventSun, "a", 10)
	r.Event(recorder.EventSun, "b", 11)
	r.Frame(recorder.Frame{Phase: recorder.PhaseStart})
	r.Frame(recorder.Frame{Phase: recorder.PhaseTurn})

	frames := r.Run().Frames
	if len(frames) != 2 {
		t.Fatal(len(frames), "frames, expected 2")
	}
	if len(frames[0].Events) != 2 || frames[0].Events[1].Agent != "b" || frames[0].Events[1].Amount != 11 {
		t.Error("first frame events are", frames[0].Events, "expected the two sun events")
	}
	if frames[1].Events == nil || len(frames[1].Events) != 0 {
		t.Error("second frame events are", frames[1].Events, "expected an empty list")
	}
}

func TestWriteCreatesFolderAndValidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "run.json")
	r := recorder.New(recorder.Settings{Agents: 2, Rounds: 1, Turns: 1, GridWidth: 3, GridHeight: 3, Threshold: 30})
	r.Event(recorder.EventRegrow, "", 5)
	r.Frame(recorder.Frame{Round: 0, Turn: -1, Phase: recorder.PhaseEnd, CPR: 7, Tiles: [][]float64{{1, 2, 4}}})

	if err := r.Write(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var run recorder.Run
	if err := json.Unmarshal(data, &run); err != nil {
		t.Fatal("not valid JSON:", err)
	}
	if run.Settings.Threshold != 30 || len(run.Frames) != 1 || run.Frames[0].CPR != 7 || run.Frames[0].Events[0].Kind != "regrow" {
		t.Errorf("read back %+v", run)
	}
	// the UI reads these exact names
	for _, key := range []string{`"settings"`, `"frames"`, `"gridWidth"`, `"phase"`, `"tiles"`, `"clouds"`, `"agents"`, `"events"`, `"threshold"`} {
		if !strings.Contains(string(data), key) {
			t.Error("JSON is missing", key)
		}
	}
	if strings.Contains(string(data), `"agent":""`) {
		t.Error("events with no agent should leave the agent field out")
	}
}
