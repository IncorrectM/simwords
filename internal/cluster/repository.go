package cluster

import "gorm.io/gorm"

func SaveClusters(db *gorm.DB, clusters []Cluster) error {
	db.AutoMigrate(&Cluster{})
	for i := range clusters {
		db.Create(&clusters[i])
	}

	return nil
}

func GetClusters(db *gorm.DB, ids []uint) ([]Cluster, error) {
	var clusters []Cluster

	query := db.Model(&Cluster{})

	if len(ids) > 0 {
		query = query.Where("id IN ?", ids)
	}

	if err := query.Find(&clusters).Error; err != nil {
		return nil, err
	}

	return clusters, nil
}
