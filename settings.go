package main

import (
	"github.com/jmartin82/mconfig"
)

// SmartThingsConfiguration has settings for controlling the thermostat
type SmartThingsConfiguration struct {
	APIToken             string `yaml:"APIToken"`
	ThermostatCommandURL string `yaml:"ThermostatCommandURL"`
}

// GmailOauth2Configuration has Google API authentication details
type GmailOauth2Configuration struct {
	ClientID     string `yaml:"ClientID"`
	ClientSecret string `yaml:"ClientSecret"`
}

// GmailOauth2TokenConfiguration has Google API token details
type GmailOauth2TokenConfiguration struct {
	AccessToken  string `yaml:"AccessToken"`
	RefreshToken string `yaml:"RefreshToken"`
}

// GmailConfiguration has settings for sending email notifications
type GmailConfiguration struct {
	Oauth2Config  GmailOauth2Configuration      `yaml:"Oauth2Config"`
	Oauth2Token   GmailOauth2TokenConfiguration `yaml:"Oauth2Token"`
	To            string                        `yaml:"To"`
	SubjectPrefix string                        `yaml:"SubjectPrefix"`
}

// Configuration is the main config struct for this app
type Configuration struct {
	SmartThings SmartThingsConfiguration `yaml:"SmartThings"`
	Gmail       GmailConfiguration       `yaml:"Gmail"`
}

var config = Configuration{}

// LoadSettings loads all app configuration settings from config.yaml
func LoadSettings() {
	err := mconfig.Parse("config.yaml", &config)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("%v\n", config)

	// contents, _ := ioutil.ReadFile("config.yaml")
	// j2, err := yaml.YAMLToJSON(contents)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Printf("%v\n", string(j2))
}
