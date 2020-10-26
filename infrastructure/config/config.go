package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/guregu/null"
)

const configFile = "config.json"

type Configuration struct {
	DatabaseFile null.String `json:"databasefile"`
	JwtToken     string      `json:"jwttoken"`
	ApiAddr      null.String `json:"apiaddr"`
	UseSwagger   bool        `json:"useswagger"`
}

var instance *Configuration

func GetConfig() Configuration {
	if instance == nil {
		c := defaultConfiguration()
		f, err := os.Open(configFile)
		if err != nil {
			log.Println("Unable to load configuration")
			return *c
		}
		defer f.Close()
		config, err := ioutil.ReadAll(f)
		if err != nil {
			log.Println("Unable to read configuration")
			return *c
		}
		if err = json.Unmarshal(config, c); err != nil {
			return *defaultConfiguration()
		}
		instance = c
		log.Println("Loaded configuration")
	}
	return *instance
}

func defaultConfiguration() *Configuration {
	return &Configuration{
		DatabaseFile: null.String{},
		JwtToken:     "secret",
		ApiAddr:      null.String{},
		UseSwagger:   false,
	}
}
