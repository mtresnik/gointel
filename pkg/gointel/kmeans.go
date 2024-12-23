package gointel

import "github.com/mtresnik/gomath/pkg/gomath"

type KMeansResult []gomath.Point

type KMeansRequest struct {
	K                int
	Points           []gomath.Point
	DistanceFunction *gomath.DistanceFunction
	Listeners        []KMeansUpdateListener
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
	}
}

func KMeans(request KMeansRequest) KMeansResult {
	n := request.K
	points := request.Points

	if n <= 0 || len(points) <= 0 {
		return nil
	}
	centroids := make([]gomath.Point, n)
	for i := range centroids {
		centroids[i] = points[i%len(points)]
	}
	distanceFunction := gomath.EuclideanDistance
	if request.DistanceFunction != nil {
		distanceFunction = *request.DistanceFunction
	}
	assignments := make([]int, len(points))
	for {
		VisitKMeansUpdateListeners(request.Listeners, centroids)
		changed := false

		for i, point := range points {
			closest := 0
			minDist := distanceFunction(centroids[0], point)
			for j, centroid := range centroids {
				distance := distanceFunction(centroid, point)
				if distance < minDist {
					closest = j
					minDist = distance
				}
			}
			if assignments[i] != closest {
				assignments[i] = closest
				changed = true
			}
		}
		if !changed {
			break
		}

		counts := make([]int, n)
		newCentroids := make([]gomath.Point, n)

		for i := range newCentroids {
			newCentroids[i] = *gomath.NewPoint(make([]float64, len(points[0].Values))...)
		}
		for i, point := range points {
			cluster := assignments[i]
			for j := range point.Values {
				newCentroids[cluster].Values[j] += point.Values[j]
			}
			counts[cluster]++
		}
		for i := range newCentroids {
			if counts[i] == 0 {
				continue
			}
			for j := range newCentroids[i].Values {
				newCentroids[i].Values[j] /= float64(counts[i])
			}
		}
		centroids = newCentroids

	}
	return centroids
}
