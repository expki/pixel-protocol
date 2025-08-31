package geolookup

import (
	"log"

	_ "embed"

	"github.com/oschwald/geoip2-golang"
)

//go:embed GeoLite2-Country.mmdb
var database []byte

var GeoIP *geoip2.Reader = func() *geoip2.Reader {
	geoip, err := geoip2.FromBytes(database)
	if err != nil {
		log.Fatalf("open geoip database: %v", err)
	}
	return geoip
}()
