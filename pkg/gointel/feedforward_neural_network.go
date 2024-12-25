package gointel

import (
	"fmt"
	"math"
	"math/rand"
)

type ActivationFunc func(float64) float64

type NeuralNetwork struct {
	layers               []int
	weights              [][][]float64
	biases               [][]float64
	prevWeightDelta      [][][]float64
	prevBiasDelta        [][]float64
	learningRate         float64
	momentum             float64
	activation           ActivationFunc
	activationDerivative ActivationFunc
	errorThreshold       float64
}

func NewNeuralNetwork(layers []int, learningRate float64, momentum float64, errorThreshold float64) *NeuralNetwork {
	nn := &NeuralNetwork{
		layers:               layers,
		learningRate:         learningRate,
		momentum:             momentum,
		activation:           sigmoid,
		activationDerivative: sigmoidDerivative,
		errorThreshold:       errorThreshold,
	}

	nn.weights = make([][][]float64, len(layers)-1)
	nn.prevWeightDelta = make([][][]float64, len(layers)-1)
	for i := 0; i < len(layers)-1; i++ {
		nn.weights[i] = make([][]float64, layers[i+1])
		nn.prevWeightDelta[i] = make([][]float64, layers[i+1])
		for j := range nn.weights[i] {
			nn.weights[i][j] = make([]float64, layers[i])
			nn.prevWeightDelta[i][j] = make([]float64, layers[i])
			for k := range nn.weights[i][j] {
				nn.weights[i][j][k] = rand.Float64()*2 - 1
				nn.prevWeightDelta[i][j][k] = 0
			}
		}
	}

	nn.biases = make([][]float64, len(layers)-1)
	nn.prevBiasDelta = make([][]float64, len(layers)-1)
	for i := 0; i < len(layers)-1; i++ {
		nn.biases[i] = make([]float64, layers[i+1])
		nn.prevBiasDelta[i] = make([]float64, layers[i+1])
		for j := range nn.biases[i] {
			nn.biases[i][j] = rand.Float64()*2 - 1
			nn.prevBiasDelta[i][j] = 0
		}
	}

	return nn
}

func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

func sigmoidDerivative(x float64) float64 {
	return x * (1.0 - x)
}

func (nn *NeuralNetwork) SetActivation(activation ActivationFunc, derivative ActivationFunc) {
	nn.activation = activation
	nn.activationDerivative = derivative
}

func (nn *NeuralNetwork) Forward(inputs []float64) []float64 {
	if len(inputs) != nn.layers[0] {
		panic("Invalid input size")
	}

	currentLayer := inputs

	for i := 0; i < len(nn.layers)-1; i++ {
		nextLayer := make([]float64, nn.layers[i+1])
		for j := 0; j < nn.layers[i+1]; j++ {
			sum := 0.0
			for k := 0; k < nn.layers[i]; k++ {
				sum += currentLayer[k] * nn.weights[i][j][k]
			}
			nextLayer[j] = nn.activation(sum + nn.biases[i][j])
		}
		currentLayer = nextLayer
	}

	return currentLayer
}

func (nn *NeuralNetwork) Train(inputs []float64, targets []float64) float64 {
	if len(inputs) != nn.layers[0] {
		panic("Invalid input size")
	}
	if len(targets) != nn.layers[len(nn.layers)-1] {
		panic(fmt.Sprintf("Invalid target size. Expected %d, got %d", nn.layers[len(nn.layers)-1], len(targets)))
	}

	activations := make([][]float64, len(nn.layers))
	activations[0] = inputs

	for i := 0; i < len(nn.layers)-1; i++ {
		activations[i+1] = make([]float64, nn.layers[i+1])
		for j := 0; j < nn.layers[i+1]; j++ {
			sum := 0.0
			for k := 0; k < nn.layers[i]; k++ {
				sum += activations[i][k] * nn.weights[i][j][k]
			}
			activations[i+1][j] = nn.activation(sum + nn.biases[i][j])
		}
	}

	errors := make([][]float64, len(nn.layers))

	lastLayer := len(nn.layers) - 1
	errors[lastLayer] = make([]float64, nn.layers[lastLayer])
	totalError := 0.0
	for i := 0; i < nn.layers[lastLayer]; i++ {
		errors[lastLayer][i] = targets[i] - activations[lastLayer][i]
		totalError += math.Abs(errors[lastLayer][i])
	}

	for i := len(nn.layers) - 2; i > 0; i-- {
		errors[i] = make([]float64, nn.layers[i])
		for j := 0; j < nn.layers[i]; j++ {
			err := 0.0
			for k := 0; k < nn.layers[i+1]; k++ {
				err += errors[i+1][k] * nn.weights[i][k][j]
			}
			errors[i][j] = err
		}
	}

	for i := 0; i < len(nn.layers)-1; i++ {
		for j := 0; j < nn.layers[i+1]; j++ {
			for k := 0; k < nn.layers[i]; k++ {
				delta := nn.learningRate * errors[i+1][j] * nn.activationDerivative(activations[i+1][j]) * activations[i][k]
				delta += nn.momentum * nn.prevWeightDelta[i][j][k]
				nn.weights[i][j][k] += delta
				nn.prevWeightDelta[i][j][k] = delta
			}

			biasDelta := nn.learningRate * errors[i+1][j] * nn.activationDerivative(activations[i+1][j])
			biasDelta += nn.momentum * nn.prevBiasDelta[i][j]
			nn.biases[i][j] += biasDelta
			nn.prevBiasDelta[i][j] = biasDelta
		}
	}

	return totalError / float64(nn.layers[lastLayer])
}

func (nn *NeuralNetwork) TrainEpoch(trainingData []struct{ inputs, targets []float64 }) (float64, bool) {
	indices := rand.Perm(len(trainingData))
	epochError := 0.0
	for _, idx := range indices {
		err := nn.Train(trainingData[idx].inputs, trainingData[idx].targets)
		epochError += err
	}
	avgError := epochError / float64(len(trainingData))
	return avgError, avgError <= nn.errorThreshold
}

type NNLayerState struct {
	layers               []int
	inputSize            int
	outputSize           int
	maxLayers            int
	minNeurons           int
	maxNeurons           int
	trainingData         []struct{ inputs, targets []float64 }
	epochs               int
	learningRate         float64
	momentum             float64
	activation           ActivationFunc
	activationDerivative ActivationFunc
}

func (s *NNLayerState) Energy() float64 {
	nn := NewNeuralNetwork(s.layers, s.learningRate, s.momentum, 0.01)

	totalError := 0.0
	for i := 0; i < s.epochs; i++ {
		err, _ := nn.TrainEpoch(s.trainingData)
		totalError += err
	}

	avgError := totalError / float64(s.epochs)
	complexityPenalty := float64(len(s.layers)-2) * 0.1

	return avgError + complexityPenalty
}

func (s *NNLayerState) Neighbor() State {
	newLayers := make([]int, len(s.layers))
	copy(newLayers, s.layers)

	// Always preserve input and output layer sizes
	if len(newLayers) < 3 {
		newLayers = []int{s.inputSize, (s.inputSize + s.outputSize) / 2, s.outputSize}
	}

	operations := []func(){
		func() { s.mutateLayerSize(newLayers) },
		func() { s.addLayer(newLayers) },
		func() { s.removeLayer(newLayers) },
	}

	op := operations[rand.Intn(len(operations))]
	op()

	// Ensure first and last layers match input/output sizes
	newLayers[0] = s.inputSize
	newLayers[len(newLayers)-1] = s.outputSize

	return &NNLayerState{
		layers:       newLayers,
		inputSize:    s.inputSize,
		outputSize:   s.outputSize,
		maxLayers:    s.maxLayers,
		minNeurons:   s.minNeurons,
		maxNeurons:   s.maxNeurons,
		trainingData: s.trainingData,
		epochs:       s.epochs,
		learningRate: s.learningRate,
		momentum:     s.momentum,
	}
}

func (s *NNLayerState) mutateLayerSize(layers []int) {
	if len(layers) <= 2 {
		return
	}

	// Only mutate hidden layers
	layerIdx := rand.Intn(len(layers)-2) + 1
	delta := rand.Intn(5) - 2
	newSize := layers[layerIdx] + delta

	if newSize >= s.minNeurons && newSize <= s.maxNeurons {
		layers[layerIdx] = newSize
	}
}

func (s *NNLayerState) addLayer(layers []int) {
	if len(layers) >= s.maxLayers {
		return
	}

	// Only add between input and output layers
	newSize := rand.Intn(s.maxNeurons-s.minNeurons+1) + s.minNeurons
	position := rand.Intn(len(layers)-1) + 1

	newLayers := make([]int, 0, len(layers)+1)
	newLayers = append(newLayers, layers[:position]...)
	newLayers = append(newLayers, newSize)
	newLayers = append(newLayers, layers[position:]...)

	copy(layers, newLayers)
}

func (s *NNLayerState) removeLayer(layers []int) {
	if len(layers) <= 3 { // Keep at least one hidden layer
		return
	}

	// Only remove hidden layers
	position := rand.Intn(len(layers)-2) + 1

	newLayers := make([]int, 0, len(layers)-1)
	newLayers = append(newLayers, layers[:position]...)
	newLayers = append(newLayers, layers[position+1:]...)

	copy(layers, newLayers[:len(layers)-1])
}

func (s *NNLayerState) Copy() State {
	newLayers := make([]int, len(s.layers))
	copy(newLayers, s.layers)

	return &NNLayerState{
		layers:               newLayers,
		inputSize:            s.inputSize,
		outputSize:           s.outputSize,
		maxLayers:            s.maxLayers,
		minNeurons:           s.minNeurons,
		maxNeurons:           s.maxNeurons,
		trainingData:         s.trainingData,
		epochs:               s.epochs,
		learningRate:         s.learningRate,
		momentum:             s.momentum,
		activation:           s.activation,
		activationDerivative: s.activationDerivative,
	}
}

func OptimizeNNLayers(
	inputSize, outputSize int,
	trainingData []struct{ inputs, targets []float64 },
	initialTemp, finalTemp, coolingRate float64,
	iterationsPerTemp, epochs int,
	learningRate, momentum float64,
	activation ActivationFunc,
	activationDerivative ActivationFunc,
) []int {
	if len(trainingData) == 0 {
		panic("No training data provided")
	}

	if len(trainingData[0].targets) != outputSize {
		panic(fmt.Sprintf("Output size mismatch. Expected %d, got %d", outputSize, len(trainingData[0].targets)))
	}

	if len(trainingData[0].inputs) != inputSize {
		panic(fmt.Sprintf("Input size mismatch. Expected %d, got %d", inputSize, len(trainingData[0].inputs)))
	}

	initialState := &NNLayerState{
		layers:               []int{inputSize, (inputSize + outputSize) / 2, outputSize},
		inputSize:            inputSize,
		outputSize:           outputSize,
		maxLayers:            6,
		minNeurons:           2,
		maxNeurons:           100,
		trainingData:         trainingData,
		epochs:               epochs,
		learningRate:         learningRate,
		momentum:             momentum,
		activation:           activation,
		activationDerivative: activationDerivative,
	}

	sa := NewSimulatedAnnealing(initialTemp, finalTemp, coolingRate, iterationsPerTemp)
	bestState := sa.Optimize(initialState)

	return bestState.(*NNLayerState).layers
}

func NewInferredNeuralNetwork(
	inputSize, outputSize int,
	trainingData []struct{ inputs, targets []float64 },
	initialTemp, finalTemp, coolingRate float64,
	iterationsPerTemp, epochs int,
	learningRate, momentum, errorThreshold float64,
	activation ActivationFunc,
	activationDerivative ActivationFunc,
) *NeuralNetwork {
	layers := OptimizeNNLayers(inputSize,
		outputSize,
		trainingData,
		initialTemp,
		finalTemp,
		coolingRate,
		iterationsPerTemp,
		epochs,
		learningRate,
		momentum,
		activation,
		activationDerivative)
	ret := NewNeuralNetwork(layers, learningRate, momentum, errorThreshold)
	ret.SetActivation(activation, activationDerivative)
	return ret
}
