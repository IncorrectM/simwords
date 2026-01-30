package word

import (
	"math"
	"yggdrasil/sim-words/internal/base"
	"yggdrasil/sim-words/internal/common"
)

func L2Normalize(v base.Float64Slice) base.Float64Slice {
	var sum float64
	for _, f := range v {
		sum += f * f
	}
	norm := math.Sqrt(sum)
	return common.Map(v, func(original float64) float64 {
		return original / norm
	})
}
