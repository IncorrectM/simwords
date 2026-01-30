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

// SelectByClusterID 返回指定簇内的所有单词
func SelectByClusterID(db *gorm.DB, clusterID uint) ([]WordEmbedding, error) {
	var words []WordEmbedding
	err := db.
		Where("cluster_id = ?", clusterID).
		Find(&words).Error
	return words, err
}
