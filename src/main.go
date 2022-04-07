package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"os"
)

type Terrain struct {
	width      int
	height     int
	height_map []float32
}

func MakeTerrain(width, height int) *Terrain {
	t := new(Terrain)
	t.width = width
	t.height = height
	t.height_map = make([]float32, width*height)
	return t
}

func (t *Terrain) HeightAt(x, y int) float32 {
	return t.height_map[y*t.width+x]
}

func Interp(a, b, c, d, x float64) float64 {
	return x*(x*(x*(-a+b-c+d)+2*a-2*b+c-d)-a+c) + b
}

func (t *Terrain) AssignRandomHeights(min, max float32) {
	i := 0
	length := max - min
	for y := 0; y < t.height; y++ {
		for x := 0; x < t.width; x++ {
			t.height_map[i] = rand.Float32()*length + min
			i++
		}
	}
}

func SampleRand(seed, x, y int64) float64 {
	rand.Seed((seed+x^1308123)*13423198 + y ^ 230813)
	return rand.Float64()
}

func (t *Terrain) GenerateTerrain() {
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
						SampleRand(p_i, x0-1, y0-1+s_i),
						SampleRand(p_i, x0, y0-1+s_i),
						SampleRand(p_i, x0+1, y0-1+s_i),
						SampleRand(p_i, x0+2, y0-1+s_i),
						xt,
					)
				}
				t.height_map[i] += float32(Interp(samples[0], samples[1], samples[2], samples[3], yt) * amplitude)
				i++
			}
		}
		amplitude /= 2
		period /= 2
	}
}

func (t *Terrain) SavePNG(path string) {

	// check range of terrain to normalize
	var min, max float32
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
	t.GenerateTerrain()
	t.SavePNG("test.png")

	fmt.Printf("%d %d %d\n", t.width, t.height, len(t.height_map))
	// overall process:
	// - generate starting map using layered perlin noise
	// - run erosion simulation to update terrain
	// - make path between 2 points on map

	// can view heightmaps at http://www.procgenesis.com/SimpleHMV/simplehmv.html
}
