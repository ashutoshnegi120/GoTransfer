package config

import (
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type HttpClient struct {
	Port string `yaml:"port" env-required:"true"`
	Host string `yaml:"host" env-required:"true"`
}

type Config struct {
	Environment      string     `yaml:"environment" env-required:"true"`
	DatabaseLocation string     `yaml:"db_location" env-required:"true"`
	HttpClient       HttpClient `yaml:"http_client" env-required:"true"`
	SecretKey        string     `yaml:"secretKey" env-required:"true"`
}

func MustConfig() *Config {

	var configPath string

	//configPath = os.Getenv("CONFIG_PATH")
	//configPath = "../../.config/config.yml"

	if configPath == "" {
		flags := flag.String("config", "", "plz provide config location to work well if you not define it in .env")
		flag.Parse()
		configPath = *flags
		println("config : ", configPath)
		if configPath == "" {
			log.Fatal("config is not define, plz provide valid config either using flag or .env")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatal("config dir have no config file !!!!!!!!!!!")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("Error parsing YAML config: %v", err)
	}

	return &cfg

}
