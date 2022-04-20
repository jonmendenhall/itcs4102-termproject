package main

import (
	"fmt"
	"math"
	"math/rand"
)

type RainDrop struct {
	x, y   float64
	vx, vy float64
	dirt   float64
}

func (t *Terrain) RunErosionSimulation(num_drops int) {
	fmt.Printf("Running erosion simulation with %d raindrops...\n", num_drops)

	t_copy := t.Copy()

	for i := 0; i < num_drops; i++ {
		// make new rain drop in random location
		drop := new(RainDrop)
		drop.x = rand.Float64() * float64(t.width)
		drop.y = rand.Float64() * float64(t.height)

		var dt float64 = 0.05

		for sim_i := 0; sim_i < 2000; sim_i++ {
			// drop is off the grid
			if drop.x < 0 || drop.y < 0 || drop.x > float64(t.width) || drop.y > float64(t.height) {
				break
			}
			// sample acceleration based on slope of terrain
			ax, ay := t.AccelerationAtFractional(drop.x, drop.y)

			momentum := 0.999

			// integrate acceleration to velocity
			drop.vx = -ax*(1.0-momentum) + drop.vx*momentum
			drop.vy = -ay*(1.0-momentum) + drop.vy*momentum

			// integrate velocity to position
			drop.x += drop.vx * dt
			drop.y += drop.vy * dt

			// absorb some dirt
			dh := 0.1
			vel := math.Sqrt(drop.vx*drop.vx + drop.vy*drop.vy)
			dh *= (vel + 2)
			t_copy.AdjustTerrainAt(drop.x, drop.y, -dh)
			drop.dirt += dh

			// stop simulation of rain drop when stops moving
			if sim_i > 100 && math.Abs(drop.vx) < 0.01 && math.Abs(drop.vy) < 0.01 {
				break
			}
		}
		// deposit collected dirt at end position
		t_copy.AdjustTerrainAt(drop.x, drop.y, drop.dirt*0.1)

		// copy adjusted terrain to original terrain
		copy(t.height_map, t_copy.height_map)
	}
}
