package ai

import (
	"net"
	"net/url"
	"strings"
)

// ClassifyEndpoint decides whether a base URL is local (on-device) or cloud.
// Ambiguous endpoints fail safe to cloud.
func ClassifyEndpoint(rawURL string) (local bool, verdict string) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Host == "" {
		return false, "Cloud · data may leave your device"
	}

	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false, "Cloud · data may leave your device"
	}

	if isLoopback(host) || strings.HasSuffix(host, ".local") {
		return true, "Private · on-device"
	}

	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return true, "Private · on-device"
		}
		return false, "Cloud · data may leave your device"
	}

	if port := u.Port(); port == "11434" || port == "1234" {
		return true, "Private · on-device"
	}

	return false, "Cloud · data may leave your device"
}

func isLoopback(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
