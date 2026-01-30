package word

import "yggdrasil/sim-words/internal/base"

type RawRecord struct {
	Index     int
	Word      string
	Frequency int
}

type WordEmbedding struct {
	base.BaseModel
	ClusterID uint
	Word      string
	Frequency int
	base.Embedding
}
