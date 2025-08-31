//go:build !sqlite

package config

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Connection         SingleOrSlice[string] `json:"connection"`
	ConnectionReadOnly SingleOrSlice[string] `json:"connection_readonly"`
	LogLevel           LogLevel              `json:"log_level"` // 0: Silent, 1: Error, 2: Warn, 3: Info, 4: Debug
}

func (c Database) GetDialectors() (readwrite, readonly []gorm.Dialector) {
	readwrite = make([]gorm.Dialector, 0, len(c.Connection))
	for _, dsn := range c.Connection {
		if dsn == "" {
			continue
		}
		readwrite = append(readwrite, postgres.Open(dsn))
	}
	readonly = make([]gorm.Dialector, 0, len(c.ConnectionReadOnly))
	for _, dsn := range c.ConnectionReadOnly {
		if dsn == "" {
			continue
		}
		readonly = append(readonly, postgres.Open(dsn))
	}
	return readwrite, readonly
}

var sampleDatabase = Database{
	Connection: []string{"postgres://username:password@host:5432/database?sslmode=disable"},
}
