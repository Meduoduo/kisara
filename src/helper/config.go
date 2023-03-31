package helper

import (
	log "github.com/Yeuoly/kisara/src/routine/log"
	"github.com/spf13/viper"
)

func InitServerConfig() {
	viper.SetConfigName("conf")
	viper.SetConfigType("toml")
	viper.AddConfigPath("./conf/")

	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("Read config file failed: %v", err)
	}

	log.Info("Read config file success")
}

func GetConfigInteger(name string) int {
	return viper.GetInt(name)
}

func GetConfigString(name string) string {
	return viper.GetString(name)
}
