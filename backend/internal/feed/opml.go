package feed

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// OPMLDocument is the parsed representation of an OPML file.
type OPMLDocument struct {
	Title   string
	Folders []*OPMLFolder
	Feeds   []*OPMLFeed // uncategorized (root-level) feeds
}

// OPMLFolder is a named category containing feeds.
type OPMLFolder struct {
	Name  string
	Feeds []*OPMLFeed
}

// OPMLFeed represents a single RSS/Atom subscription.
type OPMLFeed struct {
	Title   string
	XMLURL  string
	HTMLURL string
}

// --- raw XML types ---

type opmlXML struct {
	XMLName xml.Name    `xml:"opml"`
	Version string      `xml:"version,attr"`
	Head    opmlHead    `xml:"head"`
	Body    opmlBodyXML `xml:"body"`
}

type opmlHead struct {
	Title string `xml:"title"`
}

type opmlBodyXML struct {
	Outlines []opmlOutlineXML `xml:"outline"`
}

type opmlOutlineXML struct {
	Text     string           `xml:"text,attr"`
	Title    string           `xml:"title,attr"`
	Type     string           `xml:"type,attr"`
	XMLURL   string           `xml:"xmlUrl,attr"`
	HTMLURL  string           `xml:"htmlUrl,attr"`
	Children []opmlOutlineXML `xml:"outline"`
}

const maxOPMLBytes = 1 << 20 // 1 MiB

// ParseOPML parses OPML XML bytes into an OPMLDocument.
func ParseOPML(data []byte) (*OPMLDocument, error) {
	if len(data) > maxOPMLBytes {
		return nil, fmt.Errorf("OPML file too large (%d bytes, max %d)", len(data), maxOPMLBytes)
	}
	var raw opmlXML
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parse OPML: %w", err)
	}

	doc := &OPMLDocument{Title: raw.Head.Title}
	for _, outline := range raw.Body.Outlines {
		if isFeed(outline) {
			// Root-level feed (uncategorized).
			doc.Feeds = append(doc.Feeds, outlineToFeed(outline))
		} else {
			// Treat as folder.
			folder := &OPMLFolder{Name: outline.Text}
			if folder.Name == "" {
				folder.Name = outline.Title
			}
			for _, child := range outline.Children {
				if isFeed(child) {
					folder.Feeds = append(folder.Feeds, outlineToFeed(child))
				}
			}
			doc.Folders = append(doc.Folders, folder)
		}
	}
	return doc, nil
}

// isFeed returns true if the outline represents a feed (not a folder).
func isFeed(o opmlOutlineXML) bool {
	return o.Type == "rss" || o.Type == "atom" || o.XMLURL != ""
}

func outlineToFeed(o opmlOutlineXML) *OPMLFeed {
	title := o.Title
	if title == "" {
		title = o.Text
	}
	return &OPMLFeed{Title: title, XMLURL: o.XMLURL, HTMLURL: o.HTMLURL}
}

// GenerateOPML serializes an OPMLDocument to OPML 1.0 XML bytes.
func GenerateOPML(doc *OPMLDocument) ([]byte, error) {
	title := doc.Title
	if title == "" {
		title = "PlexReader Export"
	}

	raw := opmlXML{
		Version: "1.0",
		Head:    opmlHead{Title: title},
	}

	// Folders.
	for _, folder := range doc.Folders {
		folderOutline := opmlOutlineXML{
			Text:  folder.Name,
			Title: folder.Name,
		}
		for _, f := range folder.Feeds {
			folderOutline.Children = append(folderOutline.Children, feedToOutline(f))
		}
		raw.Body.Outlines = append(raw.Body.Outlines, folderOutline)
	}
	// Uncategorized feeds at root.
	for _, f := range doc.Feeds {
		raw.Body.Outlines = append(raw.Body.Outlines, feedToOutline(f))
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(raw); err != nil {
		return nil, fmt.Errorf("generate OPML: %w", err)
	}
	return buf.Bytes(), nil
}

func feedToOutline(f *OPMLFeed) opmlOutlineXML {
	return opmlOutlineXML{
		Text:    f.Title,
		Title:   f.Title,
		Type:    "rss",
		XMLURL:  f.XMLURL,
		HTMLURL: f.HTMLURL,
	}
}
