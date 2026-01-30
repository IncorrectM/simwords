package base

import (
	"fmt"
)

func Distance(a, b Float64Slice) float64 {
	var sum float64

	if len(a) != len(b) {
		panic(fmt.Sprintf("a(l=%d) and b(l=%d) should have same length\n", len(a), len(b)))
	}

	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}

	return sum
}
