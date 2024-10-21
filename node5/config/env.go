package config

import (
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadEnv loads environment variables from a .env file
func LoadEnv() error {
	err := godotenv.Load(filepath.Join(".", ".env"))
	if err != nil {
		return fmt.Errorf("error loading .env file: %v", err)
	}
	return nil
}
