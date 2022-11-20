package main

import (
	"fmt"

	"github.com/spf13/viper"
)

type AppCredentials struct {
	APIKey         string
	APIAccessToken string
	APISecretKey   string
}

func GetCredentials() AppCredentials {
	viper.SetConfigName("credentials") // name of config file (without extension)
	viper.SetConfigType("env")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	err := viper.ReadInConfig()        // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var tempCreds AppCredentials
	tempCreds.APIAccessToken = viper.GetString("API_ACCESS_TOKEN")
	tempCreds.APIKey = viper.GetString("API_KEY")
	tempCreds.APISecretKey = viper.GetString("API_SECRET_KEY")
	return tempCreds
}
func main() {
	var AppCredentials = GetCredentials()
	fmt.Print(AppCredentials)
}
