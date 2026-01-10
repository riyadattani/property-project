package internal

import "os"

type Config struct {
	Env  string
	Port string
}

func LoadConfig() Config {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		Env:  env,
		Port: port,
	}
}
