package patch_test

import (
	"math"
	"testing"

	"github.com/GAMINpanda/somas/internal/patch"
)

const eps = 1e-9

func close(a, b float64) bool { return math.Abs(a-b) < eps }

func checkTilesValid(t *testing.T, p *patch.Patch) {
	t.Helper()
	for y := range p.Height {
		for x := range p.Width {
			if v := p.Tile(x, y); v < -eps || v > p.TileCapacity+eps {
				t.Errorf("tile (%d, %d) holds %v, expected 0 to %v", x, y, v, p.TileCapacity)
			}
		}
	}
}

func TestNewSpreadsEvenly(t *testing.T) {
	p := patch.New(10, 10, 10, 500)
	for y := range 10 {
		for x := range 10 {
			if !close(p.Tile(x, y), 5) {
				t.Fatalf("tile (%d, %d) holds %v, expected 5", x, y, p.Tile(x, y))
			}
		}
	}
	if !close(p.Total(), 500) || p.Capacity() != 1000 {
		t.Error("total is", p.Total(), "and capacity is", p.Capacity(), "expected 500 and 1000")
	}
}

func TestNewClampsToCapacity(t *testing.T) {
	if p := patch.New(2, 2, 10, 999); !close(p.Total(), 40) {
		t.Error("total is", p.Total(), "expected it capped at 40")
	}
	if p := patch.New(2, 2, 10, -5); p.Total() != 0 {
		t.Error("total is", p.Total(), "expected 0")
	}
}

func TestTakeFromOwnTileFirst(t *testing.T) {
	p := patch.New(5, 5, 10, 250) // every tile full
	taken := p.Take(2, 2, 4)
	if !close(taken, 4) || !close(p.Tile(2, 2), 6) {
		t.Error("took", taken, "and own tile holds", p.Tile(2, 2), "expected 4 and 6")
	}
	if !close(p.Total(), 246) {
		t.Error("total is", p.Total(), "expected 246")
	}
}

func TestTakeSpreadsToNearestTiles(t *testing.T) {
	p := patch.New(5, 5, 10, 250)
	p.Take(2, 2, 50) // own tile (10) then the 4 side neighbours (40) are emptied before the diagonals are touched
	for _, c := range [][2]int{{2, 2}, {1, 2}, {3, 2}, {2, 1}, {2, 3}} {
		if p.Tile(c[0], c[1]) != 0 {
			t.Errorf("tile %v holds %v, expected it eaten first", c, p.Tile(c[0], c[1]))
		}
	}
	for _, c := range [][2]int{{1, 1}, {3, 3}, {0, 2}, {4, 4}} {
		if p.Tile(c[0], c[1]) != 10 {
			t.Errorf("tile %v holds %v, expected it untouched", c, p.Tile(c[0], c[1]))
		}
	}
}

func TestTakeFromCorner(t *testing.T) {
	p := patch.New(3, 3, 10, 90)
	p.Take(0, 0, 35) // empties (0,0) then (1,0) and (0,1), then takes 5 from the next nearest (1,1)
	if p.Tile(0, 0) != 0 || p.Tile(1, 0) != 0 || p.Tile(0, 1) != 0 || !close(p.Tile(1, 1), 5) {
		t.Error("corner bite went to the wrong tiles:", p.Grid())
	}
}

func TestTakeMoreThanPatchHolds(t *testing.T) {
	p := patch.New(3, 3, 10, 30)
	if taken := p.Take(1, 1, 100); !close(taken, 30) || p.Total() != 0 {
		t.Error("took", taken, "leaving", p.Total(), "expected 30 leaving 0")
	}
	checkTilesValid(t, p)
}

func TestTakeNothing(t *testing.T) {
	p := patch.New(3, 3, 10, 30)
	if taken := p.Take(1, 1, 0); taken != 0 || !close(p.Total(), 30) {
		t.Error("took", taken, "expected nothing")
	}
}

func TestRegrowInProportionToRoom(t *testing.T) {
	p := patch.New(2, 1, 10, 20) // both tiles full
	p.Take(0, 0, 10)             // tile (0,0) empty: room 10, tile (1,0) full: room 0
	p.Regrow(5)
	if !close(p.Tile(0, 0), 5) || !close(p.Tile(1, 0), 10) {
		t.Error("tiles are", p.Grid(), "expected all growth on the empty tile")
	}
}

func TestRegrowSplitsByRoom(t *testing.T) {
	p := patch.New(2, 1, 10, 20)
	p.Take(0, 0, 10) // (0,0) has room 10
	p.Take(1, 0, 5)  // (1,0) has room 5
	added := p.Regrow(6)
	if !close(added, 6) || !close(p.Tile(0, 0), 4) || !close(p.Tile(1, 0), 7) {
		t.Error("added", added, "tiles are", p.Grid(), "expected 6 split 4:2")
	}
}

func TestRegrowCapsAtCapacity(t *testing.T) {
	p := patch.New(3, 3, 10, 80)
	if added := p.Regrow(50); !close(added, 10) || !close(p.Total(), 90) {
		t.Error("added", added, "total", p.Total(), "expected 10 and 90")
	}
	if added := p.Regrow(5); added != 0 {
		t.Error("full patch regrew", added)
	}
	checkTilesValid(t, p)
}

func TestRegrowIgnoresNegative(t *testing.T) {
	p := patch.New(3, 3, 10, 30)
	if added := p.Regrow(-10); added != 0 || !close(p.Total(), 30) {
		t.Error("negative regrow changed the patch")
	}
}

func TestSetTotalDownKeepsPattern(t *testing.T) {
	p := patch.New(2, 1, 10, 20)
	p.Take(0, 0, 6) // tiles 4 and 10
	p.SetTotal(7)   // halve
	if !close(p.Tile(0, 0), 2) || !close(p.Tile(1, 0), 5) {
		t.Error("tiles are", p.Grid(), "expected 2 and 5")
	}
}

func TestSetTotalUpAndClamp(t *testing.T) {
	p := patch.New(3, 3, 10, 0)
	p.SetTotal(45)
	if !close(p.Total(), 45) {
		t.Error("total is", p.Total(), "expected 45")
	}
	p.SetTotal(500)
	if !close(p.Total(), 90) {
		t.Error("total is", p.Total(), "expected it capped at 90")
	}
	p.SetTotal(-1)
	if p.Total() != 0 {
		t.Error("total is", p.Total(), "expected 0")
	}
	checkTilesValid(t, p)
}

func TestGridIsACopy(t *testing.T) {
	p := patch.New(2, 2, 10, 20)
	g := p.Grid()
	g[0][0] = 99
	if p.Tile(0, 0) == 99 {
		t.Error("changing Grid's result changed the patch")
	}
	if len(g) != 2 || len(g[0]) != 2 {
		t.Error("grid is", len(g), "rows of", len(g[0]), "expected 2x2")
	}
}
