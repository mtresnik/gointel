package gointel

import (
	"github.com/mtresnik/goutils/pkg/goutils"
	"image"
	"image/gif"
	"image/png"
	"math"
	"os"
	"testing"
)

func drawNeuralNetwork(nn *NeuralNetwork) *image.RGBA {
	imageSize := 500
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	goutils.FillRectangle(img, 0, 0, imageSize, imageSize, goutils.COLOR_WHITE)
	for row := 0; row < imageSize; row++ {
		t := float64(row) / float64(imageSize)
		for col := 0; col < imageSize; col++ {
			s := float64(col) / float64(imageSize)
			input := []float64{t, s}
			output := nn.Forward(input)
			goutils.FillRectangle(img, col, row, 1, 1, goutils.Gradient(output[0], goutils.COLOR_WHITE, goutils.COLOR_BLACK))
		}
	}
	return img
}

func TestNeuralNetwork(t *testing.T) {
	layers := []int{2, 4, 1}
	nn := NewNeuralNetwork(layers, 0.1, 0.9, 0.01) // 0.01 error threshold

	trainingData := []struct {
		inputs  []float64
		targets []float64
	}{
		{[]float64{0, 0}, []float64{0}},
		{[]float64{0, 1}, []float64{1}},
		{[]float64{1, 0}, []float64{1}},
		{[]float64{1, 1}, []float64{0}},
	}

	maxEpochs := 10000
	frames := []*image.RGBA{}
	for i := 0; i < maxEpochs; i++ {
		avgError, shouldStop := nn.TrainEpoch(trainingData)
		if i%100 == 0 {
			println("Epoch:", i, "Error:", avgError)
			frames = append(frames, drawNeuralNetwork(nn))
		}
		if shouldStop {
			println("Reached error threshold at epoch:", i)
			break
		}
	}
	delays := make([]int, len(frames))
	for i := 0; i < len(frames); i++ {
		delays[i] = 10
	}
	images := goutils.Map(frames, func(img *image.RGBA) *image.Paletted {
		return goutils.ConvertImageToPaletted(img)
	})
	retGif := &gif.GIF{
		Image:     images,
		Delay:     delays,
		LoopCount: 0,
	}
	f, err := os.Create("TestNeuralNetwork.gif")
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
	err = gif.EncodeAll(f, retGif)
	if err != nil {
		panic(err)
	}
}

func TestNeuralNetworkInferred(t *testing.T) {
	trainingData := []struct {
		inputs  []float64
		targets []float64
	}{
		{[]float64{0, 0}, []float64{0}},
		{[]float64{0, 1}, []float64{1}},
		{[]float64{1, 0}, []float64{1}},
		{[]float64{1, 1}, []float64{0}},
	}
	nn := NewInferredNeuralNetwork(2, 1, trainingData, 10.0, 0.1, 0.1, 100, 10, 0.1, 0.9, 0.01, sigmoid, sigmoidDerivative)
	for i := 0; i < 10000; i++ {
		_, shouldStop := nn.TrainEpoch(trainingData)
		if shouldStop {
			break
		}
	}
	img := drawNeuralNetwork(nn)
	imageName := "TestNeuralNetworkInferred.png"
	file, err := os.Create(imageName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}

func generateSineData(start, end float64, points int) []struct{ inputs, targets []float64 } {
	data := make([]struct{ inputs, targets []float64 }, points)
	step := (end - start) / float64(points-1)

	for i := 0; i < points; i++ {
		x := start + float64(i)*step
		data[i] = struct{ inputs, targets []float64 }{
			inputs:  []float64{x},
			targets: []float64{math.Sin(x)},
		}
	}
	return data
}

func testNeuralNetworkSine(t *testing.T) {
	sineData := generateSineData(-math.Pi, math.Pi, 40)
	tanh := func(x float64) float64 {
		return math.Tanh(x)
	}

	tanhDerivative := func(x float64) float64 {
		return 1.0 - math.Pow(math.Tanh(x), 2)
	}
	sineNN := NewInferredNeuralNetwork(
		1, 1,
		sineData,
		10.0, 1.0, 0.1,
		100, 50,
		0.05, 0.9, 0.01,
		tanh, tanhDerivative,
	)
	for i := 0; i < 10000; i++ {
		_, shouldStop := sineNN.TrainEpoch(sineData)
		if shouldStop {
			break
		}
	}
	imageSize := 500
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	goutils.FillRectangle(img, 0, 0, imageSize, imageSize, goutils.COLOR_WHITE)
	for col := 0; col < imageSize; col++ {
		s := float64(col) / float64(imageSize)
		input := []float64{s*math.Pi*2 - math.Pi}
		output := sineNN.Forward(input)
		row := imageSize - int(float64(imageSize)*(output[0]+1.0)/2.0)
		if row < 0 {
			row = 0
		}
		if row > imageSize-1 {
			row = imageSize - 1
		}
		img.Set(col, row, goutils.COLOR_BLACK)
	}
	imageName := "TestNeuralNetworkSine.png"
	file, err := os.Create(imageName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}
