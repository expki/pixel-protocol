//go:build sqlite

package config

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	Connection SingleOrSlice[string] `json:"connection"`
	LogLevel   LogLevel              `json:"log_level"` // 0: Silent, 1: Error, 2: Warn, 3: Info, 4: Debug
}

func (c Database) GetDialectors() (readwrite, readonly []gorm.Dialector) {
	readwrite = make([]gorm.Dialector, 0, len(c.Connection))
	for _, dsn := range c.Connection {
		if dsn == "" {
			continue
		}
		readwrite = append(readwrite, sqlite.Open(dsn))
		break
	}
	return readwrite, nil
}

var sampleDatabase = Database{
	Connection: []string{":memory:"},
}
