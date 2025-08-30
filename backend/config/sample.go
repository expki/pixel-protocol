package config

import (
	"encoding/json"
	"errors"
	"os"
)

// CreateSample creates a sample configuration file.
func CreateSample(path string) error {
	sample := Config{
		Server: ConfigServer{
			HttpAddress:  ":7500",
			HttpsAddress: ":7501",
		},
		TLS: ConfigTLS{
			DomainNameServer: []string{},
			IP:               []string{},
			Certificates:     []*ConfigTLSPath{},
		},
		Database: Database{
			Sqlite:   "./vectorstore.db",
			Cache:    "./vectorcache",
			LogLevel: LogLevelError,
		},
		LogLevel: LogLevelInfo,
	}
	raw, err := json.MarshalIndent(sample, "", "    ")
	if err != nil {
		return errors.Join(errors.New("could not marshal sample config"), err)
	}
	err = os.WriteFile(path, raw, 0600)
	if err != nil {
		return errors.Join(errors.New("could not write sample config file"), err)
	}
	return nil
}
