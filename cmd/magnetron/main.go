package main

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"magnetron/pkg/config"
	"magnetron/pkg/registry"
	"os"
	"runtime/debug"

	cli "github.com/urfave/cli/v2"
)

var magnetronVersion = "0.1.0"

var (
	//go:embed banner.txt
	banner string
)

var Commit = func() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value
			}
		}
	} else {
		fmt.Println("oops")
	}

	return ""
}()

func main() {

	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "options for configuration",
				Subcommands: []*cli.Command{
					{
						Name:   "init",
						Usage:  "initializes a default configuration",
						Action: initConfig,
					},
					{
						Name:   "validate",
						Usage:  "validates a configuration file",
						Action: validateConfig,
					},
					{
						Name:   "show",
						Usage:  "shows an effective configuration",
						Action: showEffectiveConfig,
					},
				},
			},
			{
				Name:    "serve",
				Aliases: []string{"s"},
				Usage:   "runs the listen server",
				Action:  serve,
			},
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "prints the version",
				Action:  version,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func initConfig(cCtx *cli.Context) error {

	if cCtx.Args().Len() != 1 {
		log.Fatal("Expected config destination file path. e.g. ~/config.yaml")
	}

	configDest := cCtx.Args().First()
	defaultConfig := config.GetDefaultConfig()
	config.WriteConfig(defaultConfig, configDest)

	fmt.Println("Generated default configuration and saved at:", configDest)

	return nil
}

func validateConfig(cCtx *cli.Context) error {

	if cCtx.Args().Len() != 1 {
		log.Fatal("Expected config file path. e.g. ~/config.yaml")
	}

	configPath := cCtx.Args().First()

	c := config.ReadConfigFile(configPath)

	c.Validate()

	log.Println("Validated configuration.")
	return nil
}

func showEffectiveConfig(cCtx *cli.Context) error {
	log.Println("Effective configuration is...")

	if cCtx.Args().Len() != 1 {
		log.Fatal("Expected config file path. e.g. ~/config.yaml")
	}

	configPath := cCtx.Args().First()

	cfg := config.ReadConfigFile(configPath)

	configYaml, err := yaml.Marshal(&cfg)

	if err != nil {
		log.Fatal("Could not marshal config data.", err)
	}

	log.Println("\n\n", string(configYaml))

	return nil
}

func serve(cCtx *cli.Context) error {

	fmt.Println(banner)

	log.Printf("Magnetron Hotline Tracker Version: %s", magnetronVersion)

	if cCtx.Args().Len() != 1 {
		log.Printf("Args: %s", cCtx.Args())
		log.Fatalf("Expected config file path. e.g. ~/config.yaml, %s", cCtx.Args().First())
	}

	configPath := cCtx.Args().First()
	cfg := config.ReadConfigFile(configPath)

	if err := registry.NewRegistry(&cfg); err != nil {
		return err
	} else {
		registry.RegistryInstance.Serve()
	}

	return nil
}

func version(cCtx *cli.Context) error {
	fmt.Println("Version:", magnetronVersion)
	fmt.Println("Commit:", Commit)
	return nil
}
