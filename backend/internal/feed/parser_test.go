package feed_test

import (
	"strings"
	"testing"

	"github.com/plexreader/plexreader/backend/internal/feed"
)

const sampleRSS2 = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>A test feed</description>
    <item>
      <title>Article One</title>
      <link>https://example.com/1</link>
      <description>Summary of article one</description>
      <guid>guid-1</guid>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
      <author>Alice</author>
    </item>
    <item>
      <title>Article Two (no GUID)</title>
      <link>https://example.com/2</link>
      <description>Summary two</description>
    </item>
  </channel>
</rss>`

const sampleAtom = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Atom Test Feed</title>
  <link href="https://atom.example.com"/>
  <id>urn:atom-test</id>
  <entry>
    <title>Atom Entry</title>
    <link href="https://atom.example.com/entry1"/>
    <id>atom-entry-1</id>
    <updated>2024-03-15T10:00:00Z</updated>
    <content type="html">&lt;p&gt;Full content here.&lt;/p&gt;</content>
    <summary>Short summary</summary>
    <author><name>Bob</name></author>
  </entry>
</feed>`

func TestParseRSS2(t *testing.T) {
	pf, err := feed.ParseFromReader(strings.NewReader(sampleRSS2))
	if err != nil {
		t.Fatalf("ParseFromReader: %v", err)
	}
	if pf.Title != "Test Feed" {
		t.Errorf("title: got %q", pf.Title)
	}
	if len(pf.Articles) != 2 {
		t.Fatalf("expected 2 articles, got %d", len(pf.Articles))
	}
	a := pf.Articles[0]
	if a.GUID != "guid-1" {
		t.Errorf("guid: got %q", a.GUID)
	}
	if a.Author != "Alice" {
		t.Errorf("author: got %q", a.Author)
	}
}

func TestParseRSS2NoGUID(t *testing.T) {
	pf, _ := feed.ParseFromReader(strings.NewReader(sampleRSS2))
	a := pf.Articles[1]
	// Article two has no guid — should fallback to link.
	if a.GUID != "https://example.com/2" {
		t.Errorf("expected link as GUID fallback, got %q", a.GUID)
	}
}

func TestParseAtom(t *testing.T) {
	pf, err := feed.ParseFromReader(strings.NewReader(sampleAtom))
	if err != nil {
		t.Fatalf("ParseFromReader atom: %v", err)
	}
	if len(pf.Articles) != 1 {
		t.Fatalf("expected 1 article, got %d", len(pf.Articles))
	}
	a := pf.Articles[0]
	if a.GUID != "atom-entry-1" {
		t.Errorf("atom guid: got %q", a.GUID)
	}
	if a.Author != "Bob" {
		t.Errorf("atom author: got %q", a.Author)
	}
	if !strings.Contains(a.Content, "Full content") {
		t.Errorf("atom content: got %q", a.Content)
	}
}

func TestSummaryTruncation(t *testing.T) {
	// Generate a description longer than 300 chars.
	long := strings.Repeat("word ", 100) // 500 chars
	rss := `<?xml version="1.0"?><rss version="2.0"><channel><title>T</title><link>http://t.com</link><description>d</description>
    <item><title>X</title><link>http://t.com/x</link><description>` + long + `</description><guid>g1</guid></item>
  </channel></rss>`
	pf, _ := feed.ParseFromReader(strings.NewReader(rss))
	if len([]rune(pf.Articles[0].Summary)) > 305 { // slight buffer for ellipsis
		t.Errorf("summary not truncated: %d chars", len(pf.Articles[0].Summary))
	}
}
