package common

func Map[K any, V any](items []K, mapper func(K) V) []V {
	result := make([]V, len(items))
	for i, item := range items {
		result[i] = mapper(item)
	}
	return result
}

func Filter[K any](items []K, filter func(K) bool) []K {
	result := []K{}
	for _, item := range items {
		if filter(item) {
			result = append(result, item)
		}
	}
	return result
}
