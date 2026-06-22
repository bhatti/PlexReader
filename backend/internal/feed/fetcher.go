package feed

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const userAgent = "PlexReader/1.0 (+https://github.com/plexreader/plexreader)"

// HTTPError is returned when a feed server responds with a non-200 status.
// The scheduler uses StatusCode to apply appropriate backoff or removal logic.
type HTTPError struct {
	URL        string
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("fetch %s: HTTP %d", e.URL, e.StatusCode)
}

// blockedCIDRs lists cloud metadata IPs that must never be reachable from
// feed fetches regardless of what DNS resolves to.
var blockedCIDRs []*net.IPNet

func init() {
	for _, cidr := range []string{
		"169.254.169.254/32", // AWS / GCP / Azure instance metadata
		"100.100.100.200/32", // Alibaba Cloud ECS metadata
	} {
		_, ipNet, _ := net.ParseCIDR(cidr)
		if ipNet != nil {
			blockedCIDRs = append(blockedCIDRs, ipNet)
		}
	}
}

func isBlockedCIDR(ip net.IP) bool {
	for _, cidr := range blockedCIDRs {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}

// safeDialContext resolves the target host and validates every returned IP
// at TCP-connect time, not at URL-parse time. This prevents DNS-rebinding:
// an attacker whose DNS server returns a public IP on the first lookup (which
// passes validateFeedURL) but a private IP on the actual dial is blocked here.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("dns lookup failed: %w", err)
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("no addresses for host %s", host)
	}
	for _, a := range addrs {
		ip := a.IP
		// Normalize IPv4-mapped IPv6 (::ffff:10.0.0.1 → 10.0.0.1) so that
		// IsPrivate() and CIDR checks work correctly for both address families.
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		if isPrivateIP(ip) || isBlockedCIDR(ip) {
			return nil, fmt.Errorf("disallowed IP address: %s", a.IP)
		}
	}
	dialer := &net.Dialer{}
	return dialer.DialContext(ctx, network, net.JoinHostPort(addrs[0].IP.String(), port))
}

// CacheEntry stores HTTP caching headers for conditional requests.
type CacheEntry struct {
	ETag         string
	LastModified string
}

// Fetcher fetches and parses RSS/Atom feeds over HTTP, using conditional
// requests to avoid re-downloading unchanged feeds.
type Fetcher struct {
	client *http.Client
	cache  map[string]*CacheEntry
	mu     sync.Mutex
}

// NewFetcher creates a Fetcher with the given HTTP timeout.
// The underlying transport uses safeDialContext to prevent SSRF / DNS rebinding.
func NewFetcher(timeout time.Duration) *Fetcher {
	transport := &http.Transport{
		DialContext: safeDialContext,
	}
	return &Fetcher{
		client: &http.Client{Timeout: timeout, Transport: transport},
		cache:  make(map[string]*CacheEntry),
	}
}

// FetchResult is the outcome of fetching a feed.
type FetchResult struct {
	Feed        *ParsedFeed
	NotModified bool // true when the feed hasn't changed since last fetch
}

// ValidateFeedURL rejects non-HTTP(S) schemes and private/loopback addresses.
// Call this before storing a feed URL to give early feedback; DNS-rebinding
// protection at TCP-connect time is provided by safeDialContext in the Fetcher.
func ValidateFeedURL(rawURL string) error { return validateFeedURL(rawURL) }

// validateFeedURL rejects non-HTTP(S) schemes and private/loopback addresses.
func validateFeedURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme %q not allowed (only http/https)", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("url has no host")
	}
	// Block loopback and link-local hostnames.
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return fmt.Errorf("url host %q is not allowed", host)
	}
	// Note: this pre-flight check guards against obvious cases; DNS rebinding
	// is definitively prevented by safeDialContext at TCP-connect time.
	ips, err := net.LookupHost(host)
	if err != nil {
		// Fail closed: if we can't verify the destination is safe, block it.
		return fmt.Errorf("dns lookup for %q failed: %w", host, err)
	}
	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		// Normalize IPv4-mapped IPv6 before private-range checks.
		if v4 := ip.To4(); v4 != nil {
			ip = v4
		}
		if isPrivateIP(ip) || isBlockedCIDR(ip) {
			return fmt.Errorf("url resolves to disallowed address %s", ipStr)
		}
	}
	return nil
}

// isPrivateIP returns true for loopback, link-local, RFC1918, and IPv6
// unique-local (fc00::/7) addresses — all of which are SSRF targets.
func isPrivateIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	// ip.IsPrivate() covers IPv4 RFC1918 (10/8, 172.16/12, 192.168/16).
	if ip.IsPrivate() {
		return true
	}
	// IPv6 unique-local fc00::/7 is not covered by IsPrivate in all Go versions.
	if ip16 := ip.To16(); ip16 != nil && ip.To4() == nil {
		if ip16[0]&0xfe == 0xfc { // fc00::/7
			return true
		}
	}
	return false
}

// Fetch retrieves and parses the feed at url, using conditional HTTP requests
// when caching headers from a prior fetch are available.
func (f *Fetcher) Fetch(ctx context.Context, url string) (*FetchResult, error) {
	if err := validateFeedURL(url); err != nil {
		return nil, fmt.Errorf("feed url validation: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml, */*")

	f.mu.Lock()
	entry := f.cache[url]
	f.mu.Unlock()

	if entry != nil {
		if entry.ETag != "" {
			req.Header.Set("If-None-Match", entry.ETag)
		}
		if entry.LastModified != "" {
			req.Header.Set("If-Modified-Since", entry.LastModified)
		}
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()
	// Cap feed response at 10 MB to prevent OOM from oversized feeds.
	resp.Body = http.MaxBytesReader(nil, resp.Body, 10<<20)

	if resp.StatusCode == http.StatusNotModified {
		return &FetchResult{NotModified: true}, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{URL: url, StatusCode: resp.StatusCode}
	}

	// Update cache headers.
	f.mu.Lock()
	f.cache[url] = &CacheEntry{
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
	}
	f.mu.Unlock()

	parsed, err := ParseFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	return &FetchResult{Feed: parsed}, nil
}
