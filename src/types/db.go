package types

import (
	"encoding/json"

	"gorm.io/gorm"
)

type DBContainer struct {
	gorm.Model
	Id            int    `gorm:"primaryKey;autoIncrement;not null"`
	ContainerName string `gorm:"type:varchar(255);not null"`
	ContainerId   string `gorm:"type:varchar(255);not null;unique;index"`
	Labels        string `gorm:"type:varchar(2048);not null"`
	Image         string `gorm:"type:varchar(255);not null"`
	Uid           int    `gorm:"type:int;not null"`
}

type DBService struct {
	gorm.Model
	Id          int    `gorm:"primaryKey;autoIncrement;not null"`
	ServiceName string `gorm:"type:varchar(255);not null"`
	ServiceId   string `gorm:"type:varchar(255);not null;unique;index"`
	Containers  string `gorm:"type:varchar(2048);not null"`
	Networks    string `gorm:"type:varchar(2048);not null"`
	Flags       string `gorm:"type:varchar(2048);not null"`
}

func (c *DBService) GetService() (Service, error) {
	var service Service
	service.Id = c.ServiceId
	service.Name = c.ServiceName

	var containers []Container
	err := json.Unmarshal([]byte(c.Containers), &containers)
	if err != nil {
		return service, err
	}

	var networks []Network
	err = json.Unmarshal([]byte(c.Networks), &networks)
	if err != nil {
		return service, err
	}

	var flags []ServiceFlag
	err = json.Unmarshal([]byte(c.Flags), &flags)
	if err != nil {
		return service, err
	}

	service.Containers = containers
	service.Networks = networks
	service.Flags = flags

	return service, nil
}

func (c *DBService) InjectService(service Service) {
	c.ServiceId = service.Id
	c.ServiceName = service.Name

	containers, _ := json.Marshal(service.Containers)
	c.Containers = string(containers)

	networks, _ := json.Marshal(service.Networks)
	c.Networks = string(networks)

	flags, _ := json.Marshal(service.Flags)
	c.Flags = string(flags)
}
