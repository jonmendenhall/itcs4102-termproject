package main

import "fmt"

type RainDrop struct {
	x, y      float32
	vx, vy    float32
	dirt_mass float32
}

func (t *Terrain) RunErosionSimulation() {
	fmt.Println("SIM")
}
