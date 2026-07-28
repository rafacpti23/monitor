package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Host      string     `json:"host"`
	Port      int        `json:"port"`
	JWTSecret string     `json:"jwt_secret"`
	DBPath    string     `json:"db_path"`
	PublicURL string     `json:"public_url"`
	SMTP      SMTPConfig `json:"smtp"`
	WhatsApp  WAConfig   `json:"whatsapp"`
}

type SMTPConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pass string `json:"pass"`
	From string `json:"from"`
}

type WAConfig struct {
	APIURL string `json:"api_url"`
	APIKey string `json:"api_key"`
}

var Cfg Config

// Load loads the configuration from a JSON file.
// If it fails, it will set fallback defaults.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		setDefaults()
		return err
	}
	err = json.Unmarshal(data, &Cfg)
	if err != nil {
		setDefaults()
		return err
	}
	if Cfg.JWTSecret == "" {
		Cfg.JWTSecret = "default_unsafe_secret"
	}
	if Cfg.DBPath == "" {
		Cfg.DBPath = "./p-mon.db"
	}
	if Cfg.Port == 0 {
		Cfg.Port = 8080
	}
	return nil
}

func setDefaults() {
	Cfg = Config{
		Host:      "0.0.0.0",
		Port:      8080,
		JWTSecret: "default_unsafe_secret",
		DBPath:    "./p-mon.db",
	}
}
