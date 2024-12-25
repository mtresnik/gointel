# gointel
[![build status](https://github.com/mtresnik/gointel/actions/workflows/go.yml/badge.svg)](https://github.com/mtresnik/gointel/actions/workflows/go.yml/)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://github.com/mtresnik/gointel/blob/main/LICENSE)
[![version](https://img.shields.io/badge/version-1.1.12-blue)](https://github.com/mtresnik/gomath/releases/tag/v1.1.12)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-green.svg?style=flat-square)](https://makeapullrequest.com)
<hr>

gointel is a Go implementation of common machine learning models.

### Constraints:

Constraints are used to limit the search space given the initial variables and domains. **Local Constraints** provide *path consistency* while **Global Constraints** provide *absolute consistency*.

**Local Constraints** imply that local consistency tends to global consistency. Used for monotone solutions. </br> Ex: `sumLessThan()`, `localAllDiff()`

**Global Constraints** imply that the entire sequence is needed to know consistency. </br> Ex: `isName()`, `equals()`

**Reusable Constraints** accomplishes both local and global consistency. These can both prune the search space while exploring and be reused as global constraints at the end. </br> Ex: `min()`, `max()`

#### n-Queens

| (n=16, solutions=14772512, time=14.78s)             | 
|-----------------------------------------------------|
| <img src="res/NQueens_16_14772512.png" width="300"> |

#### Feed Forward Neural Network

| (xor problem, training set=4, data={(0,0),(0)}, {(0,1),(1)}, {(1,0),(1)}, {(1,1),(0)})            | 
|-----------------------------------------------------|
| <img src="res/NeuralNetworkInferred.png" width="300"> |



### Sample Code

In your project run:
```
go mod download github.com/mtresnik/goutils
go mod download github.com/mtresnik/gomath 
go mod download github.com/mtresnik/gointel 
```

Your `go.mod` file should look like this:
```go 
module mymodule

go 1.23.3

require github.com/mtresnik/gointel v1.1.12
```


Then in your go files you should be able to run different common models:

```go 
package main

import "github.com/mtresnik/gointel/pkg/gointel"

func main() {
	// Common XOR / Hello World for Neural Networks
	layers := []int{2, 4, 1}
	nn := gointel.NewNeuralNetwork(layers, 0.1, 0.9, 0.01) // 0.01 error threshold

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
	for i := 0; i < maxEpochs; i++ {
		_, shouldStop := nn.TrainEpoch(trainingData)
		if shouldStop {
			println("Reached error threshold at epoch:", i)
			break
		}
	}
}
```