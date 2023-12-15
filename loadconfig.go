package main

import (
	"encoding/json"
	"log"
	"os"
)

var (
	config Config
)

type Config struct {
	DB struct {
		Host   string `json:"Host"`
		Port   int    `json:"Port"`
		DBname string `json:"DBname"`
		User   string `json:"User:"`
		Pass   string `json:"Pass"`
	} `json:"DB"`
	ApiKeys struct {
		OpenaiApikey string `json:"OpenaiApikey"`
	} `json:"ApiKeys"`
}

func LoadConfiguration(file string) (Config, error) {
	var config Config
	configFile, err := os.Open(file)
	if err != nil {
		return config, err
	}
	defer func(configFile *os.File) {
		err := configFile.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(configFile)

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return config, err
	}

	return config, nil
}
