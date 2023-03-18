package model

import "time"
import "gorm.io/gorm"

type BaseModel struct {
	ID         uint           `gorm:"primarykey" json:"id,omitempty"`
	CreateTime time.Time      `gorm:"autoCreateTime" json:"-"`
	UpdateTime time.Time      `gorm:"autoUpdateTime" json:"-"`
	DeleteTime gorm.DeletedAt `gorm:"index" json:"-"`
}

// `json:",inline"`