package config

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	//go:embed default_config.yml
	defaultConfigResource string
)

type Config struct {
	ClientHost       string        `yaml:"ClientHost" validate:"required"`       // Interface/host and port the clients will connect to the tracker on
	ServerHost       string        `yaml:"ServerHost" validate:"required"`       // Interface/host and port the servers will connect to the tracker on
	ServerExpiration time.Duration `yaml:"ServerExpiration" validate:"required"` // How long a server can be inactive before it is removed from the list
	StaticEntries    []StaticEntry `yaml:"StaticEntries"`                        // Static entries are placed in order at the top of the server list
}

type StaticEntry struct {
	Name        string `yaml:"Name"`
	Description string `yaml:"Description"`
	Address     string `yaml:"Address"`
}

func LoadConfig(configPath string) (*Config, error) {

	config := &Config{}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	d.KnownFields(true)

	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func getHost(address string) (string, error) {
	parts := strings.Split(address, ":")

	if len(parts) >= 1 {
		return parts[0], nil
	} else {
		return address, nil
	}
}

func getPort(address string, defaultPort uint16) (uint16, error) {
	parts := strings.Split(address, ":")

	if len(parts) <= 1 {
		return defaultPort, nil
	} else {
		if p, err := strconv.Atoi(parts[len(parts)-1]); err != nil {
			return 0, err
		} else {
			if p <= 0 {
				return 0, fmt.Errorf("port value must be greater than or equal to zero: %d", p)
			}
			return uint16(p), nil
		}
	}
}

func (c *Config) GetHost() (string, error) {
	return getHost(c.ClientHost)
}

func (c *Config) GetPort() (uint16, error) {
	return getPort(c.ClientHost, 5998)
}

func (c *Config) Validate() {
	var errors []error

	if _, err := c.GetHost(); err != nil {
		errors = append(errors, err)
	}

	if _, err := c.GetPort(); err != nil {
		errors = append(errors, err)
	}

	for _, entry := range c.StaticEntries {

		for _, entryError := range entry.Validate() {
			errors = append(errors, entryError)
		}
	}

	if len(errors) > 0 {
		for _, err := range errors {
			log.Println(err)
		}

		log.Fatal("Errors were found while validating the configuration")
	}
}

func (e *StaticEntry) GetHost() (string, error) {
	return getHost(e.Address)
}

func (e *StaticEntry) GetPort() (uint16, error) {
	return getPort(e.Address, 5598)
}

func (e *StaticEntry) Validate() []error {
	var errors []error

	if _, err := e.GetHost(); err != nil {
		errors = append(errors, err)
	}

	if _, err := e.GetPort(); err != nil {
		errors = append(errors, err)
	}

	return errors
}

func GetDefaultConfig() Config {

	var c Config

	err := yaml.Unmarshal([]byte(defaultConfigResource), &c)
	if err != nil {
		log.Fatal("Error while getting default configuration", err)
	}
	return c

}

func ReadConfigFile(path string) Config {
	var config Config

	configYaml, err := os.ReadFile(path)

	if err != nil {
		log.Fatal("Could not load yaml config file: ", path, err)
	}

	err = yaml.Unmarshal(configYaml, &config)
	if err != nil {
		log.Fatal("Error while getting parsing yaml configuration", err)
	}
	return config
}

func WriteConfig(config Config, path string) {
	configYaml, err := yaml.Marshal(&config)
	if err != nil {
		log.Fatal("Could not marshal config data.", err)
	}

	err = os.WriteFile(path, configYaml, 0644)

	if err != nil {
		log.Fatal("Could not write config yaml to file:", path, err)
	}
}
