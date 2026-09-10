package geo

import "math"

type Point struct {
	Lat float64
	Lon float64
}

const earthRadiusM = 6371000.0

func DistanceMeters(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusM * c
}

func MinDistanceMeters(lat, lon float64, points []Point) float64 {
	if len(points) == 0 {
		return 0
	}
	min := math.MaxFloat64
	for _, p := range points {
		d := DistanceMeters(lat, lon, p.Lat, p.Lon)
		if d < min {
			min = d
		}
	}
	return min
}

// SampleEvery densifies a polyline and returns up to maxPoints samples
// evenly spaced along the full length (always including both ends).
func SampleEvery(points []Point, everyM float64, maxPoints int) []Point {
	if len(points) == 0 {
		return nil
	}
	if len(points) == 1 {
		return []Point{points[0]}
	}
	if maxPoints < 2 {
		maxPoints = 2
	}
	if everyM <= 0 {
		everyM = 100
	}

	dist := make([]float64, len(points))
	for i := 1; i < len(points); i++ {
		dist[i] = dist[i-1] + DistanceMeters(points[i-1].Lat, points[i-1].Lon, points[i].Lat, points[i].Lon)
	}
	total := dist[len(dist)-1]
	if total == 0 {
		return []Point{points[0], points[len(points)-1]}
	}

	n := int(math.Floor(total/everyM)) + 1
	if n < 2 {
		n = 2
	}
	if n > maxPoints {
		n = maxPoints
	}

	out := make([]Point, 0, n)
	for i := 0; i < n; i++ {
		target := total * float64(i) / float64(n-1)
		out = append(out, pointAt(points, dist, target))
	}
	return out
}

func pointAt(points []Point, dist []float64, target float64) Point {
	if target <= 0 {
		return points[0]
	}
	if target >= dist[len(dist)-1] {
		return points[len(points)-1]
	}
	for i := 1; i < len(points); i++ {
		if dist[i] >= target {
			seg := dist[i] - dist[i-1]
			if seg == 0 {
				return points[i]
			}
			t := (target - dist[i-1]) / seg
			return Point{
				Lat: points[i-1].Lat + t*(points[i].Lat-points[i-1].Lat),
				Lon: points[i-1].Lon + t*(points[i].Lon-points[i-1].Lon),
			}
		}
	}
	return points[len(points)-1]
}

func ElevationGainLoss(elevations []float64) (gain, loss float64) {
	for i := 1; i < len(elevations); i++ {
		d := elevations[i] - elevations[i-1]
		if d > 0 {
			gain += d
		} else {
			loss += -d
		}
	}
	return gain, loss
}

func toRadians(v float64) float64 {
	return v * math.Pi / 180
}
