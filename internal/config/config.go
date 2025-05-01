package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	Db_URL          string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read(mainDir string) (Config, error) {
	var config Config
	configFilePath, err := getConfigFilePath(mainDir)
	if err != nil {
		return Config{}, err
	}

	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, err
	}

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return Config{}, err
	}

	defer jsonFile.Close()

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

func (c *Config) SetUser(userName string, mainDir string) error {
	c.CurrentUserName = userName
	err := write(*c, mainDir)
	return err
}

func getConfigFilePath(mainDir string) (string, error) {
	return filepath.Join(mainDir, configFileName), nil
}

func write(cfg Config, mainDir string) error {
	configFilePath, err := getConfigFilePath(mainDir)
	if err != nil {
		return err
	}

	content, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configFilePath, content, 0644)
	if err != nil {
		return err
	}

	return nil
}
