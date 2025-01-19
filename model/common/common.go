package common

import (
	"time"

	"gorm.io/gorm"
)

type NODE_MODEL struct {
	ID        uint           `json:"-" gorm:"primarykey"`
	Status    int            `json:"-" gorm:"comment:status"` // 1:ACTIVE 2:INACTIVE
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
