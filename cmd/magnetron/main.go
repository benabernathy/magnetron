package main

import (
	_ "embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"magnetron/internal/config"
	"magnetron/internal/registry"
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
				Name:    "password",
				Aliases: []string{"p"},
				Usage:   "options for password management",
				Subcommands: []*cli.Command{
					{
						Name:   "init",
						Usage:  "initializes a default password configuration",
						Action: initPasswordConfig,
					},
					{
						Name:   "validate",
						Usage:  "validates a password configuration file",
						Action: validatePasswordConfig,
					},
					{
						Name:   "encrypt",
						Usage:  "encrypts a password",
						Action: encryptPassword,
					},
					{
						Name:   "check",
						Usage:  "checks a password against the supplied password file",
						Action: checkPassword,
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

	log.Printf("Magnetron Hotline Tracker %s, %s", magnetronVersion, Commit)

	if cCtx.Args().Len() != 1 {
		log.Printf("Args: %s", cCtx.Args())
		log.Fatalf("Expected config file path. e.g. ~/config.yaml, %s", cCtx.Args().First())
	}

	configPath := cCtx.Args().First()
	cfg := config.ReadConfigFile(configPath)

	var passwordCfg *config.PasswordConfig
	if cfg.EnablePasswords {
		passwordCfg = config.ReadPasswordConfigFile(cfg.PasswordFile)
	}

	if err := registry.NewRegistry(&cfg, passwordCfg); err != nil {
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

func encryptPassword(cCtx *cli.Context) error {

	password, err := registry.PromptUserForPassword()

	if err != nil {
		fmt.Println("Error while prompting for password:", err)
	}

	hashedPassword, err := registry.EncryptPassword(password)

	if err != nil {
		fmt.Println("Error while encrypting password:", err)
	}

	fmt.Println("\n\n" + hashedPassword)

	return nil
}

func checkPassword(cCtx *cli.Context) error {

	if cCtx.Args().Len() != 1 {
		log.Fatal("Expected config file path. e.g. ~/passwords.yaml")
	}

	passwordConfig := config.ReadPasswordConfigFile(cCtx.Args().First())

	password, err := registry.PromptUserForPassword()

	if err != nil {
		fmt.Println("Error while prompting for password:", err)
	}

	if registry.CheckPassword(password, *passwordConfig) {
		fmt.Println("\nPassword is valid.")
	} else {
		fmt.Println("\nPassword is invalid.")
	}

	return nil
}

func initPasswordConfig(cCtx *cli.Context) error {

	if cCtx.Args().Len() != 1 {
		log.Fatal("Expected config destination file path. e.g. ~/passwords.yaml")
	}

	configDest := cCtx.Args().First()
	defaultConfig := config.GetDefaultPasswordConfig()
	config.WritePasswordConfig(*defaultConfig, configDest)

	fmt.Println("Generated default password configuration and saved at:", configDest)

	return nil
}

func validatePasswordConfig(cCtx *cli.Context) error {

	if cCtx.Args().Len() != 1 {
		log.Fatal("Expected config file path. e.g. ~/passwords.yaml")
	}

	configPath := cCtx.Args().First()

	c := config.ReadPasswordConfigFile(configPath)

	c.Validate()

	log.Println("Validated password configuration.")
	return nil
}
