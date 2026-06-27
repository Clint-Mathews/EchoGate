package utils

import (
	"log"

	"github.com/spf13/viper"
)

func LoadEnv() {
	// 1. Tell Viper the exact file name and type
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	// 2. Register multiple search paths
	viper.AddConfigPath(".")      // Look in the root folder (where go run is executed)
	viper.AddConfigPath("../../") // Fallback: look two folders up (if running from /internal/config/)

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error loading ENV")
	}
}

func SetKey(key string, value any) {
	viper.Set(key, value)
}

func GetRESTPort() int {
	return viper.GetInt("REST_PORT")
}

func GetXInternalToken() string {
	return viper.GetString("X_INTERNAL_TOKEN")
}
