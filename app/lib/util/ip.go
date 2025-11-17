package util

import (
	"net"
	"net/http"
	"strings"
)

// ClientIP returns request real ip.
// By default, it returns the remote address of the connection to prevent IP spoofing.
// Only when trusted proxies are provided, it will check X-Real-IP and X-Forwarded-For headers.
func ClientIP(r *http.Request, trustedProxies []string) string {
	// First, try to get IP from RemoteAddr which is the most reliable source
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	if net.ParseIP(ip) == nil {
		return ""
	}

	// If no trusted proxies are provided, return the remote IP directly
	if len(trustedProxies) == 0 {
		return ip
	}

	// Check if the remote IP is from a trusted proxy
	isTrustedProxy := false
	for _, trustedProxy := range trustedProxies {
		if ip == trustedProxy || strings.HasPrefix(ip, trustedProxy+"/") {
			isTrustedProxy = true
			break
		}
	}

	// If the request is not from a trusted proxy, return the remote IP directly
	if !isTrustedProxy {
		return ip
	}

	// If the request is from a trusted proxy, check the headers
	xRealIP := r.Header.Get("X-Real-IP")
	if net.ParseIP(xRealIP) != nil {
		return xRealIP
	}

	xFwdFor := r.Header.Get("X-Forwarded-For")
	if xFwdFor != "" {
		// X-Forwarded-For can contain multiple IPs, use the first one (client IP)
		parts := strings.Split(xFwdFor, ",")
		if len(parts) > 0 {
			clientIP := strings.TrimSpace(parts[0])
			if net.ParseIP(clientIP) != nil {
				return clientIP
			}
		}
	}

	return ip
}

// ClientIPSimple returns request real ip without checking trusted proxies.
// This function is for backward compatibility and direct IP access.
func ClientIPSimple(r *http.Request) string {
	// First, try to get IP from RemoteAddr which is the most reliable source
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	if net.ParseIP(ip) != nil {
		return ip
	}

	return ""
}
