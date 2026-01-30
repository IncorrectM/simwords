package word

import "gorm.io/gorm"

func SaveWords(db *gorm.DB, words []WordEmbedding) error {
	db.AutoMigrate(&WordEmbedding{})

	batchSize := 100
	for i := 0; i < len(words); i += batchSize {
		end := min(i+batchSize, len(words))

		batch := words[i:end]
		if err := db.Save(&batch).Error; err != nil {
			return err
		}

	}

	return nil
}

func BatchUpdateClusterIDs(db *gorm.DB, words []WordEmbedding, clusterIDs []uint) error {
	batchSize := 100
	for i := 0; i < len(words); i += batchSize {
		end := min(i+batchSize, len(words))

		batch := words[i:end]
		for j := range batch {
			batch[j].ClusterID = clusterIDs[i+j]
		}

		if err := db.Save(&batch).Error; err != nil {
			return err
		}
	}
	return nil
}
