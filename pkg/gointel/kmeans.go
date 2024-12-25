package gointel

import (
	"github.com/mtresnik/gomath/pkg/gomath"
	"math"
	"math/rand"
)

type KMeansResult []gomath.Point

type KMeansRequest struct {
	K                int
	Points           []gomath.Point
	DistanceFunction *gomath.DistanceFunction
	Listeners        []KMeansUpdateListener
	NumAttempts      int
}

type KMeansUpdateListener interface {
	Update(centroids KMeansResult)
}

func VisitKMeansUpdateListeners(listeners []KMeansUpdateListener, centroids KMeansResult) {
	if len(listeners) == 0 {
		return
	}
	for _, listener := range listeners {
		listener.Update(centroids)
	}
}

func NewKMeansRequest(k int, points []gomath.Point, distanceFunction *gomath.DistanceFunction, listeners ...KMeansUpdateListener) KMeansRequest {
	return KMeansRequest{
		K:                k,
		Points:           points,
		DistanceFunction: distanceFunction,
		Listeners:        listeners,
		NumAttempts:      k * 2,
	}
}

func kmeansInit(points []gomath.Point, k int, distFunc gomath.DistanceFunction) []gomath.Point {
	centroids := make([]gomath.Point, k)
	centroids[0] = points[rand.Intn(len(points))]

	for i := 1; i < k; i++ {
		distances := make([]float64, len(points))
		sum := 0.0

		for j, point := range points {
			minDist := math.MaxFloat64
			for c := 0; c < i; c++ {
				dist := distFunc(centroids[c], point)
				if dist < minDist {
					minDist = dist
				}
			}
			distances[j] = minDist * minDist
			sum += distances[j]
		}

		target := rand.Float64() * sum
		currentSum := 0.0
		for j, dist := range distances {
			currentSum += dist
			if currentSum >= target {
				centroids[i] = points[j]
				break
			}
		}
	}
	return centroids
}

func KMeans(request KMeansRequest) KMeansResult {
	if request.K <= 0 || len(request.Points) <= 0 {
		return nil
	}

	distFunc := gomath.EuclideanDistance
	if request.DistanceFunction != nil {
		distFunc = *request.DistanceFunction
	}

	dimension := len(request.Points[0].Values)
	bestCentroids := make([]gomath.Point, request.K)
	for i := range bestCentroids {
		bestCentroids[i].Values = make([]float64, dimension)
	}

	bestInertia := math.MaxFloat64
	assignments := make([]int, len(request.Points))

	for attempt := 0; attempt < 10; attempt++ {
		VisitKMeansUpdateListeners(request.Listeners, bestCentroids)
		centroids := make([]gomath.Point, request.K)
		for i := range centroids {
			centroids[i].Values = make([]float64, dimension)
		}
		centroids[0] = request.Points[rand.Intn(len(request.Points))]

		for i := 1; i < request.K; i++ {
			maxDist := 0.0
			nextIdx := 0
			for j, point := range request.Points {
				minDist := math.MaxFloat64
				for k := 0; k < i; k++ {
					if d := distFunc(centroids[k], point); d < minDist {
						minDist = d
					}
				}
				if minDist > maxDist {
					maxDist = minDist
					nextIdx = j
				}
			}
			centroids[i] = request.Points[nextIdx]
		}

		inertia := 0.0
		changed := true
		counts := make([]int, request.K)
		sums := make([][]float64, request.K)
		for i := range sums {
			sums[i] = make([]float64, dimension)
		}

		for iter := 0; changed && iter < 100; iter++ {
			changed = false
			inertia = 0

			for i, point := range request.Points {
				closest := 0
				minDist := distFunc(centroids[0], point)

				for j := 1; j < request.K; j++ {
					if d := distFunc(centroids[j], point); d < minDist {
						minDist = d
						closest = j
					}
				}

				if assignments[i] != closest {
					assignments[i] = closest
					changed = true
				}
				inertia += minDist
			}

			if !changed {
				break
			}

			for i := range counts {
				counts[i] = 0
				for j := range sums[i] {
					sums[i][j] = 0
				}
			}

			for i, point := range request.Points {
				cluster := assignments[i]
				counts[cluster]++
				for j, val := range point.Values {
					sums[cluster][j] += val
				}
			}

			for i := range centroids {
				if counts[i] > 0 {
					for j := range centroids[i].Values {
						centroids[i].Values[j] = sums[i][j] / float64(counts[i])
					}
				}
			}
		}

		if inertia < bestInertia {
			bestInertia = inertia
			for i := range centroids {
				copy(bestCentroids[i].Values, centroids[i].Values)
			}
		}
	}

	VisitKMeansUpdateListeners(request.Listeners, bestCentroids)
	return bestCentroids
}
