package config

import (
	"encoding/json"
	"os"
)

const configFileName = "/.gatorconfig.json"

type Config struct {
	Db_url            string `json:"db_url"`
	Current_user_name string `json:"current_user_name"`
}

func getConfigPath() (string, error) {
	homePath, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return homePath + configFileName, nil
}

func Read() (Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var res Config
	err = json.Unmarshal(data, &res)
	if err != nil {
		return Config{}, err
	}

	return res, nil
}

func (c Config) SetUser(userName string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}
	c.Current_user_name = userName
	data, err := json.Marshal(c)
	if err != nil {
		return err
	}
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}
