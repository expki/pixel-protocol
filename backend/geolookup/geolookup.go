package geolookup

import (
	"log"
	"net"
	"net/http"
	"strings"

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

func GetClientCountry(r *http.Request) string {
	country, err := GeoIP.Country(GetClientIP(r))
	if err != nil || country == nil {
		return "unknown"
	}
	return country.Country.IsoCode
}

func GetClientIP(r *http.Request) net.IP {
	// Check X-Forwarded-For header (most common for proxies/load balancers)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := net.ParseIP(strings.TrimSpace(ips[0]))
			if ip != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header (Nginx proxy)
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		ip := net.ParseIP(strings.TrimSpace(xri))
		if ip != nil {
			return ip
		}
	}

	// Check X-Client-IP header (some proxies)
	xci := r.Header.Get("X-Client-IP")
	if xci != "" {
		ip := net.ParseIP(xci)
		if ip != nil {
			return ip
		}
	}

	// Check CF-Connecting-IP (Cloudflare)
	cfip := r.Header.Get("CF-Connecting-IP")
	if cfip != "" {
		ip := net.ParseIP(cfip)
		if ip != nil {
			return ip
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr // Return as-is if parsing fails
	}

	return net.ParseIP(host)
}
