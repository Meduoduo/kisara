package types

import "gorm.io/gorm"

type KisaraContainer struct {
	gorm.Model
	Id            int    `gorm:"primaryKey;autoIncrement;not null"`
	ContainerName string `gorm:"type:varchar(255);not null"`
	ContainerId   string `gorm:"type:varchar(255);not null;unique;index"`
	Labels        string `gorm:"type:varchar(2048);not null"`
	Image         string `gorm:"type:varchar(255);not null"`
	Uid           int    `gorm:"type:int;not null"`
}
