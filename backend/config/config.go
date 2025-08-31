package config

import (
	"encoding/json"
	"errors"
)

// ParseConfig parses the raw JSON configuration.
func ParseConfig(raw []byte) (config Config, err error) {
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return config, errors.Join(errors.New("unmarshal config"), err)
	}
	return config, nil
}

type Config struct {
	Server   ConfigServer `json:"server"`
	TLS      ConfigTLS    `json:"tls"`
	Database Database     `json:"database"`
	LogLevel LogLevel     `json:"log_level"`
	Claude   ConfigClaude `json:"claude"`
}

type ConfigServer struct {
	HttpAddress  string `json:"http_address"`
	HttpsAddress string `json:"https_address"`
	Http3Address string `json:"http3_address"`
}

type ConfigClaude struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}
