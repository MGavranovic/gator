package config

import (
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_URL            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func gatorConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error finding home dir: %s", err)
	}

	fileName := fmt.Sprintf("%s/%s", homeDir, configFileName)
	return fileName
}

func Read() Config {
	fileName := gatorConfigPath()

	data, err := os.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading file: %s", err)
	}

	var cfg Config

	err = json.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s", err)
	}

	fmt.Println(cfg)
	return cfg
}

func (cfg *Config) SetUser(userName string) error {
	cfg.Current_user_name = userName

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	fileName := gatorConfigPath()
	errWriting := os.WriteFile(fileName, jsonData, 0600)
	if errWriting != nil {
		return errWriting
	}
	return nil
}
