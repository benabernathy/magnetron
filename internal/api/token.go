package api

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"time"
)

type TokenEntry struct {
	Name        string    `yaml:"Name"`
	Description string    `yaml:"description"`
	Token       string    `yaml:"Token"`
	Expiry      time.Time `yaml:"Expiry"`
}

var (
	//go:embed default_tokens.yml
	defaultTokensResource string
)

type TokenConfig struct {
	TokenEntries []TokenEntry `yaml:"TokenEntries"`
}

func GetDefaultTokenConfig() *TokenConfig {
	var cfg TokenConfig

	err := yaml.Unmarshal([]byte(defaultTokensResource), &cfg)
	if err != nil {
		log.Fatal("error while getting default token configuration.", err)
	}

	return &cfg
}

func ReadTokenConfigFile(path string) *TokenConfig {
	var cfg TokenConfig

	configYaml, err := os.ReadFile(path)

	if err != nil {
		log.Fatal("Could not load yaml token config file: ", path, err)
	}

	err = yaml.Unmarshal(configYaml, &cfg)
	if err != nil {
		log.Fatal("Error while getting parsing yaml token configuration", err)
	}
	return &cfg
}

func (c *TokenConfig) Validate() {
	var errors []error

	if len(c.TokenEntries) == 0 {
		errors = append(errors, fmt.Errorf("token configuration is missing token entries"))
	}

	for _, entry := range c.TokenEntries {
		if entry.Token == "" {
			errors = append(errors, fmt.Errorf("token entry is missing token (%s)", entry.Name))
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}

		log.Fatal("Errors were found while validating the token configuration")
	}
}

func WriteTokenConfig(config TokenConfig, path string) {
	configYaml, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal("Could not marshal token config data.", err)
	}

	err = os.WriteFile(path, configYaml, 0644)

	if err != nil {
		log.Fatal("Could not write token config yaml to file:", path, err)
	}
}
