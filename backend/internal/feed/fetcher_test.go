package feed_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/plexreader/plexreader/backend/internal/feed"
)

func TestValidateFeedURL_Scheme(t *testing.T) {
	cases := []struct {
		url     string
		wantErr bool
	}{
		{"https://example.com/feed.rss", false},
		{"http://example.com/feed.rss", false},
		{"ftp://example.com/feed.rss", true},
		{"file:///etc/passwd", true},
		{"javascript:alert(1)", true},
		{"", true},
	}
	for _, tc := range cases {
		err := feed.ValidateFeedURL(tc.url)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateFeedURL(%q) err=%v wantErr=%v", tc.url, err, tc.wantErr)
		}
	}
}

func TestValidateFeedURL_PrivateHosts(t *testing.T) {
	// Hostnames that resolve to private addresses or are localhost must be blocked.
	blocked := []string{
		"http://localhost/feed",
		"http://localhost.internal/feed",
	}
	for _, u := range blocked {
		if err := feed.ValidateFeedURL(u); err == nil {
			t.Errorf("expected ValidateFeedURL(%q) to fail, got nil", u)
		}
	}
}

func TestFetcher_FetchPublic(t *testing.T) {
	// Use a local httptest server — it listens on 127.0.0.1 which is loopback
	// and will be blocked by safeDialContext. We test the blocking itself here.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write([]byte(sampleRSS2))
	}))
	defer srv.Close()

	f := feed.NewFetcher(5 * time.Second)
	// localhost server should be blocked by SSRF protection.
	_, err := f.Fetch(t.Context(), srv.URL+"/feed")
	if err == nil {
		t.Error("expected fetch of localhost server to fail (SSRF protection)")
	}
	if !strings.Contains(err.Error(), "disallowed") && !strings.Contains(err.Error(), "private") &&
		!strings.Contains(err.Error(), "dns") && !strings.Contains(err.Error(), "not allowed") {
		t.Errorf("unexpected error (wanted SSRF-related): %v", err)
	}
}

func TestBluemondaySanitizer(t *testing.T) {
	// Feed content with XSS payload should be sanitized before storage.
	maliciousContent := `<p>Normal text</p><script>alert('xss')</script><img src="x" onerror="alert(1)">`
	rssWithXSS := `<?xml version="1.0"?><rss version="2.0"><channel><title>T</title><link>http://t.com</link><description>d</description>
<item><title>XSS</title><link>http://t.com/xss</link>
<content:encoded xmlns:content="http://purl.org/rss/1.0/modules/content/"><![CDATA[` + maliciousContent + `]]></content:encoded>
<guid>xss-1</guid></item></channel></rss>`

	pf, err := feed.ParseFromReader(strings.NewReader(rssWithXSS))
	if err != nil {
		t.Fatalf("ParseFromReader: %v", err)
	}
	if len(pf.Articles) == 0 {
		t.Fatal("expected articles")
	}
	content := pf.Articles[0].Content
	if strings.Contains(content, "<script>") {
		t.Errorf("script tag not stripped from content: %q", content)
	}
	if strings.Contains(content, "onerror") {
		t.Errorf("onerror attribute not stripped from content: %q", content)
	}
	if !strings.Contains(content, "Normal text") {
		t.Errorf("expected normal text preserved in content: %q", content)
	}
}

func TestOPMLSizeLimit(t *testing.T) {
	// A 1.1MB OPML should be rejected.
	big := make([]byte, 1<<20+1)
	_, err := feed.ParseOPML(big)
	if err == nil {
		t.Error("expected error for oversized OPML, got nil")
	}
	if !strings.Contains(err.Error(), "too large") {
		t.Errorf("unexpected error message: %v", err)
	}
}
