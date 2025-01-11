package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	DbURL           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() Config {
	path, err := getConfigFilePath()
	if err != nil {
		return Config{}
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return Config{}
	}
	defer file.Close()

	config := Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		fmt.Println("Error decoding file:", err)
		return Config{}
	}

	return config

}

func (c *Config) SetUser(userName string) error {
	path, err := getConfigFilePath()
	if err != nil {
		return err
	}

	c.CurrentUserName = userName

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return nil
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	return nil

}

func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Print("couldn't find homedir\n")
		return "", err
	}

	return homeDir + "/" + configFileName, nil
}
