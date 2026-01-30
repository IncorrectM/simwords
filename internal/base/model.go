package base

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type BaseModel struct {
	gorm.Model
	ID uint `gorm:"primaryKey;autoIncrement"`
}

type Embedding struct {
	NormalizedEmbedding Float64Slice `gorm:"type:json"`
	RawEmbedding        Float64Slice `gorm:"type:json"`
}

type Float64Slice []float64

func (f Float64Slice) Value() (driver.Value, error) {
	return json.Marshal(f)
}

func (f *Float64Slice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Float64Slice")
	}
	return json.Unmarshal(bytes, f)
}
