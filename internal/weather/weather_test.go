package weather_test

import (
	"math"
	"testing"

	"github.com/GAMINpanda/somas/internal/weather"
)

const eps = 1e-9

func near(a, b float64) bool { return math.Abs(a-b) < eps }

func oneCloud(c weather.Cloud) *weather.Weather {
	return &weather.Weather{Width: 10, Height: 10, Clouds: []weather.Cloud{c}}
}

func TestCoverageFullAtCentre(t *testing.T) {
	w := oneCloud(weather.Cloud{X: 4.5, Y: 4.5, Radius: 2}) // centred on cell (4, 4)
	if got := w.Coverage(4, 4); !near(got, 1) {
		t.Error("coverage at the centre is", got, "expected 1")
	}
}

func TestCoverageFadesWithDistance(t *testing.T) {
	w := oneCloud(weather.Cloud{X: 4.5, Y: 4.5, Radius: 2})
	if got := w.Coverage(5, 4); !near(got, 0.5) { // 1 cell away out of a radius of 2
		t.Error("coverage one cell away is", got, "expected 0.5")
	}
	if got := w.Coverage(6, 4); got != 0 { // exactly at the radius
		t.Error("coverage at the radius is", got, "expected 0")
	}
	if got := w.Coverage(9, 9); got != 0 {
		t.Error("coverage far away is", got, "expected 0")
	}
}

func TestCoverageWrapsAcrossEdges(t *testing.T) {
	w := oneCloud(weather.Cloud{X: 0.5, Y: 0.5, Radius: 2}) // centred on corner cell (0, 0)
	if got := w.Coverage(9, 0); !near(got, 0.5) {
		t.Error("coverage on the far side of the left edge is", got, "expected 0.5")
	}
	if got := w.Coverage(0, 9); !near(got, 0.5) {
		t.Error("coverage on the far side of the top edge is", got, "expected 0.5")
	}
}

func TestOverlappingCloudsStayAtMostOne(t *testing.T) {
	w := &weather.Weather{Width: 10, Height: 10, Clouds: []weather.Cloud{
		{X: 4.5, Y: 4.5, Radius: 2},
		{X: 5.5, Y: 4.5, Radius: 2},
	}}
	if got := w.Coverage(4, 4); !near(got, 1) {
		t.Error("coverage under both clouds is", got, "expected 1")
	}
	// one cell from each centre: 1 - (1-0.5)(1-0.5) = 0.75
	w.Clouds[0].X, w.Clouds[1].X = 3.5, 5.5
	if got := w.Coverage(4, 4); !near(got, 0.75) {
		t.Error("coverage between two clouds is", got, "expected 0.75")
	}
}

func TestClearSkyHasNoCoverage(t *testing.T) {
	w := &weather.Weather{Width: 10, Height: 10}
	if got := w.Coverage(3, 3); got != 0 {
		t.Error("coverage with no clouds is", got, "expected 0")
	}
}

func TestStepMovesByVelocity(t *testing.T) {
	w := oneCloud(weather.Cloud{X: 2, Y: 3, Radius: 2, VX: 1.5, VY: -0.5})
	w.Step()
	if c := w.Clouds[0]; !near(c.X, 3.5) || !near(c.Y, 2.5) {
		t.Error("cloud moved to", c.X, c.Y, "expected 3.5, 2.5")
	}
}

func TestStepWrapsAround(t *testing.T) {
	w := oneCloud(weather.Cloud{X: 9.5, Y: 0.25, Radius: 2, VX: 1, VY: -0.5})
	w.Step()
	if c := w.Clouds[0]; !near(c.X, 0.5) || !near(c.Y, 9.75) {
		t.Error("cloud moved to", c.X, c.Y, "expected it to wrap to 0.5, 9.75")
	}
}

func TestNewRandomCloudsInRange(t *testing.T) {
	for range 100 {
		w := weather.NewRandom(10, 10, 4)
		if len(w.Clouds) != 4 {
			t.Fatal(len(w.Clouds), "clouds, expected 4")
		}
		for _, c := range w.Clouds {
			speed := math.Hypot(c.VX, c.VY)
			if c.X < 0 || c.X >= 10 || c.Y < 0 || c.Y >= 10 || c.Radius < 1.5 || c.Radius > 3 || speed < 0.5-eps || speed > 1.5+eps {
				t.Errorf("cloud out of range: %+v (speed %v)", c, speed)
			}
		}
	}
}

func TestCoverageAlwaysBetweenZeroAndOne(t *testing.T) {
	w := weather.NewRandom(10, 10, 4)
	for range 50 {
		w.Step()
		for y := range 10 {
			for x := range 10 {
				if got := w.Coverage(x, y); got < 0 || got > 1 {
					t.Fatalf("coverage at (%d, %d) is %v, expected 0 to 1", x, y, got)
				}
			}
		}
	}
}
