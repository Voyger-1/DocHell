package shapefinder

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type ShapeBox struct {
	X, Y, Cx, Cy int64
	Geom         string
	Fill         string
	Stroke       string
	Text         string
}

type xfrm struct {
	Off struct {
		X int64 `xml:"x,attr"`
		Y int64 `xml:"y,attr"`
	} `xml:"off"`
	Ext struct {
		Cx int64 `xml:"cx,attr"`
		Cy int64 `xml:"cy,attr"`
	} `xml:"ext"`
}

type color struct {
	Val string `xml:"val,attr"`
}

type solidFill struct {
	Clr color `xml:"srgbClr"`
}

type ln struct {
	Fill solidFill `xml:"solidFill"`
}

type spPr struct {
	Xfrm xfrm `xml:"xfrm"`
	Geom struct {
		Prst string `xml:"prst,attr"`
	} `xml:"prstGeom"`
	Fill solidFill `xml:"solidFill"`
	Line ln        `xml:"ln"`
}

type txBody struct {
	P []struct {
		R []struct {
			T string `xml:"t"`
		} `xml:"r"`
	} `xml:"p"`
}

type sp struct {
	SpPr   spPr   `xml:"spPr"`
	TxBody txBody `xml:"txBody"`
}

type cSld struct {
	SpTree struct {
		Shapes []sp `xml:"sp"`
	} `xml:"spTree"`
}

type slide struct {
	XMLName xml.Name `xml:"sld"`
	CSld    cSld     `xml:"cSld"`
}

func ExtractShapes(r *zip.ReadCloser, slideIndex int) ([]ShapeBox, error) {
	slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex)

	for _, f := range r.File {
		if filepath.ToSlash(f.Name) == slidePath {
			rc, _ := f.Open()
			defer rc.Close()
			data, _ := io.ReadAll(rc)

			var s slide
			if err := xml.Unmarshal(data, &s); err != nil {
				return nil, err
			}

			var shapes []ShapeBox
			for _, shp := range s.CSld.SpTree.Shapes {
				sb := ShapeBox{
					X:    shp.SpPr.Xfrm.Off.X,
					Y:    shp.SpPr.Xfrm.Off.Y,
					Cx:   shp.SpPr.Xfrm.Ext.Cx,
					Cy:   shp.SpPr.Xfrm.Ext.Cy,
					Geom: shp.SpPr.Geom.Prst,
				}
				if shp.SpPr.Fill.Clr.Val != "" {
					sb.Fill = "#" + shp.SpPr.Fill.Clr.Val
				}
				if shp.SpPr.Line.Fill.Clr.Val != "" {
					sb.Stroke = "#" + shp.SpPr.Line.Fill.Clr.Val
				}
				var textParts []string
				for _, p := range shp.TxBody.P {
					for _, r := range p.R {
						textParts = append(textParts, r.T)
					}
				}
				sb.Text = strings.Join(textParts, "")
				shapes = append(shapes, sb)
			}
			return shapes, nil
		}
	}
	return nil, fmt.Errorf("slide %d not found", slideIndex)
}
