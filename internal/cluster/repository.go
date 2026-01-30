package cluster

import (
	"fmt"

	"gorm.io/gorm"
)

func SaveClusters(db *gorm.DB, clusters []Cluster) error {
	db.AutoMigrate(&Cluster{})
	for i := range clusters {
		db.Create(&clusters[i])
	}

	return nil
}

func UpdateClusters(db *gorm.DB, clusters []Cluster) error {
	// 一般 AutoMigrate 只需启动时执行一次
	db.AutoMigrate(&Cluster{})

	if len(clusters) == 0 {
		return nil
	}

	// 批量保存
	if err := db.Save(&clusters).Error; err != nil {
		return fmt.Errorf("failed to update clusters: %w", err)
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
