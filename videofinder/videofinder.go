package videofinder

import (
	"archive/zip"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type VideoBox struct {
	Base64   string
	X, Y     int64
	Cx, Cy   int64
	MimeType string
}

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

func ExtractVideos(r *zip.ReadCloser, slideIndex int) ([]VideoBox, error) {
	data, err := readFileFromZip(r, fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex))
	if err != nil {
		return nil, err
	}

	rels, _ := getSlideRelationships(r, slideIndex)

	decoder := xml.NewDecoder(strings.NewReader(string(data)))
	var videos []VideoBox
	var currentX, currentY, currentCx, currentCy int64
	var inVideoShape bool
	var videoShapeCoords struct {
		x, y, cx, cy int64
	}

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

			if tok.Name.Local == "pic" {
				inVideoShape = true
				videoShapeCoords.x, videoShapeCoords.y, videoShapeCoords.cx, videoShapeCoords.cy = 0, 0, 0, 0
			}

			if tok.Name.Local == "xfrm" {

				currentX, currentY, currentCx, currentCy = 0, 0, 0, 0
			}
			if tok.Name.Local == "off" {
				for _, attr := range tok.Attr {
					if attr.Name.Local == "x" {
						fmt.Sscanf(attr.Value, "%d", &currentX)
						if inVideoShape {
							videoShapeCoords.x = currentX
						}
					}
					if attr.Name.Local == "y" {
						fmt.Sscanf(attr.Value, "%d", &currentY)
						if inVideoShape {
							videoShapeCoords.y = currentY
						}
					}
				}
			}
			if tok.Name.Local == "ext" {
				for _, attr := range tok.Attr {
					if attr.Name.Local == "cx" {
						fmt.Sscanf(attr.Value, "%d", &currentCx)
						if inVideoShape {
							videoShapeCoords.cx = currentCx
						}
					}
					if attr.Name.Local == "cy" {
						fmt.Sscanf(attr.Value, "%d", &currentCy)
						if inVideoShape {
							videoShapeCoords.cy = currentCy
						}
					}
				}
			}

			if tok.Name.Local == "videoFile" {
				var link string
				for _, attr := range tok.Attr {
					if attr.Name.Local == "link" {
						link = attr.Value
					}
				}

				if link != "" {
					target, ok := rels[link]
					if !ok {
						continue
					}

					targetPath := filepath.ToSlash("ppt/" + strings.TrimPrefix(target, "../"))
					videoData, err := readFileFromZip(r, targetPath)
					if err != nil {
						continue
					}
					b64 := base64.StdEncoding.EncodeToString(videoData)

					ext := strings.ToLower(filepath.Ext(targetPath))
					mimeType := "video/mp4"
					if ext == ".avi" {
						mimeType = "video/x-msvideo"
					} else if ext == ".mov" {
						mimeType = "video/quicktime"
					} else if ext == ".wmv" {
						mimeType = "video/x-ms-wmv"
					} else if ext == ".flv" {
						mimeType = "video/x-flv"
					} else if ext == ".webm" {
						mimeType = "video/webm"
					}

					videos = append(videos, VideoBox{
						Base64:   b64,
						X:        videoShapeCoords.x,
						Y:        videoShapeCoords.y,
						Cx:       videoShapeCoords.cx,
						Cy:       videoShapeCoords.cy,
						MimeType: mimeType,
					})
				}
			}

			if tok.Name.Local == "media" {
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

					ext := strings.ToLower(filepath.Ext(target))
					if ext == ".mp4" || ext == ".avi" || ext == ".mov" || ext == ".wmv" || ext == ".flv" || ext == ".webm" {

						targetPath := filepath.ToSlash("ppt/" + strings.TrimPrefix(target, "../"))
						videoData, err := readFileFromZip(r, targetPath)
						if err != nil {
							continue
						}

						b64 := base64.StdEncoding.EncodeToString(videoData)

						mimeType := "video/mp4"
						if ext == ".avi" {
							mimeType = "video/x-msvideo"
						} else if ext == ".mov" {
							mimeType = "video/quicktime"
						} else if ext == ".wmv" {
							mimeType = "video/x-ms-wmv"
						} else if ext == ".flv" {
							mimeType = "video/x-flv"
						} else if ext == ".webm" {
							mimeType = "video/webm"
						}

						videos = append(videos, VideoBox{
							Base64:   b64,
							X:        currentX,
							Y:        currentY,
							Cx:       currentCx,
							Cy:       currentCy,
							MimeType: mimeType,
						})
					}
				}
			}

		case xml.EndElement:

		}
	}

	return videos, nil
}
