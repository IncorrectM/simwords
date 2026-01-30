package cluster

import "gorm.io/gorm"

func SaveClusters(db *gorm.DB, clusters []Cluster) error {
	db.AutoMigrate(&Cluster{})
	for i := range clusters {
		db.Create(&clusters[i])
	}

	return nil
}
