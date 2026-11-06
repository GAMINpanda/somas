// Package weather models clouds drifting over the patch grid. Cloud cover at an agent's cell reduces the sunlight it gets.
// The grid wraps around (a torus), so a cloud leaving one edge comes back on the opposite edge.
package weather

import (
	"math"
	"math/rand/v2"
)

// Cloud is a round cloud measured in grid cells. (X, Y) is its centre, and cell (x, y) has its centre at (x+0.5, y+0.5).
// VX and VY are how far it drifts each round.
type Cloud struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Radius float64 `json:"radius"`
	VX     float64 `json:"vx"`
	VY     float64 `json:"vy"`
}

// Weather holds the clouds over a Width x Height grid
type Weather struct {
	Width, Height int
	Clouds        []Cloud
}

const (
	minRadius = 1.5 // cells
	maxRadius = 3.0
	minSpeed  = 0.5 // cells per round
	maxSpeed  = 1.5
)

// NewRandom returns numClouds clouds at random positions, sizes and headings
func NewRandom(width, height, numClouds int) *Weather {
	w := &Weather{Width: width, Height: height}
	for range numClouds {
		heading := rand.Float64() * 2 * math.Pi
		speed := minSpeed + rand.Float64()*(maxSpeed-minSpeed)
		w.Clouds = append(w.Clouds, Cloud{
			X:      rand.Float64() * float64(width),
			Y:      rand.Float64() * float64(height),
			Radius: minRadius + rand.Float64()*(maxRadius-minRadius),
			VX:     speed * math.Cos(heading),
			VY:     speed * math.Sin(heading),
		})
	}
	return w
}

// Step moves every cloud by its velocity, wrapping around the edges
func (w *Weather) Step() {
	for i := range w.Clouds {
		c := &w.Clouds[i]
		c.X = wrap(c.X+c.VX, float64(w.Width))
		c.Y = wrap(c.Y+c.VY, float64(w.Height))
	}
}

// Coverage returns how cloudy cell (x, y) is, from 0 (clear sky) to 1 (under the middle of a cloud).
// Each cloud is thickest at its centre and fades to nothing at its radius; overlapping clouds combine
// as 1 - (1-a)(1-b)..., so cover can never go above 1.
func (w *Weather) Coverage(x, y int) float64 {
	clear := 1.0
	for _, c := range w.Clouds {
		dx := wrappedGap(float64(x)+0.5, c.X, float64(w.Width))
		dy := wrappedGap(float64(y)+0.5, c.Y, float64(w.Height))
		if c.Radius <= 0 {
			continue
		}
		cover := max(0, 1-math.Hypot(dx, dy)/c.Radius)
		clear *= 1 - cover
	}
	return 1 - clear
}

// wrap brings v into [0, size)
func wrap(v, size float64) float64 {
	v = math.Mod(v, size)
	if v < 0 {
		v += size
	}
	return v
}

// wrappedGap is the shortest distance between a and b on a loop of the given size
func wrappedGap(a, b, size float64) float64 {
	d := math.Abs(a - b)
	return min(d, size-d)
}
