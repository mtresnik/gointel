package gointel

import (
	"math"
	"math/rand"
)

type State interface {
	Energy() float64
	Neighbor() State
	Copy() State
}

type SimulatedAnnealing struct {
	initialTemp       float64
	finalTemp         float64
	coolingRate       float64
	iterationsPerTemp int
}

func NewSimulatedAnnealing(initialTemp, finalTemp, coolingRate float64, iterationsPerTemp int) *SimulatedAnnealing {
	return &SimulatedAnnealing{
		initialTemp:       initialTemp,
		finalTemp:         finalTemp,
		coolingRate:       coolingRate,
		iterationsPerTemp: iterationsPerTemp,
	}
}

func (sa *SimulatedAnnealing) Optimize(initialState State) State {
	currentState := initialState.Copy()
	bestState := initialState.Copy()
	currentTemp := sa.initialTemp

	for currentTemp > sa.finalTemp {
		for i := 0; i < sa.iterationsPerTemp; i++ {
			neighborState := currentState.Neighbor()
			currentEnergy := currentState.Energy()
			neighborEnergy := neighborState.Energy()
			energyDiff := neighborEnergy - currentEnergy

			if acceptTransition(energyDiff, currentTemp) {
				currentState = neighborState.Copy()

				if currentState.Energy() < bestState.Energy() {
					bestState = currentState.Copy()
				}
			}
		}

		currentTemp *= 1 - sa.coolingRate
	}

	return bestState
}

func acceptTransition(energyDiff, temperature float64) bool {
	if energyDiff < 0 {
		return true
	}
	probability := math.Exp(-energyDiff / temperature)
	return rand.Float64() < probability
}
