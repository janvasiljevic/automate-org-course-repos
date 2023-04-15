package cmd

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type config struct {
	OrgName            string   `validate:"required"`
	Token              string   `validate:"required"`
	ContentDir         string   `validate:"required"`
	InitReadme         string   `validate:"required"`
	WhiteListedMembers []string `validate:"required"`
	InviteTo           []string `validate:"required"`
}

var Config config

func InitConfig() error {

	vp := viper.New()

	// set config file options
	vp.SetConfigName(".env")
	vp.SetConfigType("toml")
	vp.AddConfigPath(".")

	// Read config file
	if err := vp.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file %v", err)
	}

	// Parse the config file
	if err := vp.Unmarshal(&Config); err != nil {
		return fmt.Errorf("error parsing env file %v", err)
	}

	validate := validator.New()

	// Validate the the config struct
	if err := validate.Struct(&Config); err != nil {
		return fmt.Errorf("missing required attributes %v", err)
	}

	return nil
}
