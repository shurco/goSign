package app

import "github.com/shurco/gosign/internal/config"

// GenConfigFile is ...
func GenConfigFile() error {
	cfg := config.Default()
	if err := config.Save(cfg); err != nil {
		return err
	}
	return nil
}
