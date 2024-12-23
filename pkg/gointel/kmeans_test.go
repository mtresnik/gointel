package gointel

import (
	"github.com/mtresnik/gomath/pkg/gomath"
	"github.com/mtresnik/goutils/pkg/goutils"
	"image"
	"image/color"
	"image/gif"
	"math"
	"math/rand"
	"os"
	"testing"
)

func testKMeans_GIF(t *testing.T) {
	allPoints := generateAllPoints()
	boundingBox := gomath.NewBoundingBox(gomath.PointsToSpatial(allPoints...)...)

	renderer := LivePointRenderer{
		bounds:      boundingBox,
		points:      allPoints,
		Width:       1000,
		Height:      1000,
		PointRadius: 5,
		Padding:     50,
		delay:       100,
	}
	g := renderer.RenderFrames()
	f, err := os.Create("TestKMeans_GIF.gif")
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(f)

	// Encode and save the GIF
	err = gif.EncodeAll(f, g)
	if err != nil {
		panic(err)
	}
}

func generateAllPoints() []gomath.Point {
	allPoints := make([]gomath.Point, 0)
	numPointsPerCluster := 100
	center1 := gomath.NewPoint(0, 0, 1.0, 0.0, 0.0)
	radius1 := 0.5
	for i := 0; i < numPointsPerCluster; i++ {
		radius := rand.Float64() * radius1
		theta := rand.Float64() * 2 * math.Pi
		x := math.Cos(theta)*radius + center1.X()
		y := math.Sin(theta)*radius + center1.Y()
		randomPoint := gomath.NewPoint(x, y, 1.0, 0.0, 0.0)
		allPoints = append(allPoints, *randomPoint)
	}

	center2 := gomath.NewPoint(1, 0, 0.0, 1.0, 0.0)
	radius2 := 0.5
	for i := 0; i < numPointsPerCluster; i++ {
		radius := rand.Float64() * radius2
		theta := rand.Float64() * 2 * math.Pi
		x := math.Cos(theta)*radius + center2.X()
		y := math.Sin(theta)*radius + center2.Y()
		randomPoint := gomath.NewPoint(x, y, 0.0, 1.0, 0.0)
		allPoints = append(allPoints, *randomPoint)
	}

	center3 := gomath.NewPoint(0.5, 1, 0.0, 0.0, 1.0)
	radius3 := 0.5

	for i := 0; i < numPointsPerCluster; i++ {
		radius := rand.Float64() * radius3
		theta := rand.Float64() * 2 * math.Pi
		x := math.Cos(theta)*radius + center3.X()
		y := math.Sin(theta)*radius + center3.Y()
		randomPoint := gomath.NewPoint(x, y, 0.0, 0.0, 1.0)
		allPoints = append(allPoints, *randomPoint)
	}
	return allPoints
}

type PointRenderer struct {
	bounds      gomath.BoundingBox
	points      []gomath.Point
	centroids   []gomath.Point
	Width       int
	Height      int
	PointRadius int
	Padding     int
}

func (p *PointRenderer) convertPixels(point gomath.Spatial) image.Point {
	if p.bounds.Area() <= 0 {
		panic("Bounds not set")
	}
	if !p.bounds.Contains(point) {
		return image.Point{}
	}
	x := point.GetValues()[0]
	y := point.GetValues()[1]

	relX := (x - p.bounds.MinX) / (p.bounds.MaxX - p.bounds.MinX)
	relY := (y - p.bounds.MinY) / (p.bounds.MaxY - p.bounds.MinY)
	return image.Point{
		X: int(relX * float64(p.Width)),
		Y: int(relY * float64(p.Height)),
	}
}

func (p *PointRenderer) drawPoint(img *image.RGBA, point gomath.Spatial, color color.RGBA, radius int) {
	coords := p.convertPixels(point)
	goutils.FillCircle(img, coords.X, coords.Y, radius, color)
}

func (p *PointRenderer) padBounds() {
	minPoint := gomath.Point{Values: []float64{p.bounds.MinX, p.bounds.MinY}}
	maxPoint := gomath.Point{Values: []float64{p.bounds.MaxX, p.bounds.MaxY}}
	dx := maxPoint.X() - minPoint.X()
	dy := maxPoint.Y() - minPoint.Y()

	dt := dx * (float64(p.Padding) / float64(p.Width))
	dv := dy * (float64(p.Padding) / float64(p.Height))

	minPoint = gomath.Point{Values: []float64{p.bounds.MinX - dt, p.bounds.MinY - dv}}
	maxPoint = gomath.Point{Values: []float64{p.bounds.MaxX + dt, p.bounds.MaxY + dv}}

	p.bounds.MinX = minPoint.X()
	p.bounds.MinY = minPoint.Y()
	p.bounds.MaxX = maxPoint.X()
	p.bounds.MaxY = maxPoint.Y()
}

func (p *PointRenderer) Build() *image.RGBA {
	p.padBounds()
	img := image.NewRGBA(image.Rect(0, 0, p.Width, p.Height))
	goutils.FillRectangle(img, 0, 0, p.Width, p.Height, goutils.COLOR_WHITE)
	for _, point := range p.points {
		r := uint8(point.Values[2] * 255.0)
		g := uint8(point.Values[3] * 255.0)
		b := uint8(point.Values[4] * 255.0)
		p.drawPoint(img, point, color.RGBA{r, g, b, 255}, p.PointRadius)
	}

	for _, centroid := range p.centroids {
		p.drawPoint(img, centroid, goutils.COLOR_BLACK, 2*p.PointRadius)
	}
	return img
}

func TestKMeans_Live(t *testing.T) {

}

type LivePointRenderer struct {
	bounds      gomath.BoundingBox
	points      []gomath.Point
	Frames      []*image.Paletted
	Width       int
	Height      int
	PointRadius int
	Padding     int
	delay       int
}

func (p *LivePointRenderer) Build() {
	var distanceFunction gomath.DistanceFunction = gomath.EuclideanDistance
	request := KMeansRequest{
		K:                3,
		Points:           p.points,
		DistanceFunction: &distanceFunction,
		Listeners:        []KMeansUpdateListener{p},
	}
	KMeans(request)
}

func (p *LivePointRenderer) Update(result KMeansResult) {
	internalRenderer := PointRenderer{
		bounds:      p.bounds,
		points:      p.points,
		centroids:   result,
		Width:       p.Width,
		Height:      p.Height,
		PointRadius: p.PointRadius,
		Padding:     p.Padding,
	}

	img := internalRenderer.Build()
	p.Frames = append(p.Frames, goutils.ConvertImageToPaletted(img))
}

func (p *LivePointRenderer) RenderFrames() *gif.GIF {
	p.Build()
	if len(p.Frames) == 0 {
		return nil
	}
	images := p.Frames
	delays := make([]int, len(images))
	for i := 0; i < len(images); i++ {
		delays[i] = p.delay
	}
	retGif := &gif.GIF{
		Image:     images,
		Delay:     delays,
		LoopCount: 0,
	}
	return retGif
}
