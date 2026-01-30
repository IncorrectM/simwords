package common

func Map[K any, V any](items []K, mapper func(K) V) []V {
	result := make([]V, len(items))
	for i, item := range items {
		result[i] = mapper(item)
	}
	return result
}
