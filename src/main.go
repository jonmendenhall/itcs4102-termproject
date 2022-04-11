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

func MakeTerrain(width, height int) *Terrain {
	t := new(Terrain)
	t.width = width
	t.height = height
	t.height_map = make([]float64, width*height)
	return t
}

func (t *Terrain) HeightAt(x, y int) float64 {
	return t.height_map[y*t.width+x]
}

// sample terrain at a float location on the height map
func (t *Terrain) HeightAtFractional(x, y float64) float64 {
	// find which corner coordinates. and what percent along x and y of the cell
	x0f, xt := math.Modf(x)
	y0f, yt := math.Modf(y)
	x0 := int(x0f)
	y0 := int(y0f)
	x1 := int(math.Ceil(x))
	y1 := int(math.Ceil(y))

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
	var amplitude float64 = 1
	var period float64 = 32

	var p_i int64
	for p_i = 0; p_i < 4; p_i++ {
		i := 0
		fmt.Println("LAYER")
		for y := 0; y < t.height; y++ {
			for x := 0; x < t.width; x++ {
				xp := float64(x) / period
				yp := float64(y) / period
				x0 := int64(xp)
				y0 := int64(yp)
				xt := xp - float64(x0)
				yt := yp - float64(y0)

				var samples [4]float64
				var s_i int64
				for s_i = 0; s_i < 4; s_i++ {
					samples[s_i] = Interp(
						SampleRand(p_i+seed, x0-1, y0-1+s_i),
						SampleRand(p_i+seed, x0, y0-1+s_i),
						SampleRand(p_i+seed, x0+1, y0-1+s_i),
						SampleRand(p_i+seed, x0+2, y0-1+s_i),
						xt,
					)
				}
				t.height_map[i] += Interp(samples[0], samples[1], samples[2], samples[3], yt) * amplitude
				i++
			}
		}
		amplitude /= 2
		period /= 2
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

}

func main() {
	fmt.Println("ITCS-4102 Term Project: Mountain Map")
	fmt.Println()

	t := MakeTerrain(128, 128)
	t.GenerateTerrain(11)
	t.SavePNG("test3.png")
	t.RunErosionSimulation()

	// fmt.Printf("%d %d %d\n", t.width, t.height, len(t.height_map))
	// overall process:
	// - generate starting map using layered perlin noise
	// - run erosion simulation to update terrain
	// - make path between 2 points on map

	// can view heightmaps at http://www.procgenesis.com/SimpleHMV/simplehmv.html
}
