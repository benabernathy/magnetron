package config

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var (
	//go:embed default_passwords.yml
	defaultPasswordsResource string
)

type PasswordConfig struct {
	PasswordEntries []PasswordEntry `yaml:"PasswordEntries"`
}

type PasswordEntry struct {
	Name        string `yaml:"Name"`        // A helpful name
	Description string `yaml:"Description"` // A helpful description
	Password    string `yaml:"Password"`    // The hashed/encrypted password
}

func GetDefaultPasswordConfig() *PasswordConfig {
	var cfg PasswordConfig

	err := yaml.Unmarshal([]byte(defaultPasswordsResource), &cfg)
	if err != nil {
		log.Fatal("error while getting default password configuration.", err)
	}

	return &cfg
}

func ReadPasswordConfigFile(path string) *PasswordConfig {
	var cfg PasswordConfig

	configYaml, err := os.ReadFile(path)

	if err != nil {
		log.Fatal("Could not load yaml password config file: ", path, err)
	}

	err = yaml.Unmarshal(configYaml, &cfg)
	if err != nil {
		log.Fatal("Error while getting parsing yaml password configuration", err)
	}
	return &cfg
}

func (c *PasswordConfig) Validate() {
	var errors []error

	if len(c.PasswordEntries) == 0 {
		errors = append(errors, fmt.Errorf("password configuration is missing password entries"))
	}

	for _, entry := range c.PasswordEntries {
		if entry.Password == "" {
			errors = append(errors, fmt.Errorf("password entry is missing password (%s)", entry.Name))
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}

		log.Fatal("Errors were found while validating the password configuration")
	}
}

func WritePasswordConfig(config PasswordConfig, path string) {
	configYaml, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal("Could not marshal password config data.", err)
	}

	err = os.WriteFile(path, configYaml, 0644)

	if err != nil {
		log.Fatal("Could not write password config yaml to file:", path, err)
	}
}
