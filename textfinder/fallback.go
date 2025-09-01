package textfinder

import (
	"archive/zip"
	"encoding/xml"
	"path/filepath"
	"strconv"
	"strings"
)

type relEntry struct {
	ID     string `xml:"Id,attr"`
	Target string `xml:"Target,attr"`
	Type   string `xml:"Type,attr"`
}

type relFile struct {
	Rel []relEntry `xml:"Relationship"`
}

type listStyle struct {
	Lvl1pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl1pPr"`
	Lvl2pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl2pPr"`
	Lvl3pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl3pPr"`
	Lvl4pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl4pPr"`
	Lvl5pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl5pPr"`
	Lvl6pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl6pPr"`
	Lvl7pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl7pPr"`
	Lvl8pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl8pPr"`
	Lvl9pPr *struct{ DefRPr *rPr `xml:"defRPr"` } `xml:"lvl9pPr"`
}

type txStylesDoc struct {
	TxStyles *struct {
		TitleStyle *listStyle `xml:"titleStyle"`
		BodyStyle  *listStyle `xml:"bodyStyle"`
		OtherStyle *listStyle `xml:"otherStyle"`
	} `xml:"txStyles"`
}

func readRels(r *zip.ReadCloser, relsPath string) (*relFile, error) {
	data, err := readFileFromZip(r, relsPath)
	if err != nil {
		return nil, err
	}
	var rf relFile
	if err := xml.Unmarshal(data, &rf); err != nil {
		return nil, err
	}
	return &rf, nil
}

func resolveTarget(baseXML, target string) string {
	baseDir := filepath.ToSlash(filepath.Dir(baseXML))
	return filepath.ToSlash(filepath.Join(baseDir, target))
}

func getLayoutAndMaster(r *zip.ReadCloser, slideIndex int) (layoutXML, masterXML string) {
	slideXML := "ppt/slides/slide" + strconv.Itoa(slideIndex) + ".xml"
	slideRels := "ppt/slides/_rels/slide" + strconv.Itoa(slideIndex) + ".xml.rels"
	rf, err := readRels(r, slideRels)
	if err != nil {
		return "", ""
	}
	for _, rel := range rf.Rel {
		if strings.Contains(rel.Type, "slideLayout") {
			layoutXML = resolveTarget(slideXML, rel.Target)
			break
		}
	}
	if layoutXML == "" {
		return "", ""
	}
	lrels := strings.Replace(layoutXML, "slideLayouts/", "slideLayouts/_rels/", 1) + ".rels"
	lrf, err := readRels(r, lrels)
	if err != nil {
		return layoutXML, ""
	}
	for _, rel := range lrf.Rel {
		if strings.Contains(rel.Type, "slideMaster") {
			masterXML = resolveTarget(layoutXML, rel.Target)
			break
		}
	}
	return layoutXML, masterXML
}

func pickDefRPr(ls *listStyle, level string) *rPr {
	if ls == nil {
		return nil
	}
	switch level {
	case "0", "1":
		if ls.Lvl1pPr != nil {
			return ls.Lvl1pPr.DefRPr
		}
	case "2":
		if ls.Lvl2pPr != nil {
			return ls.Lvl2pPr.DefRPr
		}
	case "3":
		if ls.Lvl3pPr != nil {
			return ls.Lvl3pPr.DefRPr
		}
	case "4":
		if ls.Lvl4pPr != nil {
			return ls.Lvl4pPr.DefRPr
		}
	case "5":
		if ls.Lvl5pPr != nil {
			return ls.Lvl5pPr.DefRPr
		}
	case "6":
		if ls.Lvl6pPr != nil {
			return ls.Lvl6pPr.DefRPr
		}
	case "7":
		if ls.Lvl7pPr != nil {
			return ls.Lvl7pPr.DefRPr
		}
	case "8":
		if ls.Lvl8pPr != nil {
			return ls.Lvl8pPr.DefRPr
		}
	case "9":
		if ls.Lvl9pPr != nil {
			return ls.Lvl9pPr.DefRPr
		}
	}
	if ls.Lvl1pPr != nil {
		return ls.Lvl1pPr.DefRPr
	}
	return nil
}

func fallbackColorFromRPr(def *rPr) string {
	if def == nil || def.SolidFill == nil {
		return ""
	}
	if def.SolidFill.SrgbClr != nil {
		return def.SolidFill.SrgbClr.Val
	}
	if def.SolidFill.SchemeClr != nil {
		return convertSchemeColor(def.SolidFill.SchemeClr.Val)
	}
	if def.SolidFill.PrstClr != nil {
		return convertPresetColor(def.SolidFill.PrstClr.Val)
	}
	return ""
}

func fallbackColor(r *zip.ReadCloser, slideIndex int, level string) string {
	layoutXML, masterXML := getLayoutAndMaster(r, slideIndex)
	if layoutXML != "" {
		if data, err := readFileFromZip(r, layoutXML); err == nil {
			var doc txStylesDoc
			if xml.Unmarshal(data, &doc) == nil && doc.TxStyles != nil {
				if c := fallbackColorFromRPr(pickDefRPr(doc.TxStyles.BodyStyle, level)); c != "" {
					return c
				}
				if c := fallbackColorFromRPr(pickDefRPr(doc.TxStyles.OtherStyle, level)); c != "" {
					return c
				}
				if c := fallbackColorFromRPr(pickDefRPr(doc.TxStyles.TitleStyle, level)); c != "" {
					return c
				}
			}
		}
	}
	if masterXML != "" {
		if data, err := readFileFromZip(r, masterXML); err == nil {
			var doc txStylesDoc
			if xml.Unmarshal(data, &doc) == nil && doc.TxStyles != nil {
				if c := fallbackColorFromRPr(pickDefRPr(doc.TxStyles.BodyStyle, level)); c != "" {
					return c
				}
				if c := fallbackColorFromRPr(pickDefRPr(doc.TxStyles.OtherStyle, level)); c != "" {
					return c
				}
				if c := fallbackColorFromRPr(pickDefRPr(doc.TxStyles.TitleStyle, level)); c != "" {
					return c
				}
			}
		}
	}
	return ""
}

// ApplyColorFallback sets colors for runs that have no color derived from the slide.
// Uses level "0" as a default when paragraph level is unknown.
func ApplyColorFallback(r *zip.ReadCloser, slideIndex int, boxes []TextBox) []TextBox {
	for bi := range boxes {
		for ri := range boxes[bi].Runs {
			if strings.TrimSpace(boxes[bi].Runs[ri].Color) == "" {
				c := fallbackColor(r, slideIndex, "0")
				if c != "" {
					boxes[bi].Runs[ri].Color = c
				}
			}
		}
	}
	return boxes
}


