package utilities

import "math"

func LogN(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
