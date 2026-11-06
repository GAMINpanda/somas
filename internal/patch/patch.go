// Package patch models the vegetable patch as a grid of tiles, so the UI can show where vegetables are eaten and regrow.
// The server's CPR total stays authoritative; the patch is kept in line with it using SetTotal.
package patch

import (
	"cmp"
	"slices"
)

// Patch is a Width x Height grid of tiles, each holding between 0 and TileCapacity vegetables
type Patch struct {
	Width, Height int
	TileCapacity  float64
	tiles         []float64 // row-major: tile (x, y) is tiles[y*Width+x]
}

// New returns a patch with the initial total spread evenly across its tiles (capped at full capacity)
func New(width, height int, tileCapacity, initial float64) *Patch {
	p := &Patch{Width: width, Height: height, TileCapacity: tileCapacity, tiles: make([]float64, width*height)}
	p.SetTotal(initial)
	return p
}

// Tile returns the vegetables on tile (x, y)
func (p *Patch) Tile(x, y int) float64 {
	return p.tiles[y*p.Width+x]
}

// Grid returns a copy of the tiles as rows, grid[y][x], ready for recording or drawing
func (p *Patch) Grid() [][]float64 {
	grid := make([][]float64, p.Height)
	for y := range grid {
		grid[y] = slices.Clone(p.tiles[y*p.Width : (y+1)*p.Width])
	}
	return grid
}

// Total returns the vegetables across all tiles
func (p *Patch) Total() float64 {
	total := 0.0
	for _, v := range p.tiles {
		total += v
	}
	return total
}

// Capacity returns the most the whole patch can hold
func (p *Patch) Capacity() float64 {
	return p.TileCapacity * float64(len(p.tiles))
}

// Take eats up to amount, starting from tile (x, y) and moving to the nearest tiles outward.
// It returns how much was actually taken, which is less than amount if the patch runs out.
func (p *Patch) Take(x, y int, amount float64) float64 {
	taken := 0.0
	for _, i := range p.nearestFirst(x, y) {
		if taken >= amount {
			break
		}
		bite := min(p.tiles[i], amount-taken)
		p.tiles[i] -= bite
		taken += bite
	}
	return taken
}

// Regrow adds up to amount, giving each tile a share in proportion to its free room, so no tile goes over capacity.
// It returns how much was actually added, which is less than amount if the patch fills up.
func (p *Patch) Regrow(amount float64) float64 {
	if amount <= 0 {
		return 0
	}
	room := p.Capacity() - p.Total()
	if room <= 0 {
		return 0
	}
	if amount >= room {
		for i := range p.tiles {
			p.tiles[i] = p.TileCapacity
		}
		return room
	}
	for i, v := range p.tiles {
		p.tiles[i] = v + amount*(p.TileCapacity-v)/room
	}
	return amount
}

// SetTotal brings the patch to the given total (clamped to 0..Capacity).
// Extra vegetables regrow in proportion to free room, and missing vegetables are removed in proportion to what each tile holds,
// so the pattern of eaten and full tiles is kept.
func (p *Patch) SetTotal(total float64) {
	total = max(0, min(total, p.Capacity()))
	current := p.Total()
	switch {
	case total > current:
		p.Regrow(total - current)
	case total < current:
		scale := total / current
		for i := range p.tiles {
			p.tiles[i] *= scale
		}
	}
}

// nearestFirst lists tile indices by distance from (x, y), breaking ties by row then column so the order is the same every time
func (p *Patch) nearestFirst(x, y int) []int {
	order := make([]int, len(p.tiles))
	for i := range order {
		order[i] = i
	}
	dist := func(i int) int {
		dx, dy := i%p.Width-x, i/p.Width-y
		return dx*dx + dy*dy
	}
	slices.SortStableFunc(order, func(a, b int) int { return cmp.Compare(dist(a), dist(b)) })
	return order
}
