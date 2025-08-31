package imagefinder

import (
	"archive/zip"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type ImageBox struct {
	Base64 string
	X, Y   int64
	Cx, Cy int64
}

// ---------------- Read file from zip ----------------
func readFileFromZip(r *zip.ReadCloser, path string) ([]byte, error) {
	path = filepath.ToSlash(path)
	for _, f := range r.File {
		if filepath.ToSlash(f.Name) == path {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("not found: %s", path)
}

// ---------------- Slide relationships ----------------
type Relationship struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
	Type   string `xml:"Type,attr"`
}

type Relationships struct {
	Relationship []Relationship `xml:"Relationship"`
}

func getSlideRelationships(r *zip.ReadCloser, slideIndex int) (map[string]string, error) {
	relsPath := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideIndex)
	data, err := readFileFromZip(r, relsPath)
	if err != nil {
		return nil, err
	}

	var rels Relationships
	if err := xml.Unmarshal(data, &rels); err != nil {
		return nil, err
	}

	m := make(map[string]string)
	for _, rel := range rels.Relationship {
		m[rel.ID] = rel.Target
	}
	return m, nil
}

// ---------------- Extract images ----------------
func ExtractImages(r *zip.ReadCloser, slideIndex int) ([]ImageBox, error) {
	data, err := readFileFromZip(r, fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex))
	if err != nil {
		return nil, err
	}

	rels, _ := getSlideRelationships(r, slideIndex)

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var images []ImageBox
	var currentX, currentY, currentCx, currentCy int64

	for {
		t, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch tok := t.(type) {
		case xml.StartElement:
			// Track coordinates
			if tok.Name.Local == "off" {
				for _, attr := range tok.Attr {
					if attr.Name.Local == "x" {
						fmt.Sscanf(attr.Value, "%d", &currentX)
					}
					if attr.Name.Local == "y" {
						fmt.Sscanf(attr.Value, "%d", &currentY)
					}
				}
			}
			if tok.Name.Local == "ext" {
				for _, attr := range tok.Attr {
					if attr.Name.Local == "cx" {
						fmt.Sscanf(attr.Value, "%d", &currentCx)
					}
					if attr.Name.Local == "cy" {
						fmt.Sscanf(attr.Value, "%d", &currentCy)
					}
				}
			}

			// Detect blip / image
			if tok.Name.Local == "blip" {
				var embed string
				for _, attr := range tok.Attr {
					if attr.Name.Local == "embed" {
						embed = attr.Value
					}
				}

				if embed != "" {
					target, ok := rels[embed]
					if !ok {
						continue
					}

					// Target path might be like "../media/image1.png"
					// Remove leading "../"
					targetPath := filepath.ToSlash("ppt/" + strings.TrimPrefix(target, "../"))
					imgData, err := readFileFromZip(r, targetPath)
					if err != nil {
						continue
					}

					b64 := base64.StdEncoding.EncodeToString(imgData)

					images = append(images, ImageBox{
						Base64: b64,
						X:      currentX,
						Y:      currentY,
						Cx:     currentCx,
						Cy:     currentCy,
					})
				}
			}

		case xml.EndElement:
			// nothing to do
		}
	}

	return images, nil
}

// // // import (
// // // 	"archive/zip"
// // // 	"encoding/xml"
// // // 	"fmt"
// // // 	"io"
// // // 	"path/filepath"
// // // 	"strings"
// // // )

// // // type ImageBox struct {
// // // 	Path   string
// // // 	X, Y   int64
// // // 	Cx, Cy int64
// // // }

// // // // ---------------- Read file from zip ----------------
// // // func readFileFromZip(r *zip.ReadCloser, path string) ([]byte, error) {
// // // 	path = filepath.ToSlash(path)
// // // 	for _, f := range r.File {
// // // 		if filepath.ToSlash(f.Name) == path {
// // // 			rc, err := f.Open()
// // // 			if err != nil {
// // // 				return nil, err
// // // 			}
// // // 			defer rc.Close()
// // // 			return io.ReadAll(rc)
// // // 		}
// // // 	}
// // // 	return nil, fmt.Errorf("not found: %s", path)
// // // }

// // // // ---------------- Extract images ----------------
// // // func ExtractImages(r *zip.ReadCloser, slideIndex int) ([]ImageBox, error) {
// // // 	data, err := readFileFromZip(r, fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex))
// // // 	if err != nil {
// // // 		return nil, err
// // // 	}

// // // 	decoder := xml.NewDecoder(strings.NewReader(string(data)))
// // // 	var images []ImageBox
// // // 	var stack []string // track path for offsets
// // // 	var currentX, currentY, currentCx, currentCy int64

// // // 	for {
// // // 		t, err := decoder.Token()
// // // 		if err != nil {
// // // 			if err == io.EOF {
// // // 				break
// // // 			}
// // // 			return nil, err
// // // 		}

// // // 		switch tok := t.(type) {
// // // 		case xml.StartElement:
// // // 			stack = append(stack, tok.Name.Local)

// // // 			// Track coordinates
// // // 			if tok.Name.Local == "off" {
// // // 				for _, attr := range tok.Attr {
// // // 					if attr.Name.Local == "x" {
// // // 						fmt.Sscanf(attr.Value, "%d", &currentX)
// // // 					}
// // // 					if attr.Name.Local == "y" {
// // // 						fmt.Sscanf(attr.Value, "%d", &currentY)
// // // 					}
// // // 				}
// // // 			}
// // // 			if tok.Name.Local == "ext" {
// // // 				for _, attr := range tok.Attr {
// // // 					if attr.Name.Local == "cx" {
// // // 						fmt.Sscanf(attr.Value, "%d", &currentCx)
// // // 					}
// // // 					if attr.Name.Local == "cy" {
// // // 						fmt.Sscanf(attr.Value, "%d", &currentCy)
// // // 					}
// // // 				}
// // // 			}

// // // 			// Detect blip / image
// // // 			if tok.Name.Local == "blip" {
// // // 				var embed, link string
// // // 				for _, attr := range tok.Attr {
// // // 					if attr.Name.Local == "embed" {
// // // 						embed = attr.Value
// // // 					}
// // // 					if attr.Name.Local == "link" {
// // // 						link = attr.Value
// // // 					}
// // // 				}
// // // 				if embed != "" || link != "" {
// // // 					images = append(images, ImageBox{
// // // 						Path: embed,
// // // 						X:    currentX,
// // // 						Y:    currentY,
// // // 						Cx:   currentCx,
// // // 						Cy:   currentCy,
// // // 					})
// // // 				}
// // // 			}

// // // 		case xml.EndElement:
// // // 			if len(stack) > 0 {
// // // 				stack = stack[:len(stack)-1]
// // // 			}
// // // 		}
// // // 	}

// // // 	fmt.Println("Extracted images:", images)
// // // 	return images, nil
// // // }
