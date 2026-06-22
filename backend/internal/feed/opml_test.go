package feed_test

import (
	"os"
	"testing"

	"github.com/plexreader/plexreader/backend/internal/feed"
)

func TestParseOPML(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.opml")
	if err != nil {
		t.Fatalf("read sample.opml: %v", err)
	}
	doc, err := feed.ParseOPML(data)
	if err != nil {
		t.Fatalf("ParseOPML: %v", err)
	}
	if len(doc.Folders) != 4 {
		t.Errorf("expected 4 folders, got %d", len(doc.Folders))
	}
	// Find Technology folder.
	var tech *feed.OPMLFolder
	for _, f := range doc.Folders {
		if f.Name == "Technology" {
			tech = f
			break
		}
	}
	if tech == nil {
		t.Fatal("Technology folder not found")
	}
	if len(tech.Feeds) != 5 {
		t.Errorf("Technology: expected 5 feeds, got %d", len(tech.Feeds))
	}
	// Verify well-known URLs are present.
	var hnFound bool
	for _, f := range tech.Feeds {
		if f.XMLURL == "https://news.ycombinator.com/rss" {
			hnFound = true
		}
	}
	if !hnFound {
		t.Error("Hacker News feed not found in Technology folder")
	}
}

func TestOPMLRoundTrip(t *testing.T) {
	data, _ := os.ReadFile("testdata/sample.opml")
	doc, err := feed.ParseOPML(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	generated, err := feed.GenerateOPML(doc)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	// Re-parse the generated OPML and verify structure matches.
	doc2, err := feed.ParseOPML(generated)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(doc.Folders) != len(doc2.Folders) {
		t.Errorf("folder count mismatch: %d vs %d", len(doc.Folders), len(doc2.Folders))
	}
	for i, f1 := range doc.Folders {
		f2 := doc2.Folders[i]
		if f1.Name != f2.Name {
			t.Errorf("folder[%d] name: %q vs %q", i, f1.Name, f2.Name)
		}
		if len(f1.Feeds) != len(f2.Feeds) {
			t.Errorf("folder[%d] %q feed count: %d vs %d", i, f1.Name, len(f1.Feeds), len(f2.Feeds))
		}
	}
}
