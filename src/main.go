package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
)

type Terrain struct {
	width      int
	height     int
	height_map []float64
}

// initialize a new terrain instance
func MakeTerrain(width, height int) *Terrain {
	t := new(Terrain)
	t.width = width
	t.height = height
	t.height_map = make([]float64, width*height)
	return t
}

// create a duplicate terrain instance
func (t *Terrain) Copy() *Terrain {
	t2 := MakeTerrain(t.width, t.height)
	copy(t2.height_map, t.height_map)
	return t2
}

// retreive the height of the terrain at a certain location
func (t *Terrain) HeightAt(x, y int) float64 {
	if x >= 0 && y >= 0 && x <= t.width-1 && y <= t.height-1 {
		return t.height_map[y*t.width+x]
	}
	return 0
}

// safely adjust the height of the terrain at a given location within some specific bounds
func (t *Terrain) AdjustHeightAt(x, y int, dh, h_min, h_max float64) {
	if x >= 0 && y >= 0 && x <= t.width-1 && y <= t.height-1 {
		h := math.Min(math.Max(t.height_map[y*t.width+x]+dh, h_min), h_max)
		t.height_map[y*t.width+x] = h
	}
}

// sample terrain at a float location on the height map
func (t *Terrain) HeightAtFractional(x, y float64) float64 {
	// find which corner coordinates. and what percent along x and y of the cell
	x0f, xt := math.Modf(x)
	y0f, yt := math.Modf(y)
	x0 := int(x0f)
	y0 := int(y0f)
	x1 := x0 + 1
	y1 := y0 + 1

	// sample height on top and bottom edges
	a := t.HeightAt(x0, y0)*(1.0-xt) + t.HeightAt(x1, y0)*(xt)
	b := t.HeightAt(x0, y1)*(1.0-xt) + t.HeightAt(x1, y1)*(xt)

	// average heights to match location of target point
	return a*(1.0-yt) + b*yt
}

// sample terrain at a float location on the height map
func (t *Terrain) AccelerationAtFractional(x, y float64) (ax, ay float64) {
	// find which corner coordinates. and what percent along x and y of the cell
	x0f, xt := math.Modf(x)
	y0f, yt := math.Modf(y)
	x0 := int(x0f)
	y0 := int(y0f)
	x1 := x0 + 1
	y1 := y0 + 1

	// sample height on left and right edges of this cell
	xa := t.HeightAt(x0, y0)*(1.0-yt) + t.HeightAt(x0, y1)*yt
	xb := t.HeightAt(x1, y0)*(1.0-yt) + t.HeightAt(x1, y1)*yt

	// sample height on top and bottom edges of this cell
	ya := t.HeightAt(x0, y0)*(1.0-xt) + t.HeightAt(x1, y0)*xt
	yb := t.HeightAt(x0, y1)*(1.0-xt) + t.HeightAt(x1, y1)*xt

	// difference in height will match acceleration on each axis
	return xb - xa, yb - ya
}

func (t *Terrain) AdjustTerrainAt(x, y, dh float64) {
	// find which corner coordinates. and what percent along x and y of the cell
	x0f, xt := math.Modf(x)
	y0f, yt := math.Modf(y)
	x0 := int(x0f)
	y0 := int(y0f)
	x1 := x0 + 1
	y1 := y0 + 1

	// find binds of adjustment based on the height at each location in this cell
	h_min := math.Inf(1)
	h_max := math.Inf(-1)
	hs := [4]float64{
		t.HeightAt(x0, y0),
		t.HeightAt(x1, y0),
		t.HeightAt(x0, y1),
		t.HeightAt(x1, y1),
	}
	for _, h := range hs {
		h_min = math.Min(h_min, h)
		h_max = math.Max(h_max, h)
	}

	// adjust corner values weighted by location within the cell
	t.AdjustHeightAt(x0, y0, dh*(1.0-xt)*(1.0-yt), h_min, h_max)
	t.AdjustHeightAt(x1, y0, dh*(xt)*(1.0-yt), h_min, h_max)
	t.AdjustHeightAt(x0, y1, dh*(1.0-xt)*(yt), h_min, h_max)
	t.AdjustHeightAt(x1, y1, dh*(xt)*(yt), h_min, h_max)
}

// perform cubic interpolation
func Interp(a, b, c, d, x float64) float64 {
	return x*(x*(x*(-a+b-c+d)+2*a-2*b+c-d)-a+c) + b
}

func (t *Terrain) AssignRandomHeights(min, max float64) {
	i := 0
	length := max - min
	for y := 0; y < t.height; y++ {
		for x := 0; x < t.width; x++ {
			t.height_map[i] = rand.Float64()*length + min
			i++
		}
	}
}

func SampleRand(seed, x, y int64) float64 {
	var s = seed + x*374761393 + y*668265263
	s = (s ^ (s >> 13)) * 1274126177
	rand.Seed(s ^ (s >> 16))
	return rand.Float64()
}

func (t *Terrain) GenerateTerrain(seed int64) {
	fmt.Println("Generating initial terrain with noise...")

	// initial layer properties
	var amplitude float64 = 100
	var period float64 = 16

	var p_i int64
	for p_i = 0; p_i < 4; p_i++ {
		i := 0
		fmt.Printf("[Layer %d] period=%.2f  amplitude=%.2f\n", p_i+1, period, amplitude)
		for y := 0; y < t.height; y++ {
			for x := 0; x < t.width; x++ {
				// find which cell in the noise layer this corresponds to
				xp := float64(x) / period
				yp := float64(y) / period
				x0 := int64(xp)
				y0 := int64(yp)
				xt := xp - float64(x0)
				yt := yp - float64(y0)

				var samples [4]float64
				var s_i int64
				for s_i = 0; s_i < 4; s_i++ {
					// cubic interpolation on sampling across each row
					samples[s_i] = Interp(
						SampleRand(p_i+seed, x0-1, y0-1+s_i),
						SampleRand(p_i+seed, x0, y0-1+s_i),
						SampleRand(p_i+seed, x0+1, y0-1+s_i),
						SampleRand(p_i+seed, x0+2, y0-1+s_i),
						xt,
					)
				}
				// cubic interpolation based on resulting row samples for the entire cell
				t.height_map[i] += Interp(samples[0], samples[1], samples[2], samples[3], yt) * amplitude
				i++
			}
		}
		// move to higher frequency (lower period) waves with lower amplitude
		amplitude *= 0.3
		period *= 0.5
	}
}

func (t *Terrain) SavePNG(path string) {

	// check range of terrain to normalize
	var min, max float64
	for i, x := range t.height_map {
		if i == 0 || x < min {
			min = x
		}
		if i == 0 || x > max {
			max = x
		}
	}

	length := max - min

	// create image from height map
	img := image.NewGray(image.Rect(0, 0, t.width, t.height))
	i := 0
	for y := 0; y < t.height; y++ {
		for x := 0; x < t.width; x++ {
			value := uint8((t.height_map[i] - min) / length * 255)
			img.Set(x, y, color.Gray{value})
			i++
		}
	}

	file, err := os.Create(path)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()
	png.Encode(file, img)
	fmt.Printf("Image saved as %s\n", path)
}

func (t *Terrain) ScaleUp(scale int) *Terrain {
	new_t := MakeTerrain(t.width*scale, t.height*scale)

	scale_x := float64(t.width-1) / float64(new_t.width-1)
	scale_y := float64(t.height-1) / float64(new_t.height-1)

	i := 0
	for y := 0; y < new_t.height; y++ {
		for x := 0; x < new_t.width; x++ {
			new_t.height_map[i] = t.HeightAtFractional(float64(x)*scale_x, float64(y)*scale_y)
			i++
		}
	}
	return new_t
}

func main() {
	// overall process:
	// - generate starting map using layered perlin noise
	// - run erosion simulation to update terrain

	// can view heightmaps at http://www.procgenesis.com/SimpleHMV/simplehmv.html

	fmt.Println("ITCS-4102 Term Project: Mountain Map")
	fmt.Println()

	fmt.Print("Enter a Grid Size: ")
	var grid_size int
	fmt.Scanf("%d", &grid_size)

	fmt.Print("Enter a seed for random generation: ")
	var seed int64
	fmt.Scanf("%d", &seed)

	fmt.Print("Enter the file name: ")
	var file_name string
	fmt.Scanf("%s", &file_name)

	// make initial terrain
	t := MakeTerrain(grid_size, grid_size)
	t.GenerateTerrain(seed)

	// scale up for more detail
	t4 := t.ScaleUp(4)
	t4.SavePNG(fmt.Sprintf("%s.png", file_name))

	// run simulation
	t4.RunErosionSimulation(3000)
	t4.SavePNG(fmt.Sprintf("%s_sim.png", file_name))

}
