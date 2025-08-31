package bgfinder

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// Slide XML structs
type Slide struct {
	Bg *Bg `xml:"cSld>bg"`
}

type Bg struct {
	BgPr *BgPr `xml:"bgPr"`
}

type BgPr struct {
	BlipFill *BlipFill `xml:"blipFill"`
}

type BlipFill struct {
	Blip Blip `xml:"blip"`
}

type Blip struct {
	Embed string `xml:"embed,attr"` // r:embed
}

type Relationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
	Type   string `xml:"Type,attr"`
}

type Relationships struct {
	Rel []Relationship `xml:"Relationship"`
}

// FindBackground recursively searches for a background image starting at a slide/layout/master
func FindBackground(r *zip.ReadCloser, xmlPath, relsPath string, depth int) (string, string, error) {
	if depth > 5 {
		return "", "", fmt.Errorf("recursion too deep, stopping")
	}

	// get XML file
	xmlData, err := readFileFromZip(r, xmlPath)
	if err != nil {
		return "", "", err
	}

	// unmarshal for<bg>
	var slide Slide
	_ = xml.Unmarshal(xmlData, &slide)

	if slide.Bg != nil && slide.Bg.BgPr != nil && slide.Bg.BgPr.BlipFill != nil {
		embedID := slide.Bg.BgPr.BlipFill.Blip.Embed
		// read relationships
		relsData, _ := readFileFromZip(r, relsPath)
		if relsData != nil {
			var rels Relationships
			if err := xml.Unmarshal(relsData, &rels); err == nil {
				for _, rel := range rels.Rel {
					if rel.ID == embedID && strings.Contains(rel.Type, "image") {
						mediaPath := filepath.ToSlash(filepath.Join(filepath.Dir(relsPath), rel.Target))
						return embedID, mediaPath, nil
					}
				}
			}
		}
	}

	// If no bg, try to follow layout/master
	relsData, _ := readFileFromZip(r, relsPath)
	if relsData == nil {
		return "", "", fmt.Errorf("no relationships for %s", xmlPath)
	}

	var rels Relationships
	if err := xml.Unmarshal(relsData, &rels); err != nil {
		return "", "", err
	}

	for _, rel := range rels.Rel {
		if strings.Contains(rel.Type, "slideLayout") {
			nextXML := filepath.ToSlash(filepath.Join(filepath.Dir(xmlPath), rel.Target))
			nextRels := strings.Replace(nextXML, "slideLayouts/", "slideLayouts/_rels/", 1) + ".rels"
			return FindBackground(r, nextXML, nextRels, depth+1)
		}
		if strings.Contains(rel.Type, "slideMaster") {
			nextXML := filepath.ToSlash(filepath.Join(filepath.Dir(xmlPath), rel.Target))
			nextRels := strings.Replace(nextXML, "slideMasters/", "slideMasters/_rels/", 1) + ".rels"
			return FindBackground(r, nextXML, nextRels, depth+1)
		}
	}

	return "", "", fmt.Errorf("no background found in %s", xmlPath)
}

// helper to read from zip
func readFileFromZip(r *zip.ReadCloser, name string) ([]byte, error) {
	for _, f := range r.File {
		if filepath.ToSlash(f.Name) == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, nil
}
