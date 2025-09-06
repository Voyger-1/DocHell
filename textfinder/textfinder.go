package textfinder

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
)

type TextRun struct {
	Text      string
	Bold      bool
	Italic    bool
	Underline bool
	FontSize  int
	Font      string
	Color     string
}

type TextBox struct {
	Runs            []TextRun
	X, Y, Cx, Cy    int64
	ShapeID         string
	ShapeName       string
	BackgroundColor string
	BorderColor     string
	ShapeType       string
}

type slideXML struct {
	SpTree           spTree `xml:"cSld>spTree"`
	AlternateContent []struct {
		Choice struct {
			Sp sp `xml:"sp"`
		} `xml:"Choice>sp"`
		Fallback struct {
			Sp sp `xml:"sp"`
		} `xml:"Fallback>sp"`
	} `xml:"AlternateContent"`
}

type spTree struct {
	Sp           []sp           `xml:"sp"`
	GrpSp        []grpSp        `xml:"grpSp"`
	GraphicFrame []graphicFrame `xml:"graphicFrame"`
}

type sp struct {
	NvPr struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"nvSpPr>cNvPr"`
	SpPr struct {
		Xfrm struct {
			Off struct {
				X string `xml:"x,attr"`
				Y string `xml:"y,attr"`
			} `xml:"off"`
			Ext struct {
				Cx string `xml:"cx,attr"`
				Cy string `xml:"cy,attr"`
			} `xml:"ext"`
		} `xml:"xfrm"`
		PrstGeom struct {
			Prst string `xml:"prst,attr"`
		} `xml:"prstGeom"`
		SolidFill *struct {
			SrgbClr *struct {
				Val string `xml:"val,attr"`
			} `xml:"srgbClr"`
			SchemeClr *struct {
				Val string `xml:"val,attr"`
			} `xml:"schemeClr"`
		} `xml:"solidFill"`
		Ln *struct {
			SolidFill *struct {
				SrgbClr *struct {
					Val string `xml:"val,attr"`
				} `xml:"srgbClr"`
				SchemeClr *struct {
					Val string `xml:"val,attr"`
				} `xml:"schemeClr"`
			} `xml:"solidFill"`
		} `xml:"ln"`
	} `xml:"spPr"`
	TxBody *txBody `xml:"txBody"`
}

type grpSp struct {
	NvPr struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"nvGrpSpPr>cNvPr"`
	GrpSpPr struct {
		Xfrm xfrm `xml:"xfrm"`
	} `xml:"grpSpPr"`
	Sp           []sp           `xml:"sp"`
	GrpSp        []grpSp        `xml:"grpSp"`
	GraphicFrame []graphicFrame `xml:"graphicFrame"`
}

type graphicFrame struct {
	NvPr struct {
		ID   string `xml:"id,attr"`
		Name string `xml:"name,attr"`
	} `xml:"nvGraphicFramePr>cNvPr"`
	Xfrm    xfrm `xml:"xfrm"`
	Graphic *struct {
		GraphicData *struct {
			Tbl *table `xml:"tbl"`
		} `xml:"graphicData"`
	} `xml:"graphic"`
}

type table struct {
	Tr []struct {
		Tc []struct {
			TxBody *txBody `xml:"txBody"`
		} `xml:"tc"`
	} `xml:"tr"`
}

type txBody struct {
	LstStyle *struct {
		Lvl1pPr *struct {
			DefRPr *rPr `xml:"defRPr"`
		} `xml:"lvl1pPr"`
		Lvl2pPr *struct {
			DefRPr *rPr `xml:"lvl2pPr"`
		} `xml:"lvl2pPr"`
		Lvl3pPr *struct {
			DefRPr *rPr `xml:"lvl3pPr"`
		} `xml:"lvl3pPr"`
		Lvl4pPr *struct {
			DefRPr *rPr `xml:"lvl4pPr"`
		} `xml:"lvl4pPr"`
		Lvl5pPr *struct {
			DefRPr *rPr `xml:"lvl5pPr"`
		} `xml:"lvl5pPr"`
		Lvl6pPr *struct {
			DefRPr *rPr `xml:"lvl6pPr"`
		} `xml:"lvl6pPr"`
		Lvl7pPr *struct {
			DefRPr *rPr `xml:"lvl7pPr"`
		} `xml:"lvl7pPr"`
		Lvl8pPr *struct {
			DefRPr *rPr `xml:"lvl8pPr"`
		} `xml:"lvl8pPr"`
		Lvl9pPr *struct {
			DefRPr *rPr `xml:"lvl9pPr"`
		} `xml:"lvl9pPr"`
	} `xml:"lstStyle"`
	Para []struct {
		PPr *struct {
			DefRPr *rPr   `xml:"defRPr"`
			Lvl    string `xml:"lvl,attr"`
		} `xml:"pPr"`
		R []struct {
			RPr *rPr   `xml:"rPr"`
			T   string `xml:"t"`
		} `xml:"r"`
		AR []struct {
			RPr *rPr   `xml:"rPr"`
			T   string `xml:"t"`
		} `xml:"a:r"`
	} `xml:"p"`
}

type rPr struct {
	Sz    string `xml:"sz,attr"`
	B     string `xml:"b,attr"`
	I     string `xml:"i,attr"`
	U     string `xml:"u,attr"`
	Latin *struct {
		Typeface string `xml:"typeface,attr"`
	} `xml:"latin"`
	SolidFill *struct {
		SrgbClr *struct {
			Val string `xml:"val,attr"`
		} `xml:"srgbClr"`
		SchemeClr *struct {
			Val string `xml:"val,attr"`
		} `xml:"schemeClr"`
		PrstClr *struct {
			Val string `xml:"val,attr"`
		} `xml:"prstClr"`
	} `xml:"solidFill"`
}

type xfrm struct {
	Off struct {
		X string `xml:"x,attr"`
		Y string `xml:"y,attr"`
	} `xml:"off"`
	Ext struct {
		Cx string `xml:"cx,attr"`
		Cy string `xml:"cy,attr"`
	} `xml:"ext"`
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

func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}

func boolAttr(val string) bool {
	return val == "1" || strings.ToLower(val) == "true"
}

func convertSchemeColor(schemeColor string) string {
	switch strings.ToLower(schemeColor) {
	case "accent1":
		return "4472C4"
	case "accent2":
		return "ED7D31"
	case "accent3":
		return "A5A5A5"
	case "accent4":
		return "FFC000"
	case "accent5":
		return "5B9BD5"
	case "accent6":
		return "70AD47"
	case "dk1":
		return "000000"
	case "lt1":
		return "FFFFFF"
	case "dk2":
		return "44546A"
	case "lt2":
		return "E7E6E6"
	case "hlink":
		return "0563C1"
	case "folhlink":
		return "954F72"
	default:
		return "000000"
	}
}

func convertPresetColor(presetColor string) string {
	switch strings.ToLower(presetColor) {
	case "yellow":
		return "FFFF00"
	case "red":
		return "FF0000"
	case "green":
		return "00FF00"
	case "blue":
		return "0000FF"
	case "black":
		return "000000"
	case "white":
		return "FFFFFF"
	case "gray":
		return "808080"
	case "orange":
		return "FFA500"
	case "purple":
		return "800080"
	case "pink":
		return "FFC0CB"
	case "brown":
		return "A52A2A"
	case "cyan":
		return "00FFFF"
	case "magenta":
		return "FF00FF"
	default:
		return "000000"
	}
}

func ExtractTextBoxes(r *zip.ReadCloser, slideIndex int) ([]TextBox, error) {
	slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex)
	data, err := readFileFromZip(r, slidePath)
	if err != nil {
		return nil, err
	}

	var s slideXML
	if err := xml.Unmarshal(data, &s); err != nil {
		return nil, err
	}

	var out []TextBox
	fmt.Printf("Processing slide %d: Found %d shapes, %d groups, %d graphic frames\n",
		slideIndex, len(s.SpTree.Sp), len(s.SpTree.GrpSp), len(s.SpTree.GraphicFrame))

	for i, sp := range s.SpTree.Sp {
		shapes := processShape(sp, 0, 0)
		if len(shapes) > 0 {
			fmt.Printf("Shape %d: %s - Found %d text boxes\n", i, sp.NvPr.Name, len(shapes))
		}
		out = append(out, shapes...)
	}
	for i, g := range s.SpTree.GrpSp {
		shapes := processGroup(g, 0, 0)
		if len(shapes) > 0 {
			fmt.Printf("Group %d: %s - Found %d text boxes\n", i, g.NvPr.Name, len(shapes))
		}
		out = append(out, shapes...)
	}
	for i, gf := range s.SpTree.GraphicFrame {
		shapes := processGraphicFrame(gf, 0, 0)
		if len(shapes) > 0 {
			fmt.Printf("GraphicFrame %d: %s - Found %d text boxes\n", i, gf.NvPr.Name, len(shapes))
		}
		out = append(out, shapes...)
	}

	for i, ac := range s.AlternateContent {

		if ac.Choice.Sp.NvPr.ID != "" {
			shapes := processShape(ac.Choice.Sp, 0, 0)
			if len(shapes) > 0 {
				fmt.Printf("AlternateContent Choice %d: %s - Found %d text boxes\n", i, ac.Choice.Sp.NvPr.Name, len(shapes))
			}
			out = append(out, shapes...)
		}

		if ac.Fallback.Sp.NvPr.ID != "" {
			shapes := processShape(ac.Fallback.Sp, 0, 0)
			if len(shapes) > 0 {
				fmt.Printf("AlternateContent Fallback %d: %s - Found %d text boxes\n", i, ac.Fallback.Sp.NvPr.Name, len(shapes))
			}
			out = append(out, shapes...)
		}
	}

	fmt.Printf("Total text boxes found: %d\n", len(out))
	return out, nil
}

func processShape(sp sp, parentX, parentY int64) []TextBox {
	var out []TextBox
	x := parseInt64(sp.SpPr.Xfrm.Off.X)
	y := parseInt64(sp.SpPr.Xfrm.Off.Y)
	cx := parseInt64(sp.SpPr.Xfrm.Ext.Cx)
	cy := parseInt64(sp.SpPr.Xfrm.Ext.Cy)
	absX, absY := parentX+x, parentY+y

	if sp.TxBody != nil {
		var tb TextBox
		tb.X, tb.Y, tb.Cx, tb.Cy = absX, absY, cx, cy
		tb.ShapeID, tb.ShapeName = sp.NvPr.ID, sp.NvPr.Name
		tb.ShapeType = sp.SpPr.PrstGeom.Prst

		if sp.SpPr.SolidFill != nil {
			if sp.SpPr.SolidFill.SrgbClr != nil {
				tb.BackgroundColor = sp.SpPr.SolidFill.SrgbClr.Val
			} else if sp.SpPr.SolidFill.SchemeClr != nil {
				tb.BackgroundColor = convertSchemeColor(sp.SpPr.SolidFill.SchemeClr.Val)
			}
		}

		if sp.SpPr.Ln != nil && sp.SpPr.Ln.SolidFill != nil {
			if sp.SpPr.Ln.SolidFill.SrgbClr != nil {
				tb.BorderColor = sp.SpPr.Ln.SolidFill.SrgbClr.Val
			} else if sp.SpPr.Ln.SolidFill.SchemeClr != nil {
				tb.BorderColor = convertSchemeColor(sp.SpPr.Ln.SolidFill.SchemeClr.Val)
			}
		}
		for _, p := range sp.TxBody.Para {

			var defaultRPr *rPr
			if p.PPr != nil && p.PPr.DefRPr != nil {
				defaultRPr = p.PPr.DefRPr
			}

			if sp.TxBody.LstStyle != nil {
				level := "0"
				if p.PPr != nil && p.PPr.Lvl != "" {
					level = p.PPr.Lvl
				}

				switch level {
				case "0", "1":
					if sp.TxBody.LstStyle.Lvl1pPr != nil && sp.TxBody.LstStyle.Lvl1pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl1pPr.DefRPr

					}
				case "2":
					if sp.TxBody.LstStyle.Lvl2pPr != nil && sp.TxBody.LstStyle.Lvl2pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl2pPr.DefRPr
					}
				case "3":
					if sp.TxBody.LstStyle.Lvl3pPr != nil && sp.TxBody.LstStyle.Lvl3pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl3pPr.DefRPr
					}
				case "4":
					if sp.TxBody.LstStyle.Lvl4pPr != nil && sp.TxBody.LstStyle.Lvl4pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl4pPr.DefRPr
					}
				case "5":
					if sp.TxBody.LstStyle.Lvl5pPr != nil && sp.TxBody.LstStyle.Lvl5pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl5pPr.DefRPr
					}
				case "6":
					if sp.TxBody.LstStyle.Lvl6pPr != nil && sp.TxBody.LstStyle.Lvl6pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl6pPr.DefRPr
					}
				case "7":
					if sp.TxBody.LstStyle.Lvl7pPr != nil && sp.TxBody.LstStyle.Lvl7pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl7pPr.DefRPr
					}
				case "8":
					if sp.TxBody.LstStyle.Lvl8pPr != nil && sp.TxBody.LstStyle.Lvl8pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl8pPr.DefRPr
					}
				case "9":
					if sp.TxBody.LstStyle.Lvl9pPr != nil && sp.TxBody.LstStyle.Lvl9pPr.DefRPr != nil {
						defaultRPr = sp.TxBody.LstStyle.Lvl9pPr.DefRPr
					}
				}
			}

			for _, r := range p.R {
				run := TextRun{Text: r.T}

				if r.RPr != nil {

					run.Bold = boolAttr(r.RPr.B)
					run.Italic = boolAttr(r.RPr.I)
					run.Underline = r.RPr.U != ""
					if r.RPr.Sz != "" {
						fs, _ := strconv.Atoi(r.RPr.Sz)
						run.FontSize = fs / 100
					}
					if r.RPr.Latin != nil {
						run.Font = r.RPr.Latin.Typeface
					}
					if r.RPr.SolidFill != nil {
						if r.RPr.SolidFill.SrgbClr != nil {
							run.Color = r.RPr.SolidFill.SrgbClr.Val
						} else if r.RPr.SolidFill.SchemeClr != nil {

							schemeColor := r.RPr.SolidFill.SchemeClr.Val
							run.Color = convertSchemeColor(schemeColor)
						} else if r.RPr.SolidFill.PrstClr != nil {

							presetColor := r.RPr.SolidFill.PrstClr.Val
							run.Color = convertPresetColor(presetColor)
						}
					}

					if run.Color == "" && defaultRPr != nil && defaultRPr.SolidFill != nil {
						if defaultRPr.SolidFill.SrgbClr != nil {
							run.Color = defaultRPr.SolidFill.SrgbClr.Val
						} else if defaultRPr.SolidFill.SchemeClr != nil {
							schemeColor := defaultRPr.SolidFill.SchemeClr.Val
							run.Color = convertSchemeColor(schemeColor)
						} else if defaultRPr.SolidFill.PrstClr != nil {
							presetColor := defaultRPr.SolidFill.PrstClr.Val
							run.Color = convertPresetColor(presetColor)
						}
					}
				} else if defaultRPr != nil {

					run.Bold = boolAttr(defaultRPr.B)
					run.Italic = boolAttr(defaultRPr.I)
					run.Underline = defaultRPr.U != ""
					if defaultRPr.Sz != "" {
						fs, _ := strconv.Atoi(defaultRPr.Sz)
						run.FontSize = fs / 100
					}
					if defaultRPr.Latin != nil {
						run.Font = defaultRPr.Latin.Typeface
					}
					if defaultRPr.SolidFill != nil {
						if defaultRPr.SolidFill.SrgbClr != nil {
							run.Color = defaultRPr.SolidFill.SrgbClr.Val

						} else if defaultRPr.SolidFill.SchemeClr != nil {

							schemeColor := defaultRPr.SolidFill.SchemeClr.Val
							run.Color = convertSchemeColor(schemeColor)

						} else if defaultRPr.SolidFill.PrstClr != nil {

							presetColor := defaultRPr.SolidFill.PrstClr.Val
							run.Color = convertPresetColor(presetColor)

						}
					} else {
						fmt.Printf("DEBUG: defaultRPr.SolidFill is nil\n")
					}
				}
				if strings.TrimSpace(run.Text) != "" {
					fmt.Printf("DEBUG: Text='%s', Color='%s'\n", run.Text, run.Color)
					tb.Runs = append(tb.Runs, run)
				}
			}

			for _, r := range p.AR {
				run := TextRun{Text: r.T}
				if r.RPr != nil {
					run.Bold = boolAttr(r.RPr.B)
					run.Italic = boolAttr(r.RPr.I)
					run.Underline = r.RPr.U != ""
					if r.RPr.Sz != "" {
						fs, _ := strconv.Atoi(r.RPr.Sz)
						run.FontSize = fs / 100
					}
					if r.RPr.Latin != nil {
						run.Font = r.RPr.Latin.Typeface
					}
					if r.RPr.SolidFill != nil {
						if r.RPr.SolidFill.SrgbClr != nil {
							run.Color = r.RPr.SolidFill.SrgbClr.Val
						} else if r.RPr.SolidFill.SchemeClr != nil {

							schemeColor := r.RPr.SolidFill.SchemeClr.Val
							run.Color = convertSchemeColor(schemeColor)
						} else if r.RPr.SolidFill.PrstClr != nil {

							presetColor := r.RPr.SolidFill.PrstClr.Val
							run.Color = convertPresetColor(presetColor)
						}
					}
				} else if defaultRPr != nil {

					run.Bold = boolAttr(defaultRPr.B)
					run.Italic = boolAttr(defaultRPr.I)
					run.Underline = defaultRPr.U != ""
					if defaultRPr.Sz != "" {
						fs, _ := strconv.Atoi(defaultRPr.Sz)
						run.FontSize = fs / 100
					}
					if defaultRPr.Latin != nil {
						run.Font = defaultRPr.Latin.Typeface
					}
					if defaultRPr.SolidFill != nil {
						if defaultRPr.SolidFill.SrgbClr != nil {
							run.Color = defaultRPr.SolidFill.SrgbClr.Val
						} else if defaultRPr.SolidFill.SchemeClr != nil {

							schemeColor := defaultRPr.SolidFill.SchemeClr.Val
							run.Color = convertSchemeColor(schemeColor)
						} else if defaultRPr.SolidFill.PrstClr != nil {
							presetColor := defaultRPr.SolidFill.PrstClr.Val
							run.Color = convertPresetColor(presetColor)
						}
					}
				}
				if strings.TrimSpace(run.Text) != "" {
					fmt.Printf("DEBUG: Text='%s', Color='%s'\n", run.Text, run.Color)
					tb.Runs = append(tb.Runs, run)
				}
			}
		}
		if len(tb.Runs) > 0 {
			out = append(out, tb)
		}
	}
	return out
}

func processGroup(g grpSp, parentX, parentY int64) []TextBox {
	var out []TextBox
	x := parseInt64(g.GrpSpPr.Xfrm.Off.X)
	y := parseInt64(g.GrpSpPr.Xfrm.Off.Y)
	absX, absY := parentX+x, parentY+y

	for _, sp := range g.Sp {
		out = append(out, processShape(sp, absX, absY)...)
	}
	for _, ng := range g.GrpSp {
		out = append(out, processGroup(ng, absX, absY)...)
	}
	for _, gf := range g.GraphicFrame {
		out = append(out, processGraphicFrame(gf, absX, absY)...)
	}
	return out
}

func processGraphicFrame(gf graphicFrame, parentX, parentY int64) []TextBox {
	var out []TextBox
	x := parseInt64(gf.Xfrm.Off.X)
	y := parseInt64(gf.Xfrm.Off.Y)
	cx := parseInt64(gf.Xfrm.Ext.Cx)
	cy := parseInt64(gf.Xfrm.Ext.Cy)
	absX, absY := parentX+x, parentY+y

	if gf.Graphic != nil && gf.Graphic.GraphicData != nil && gf.Graphic.GraphicData.Tbl != nil {
		for _, row := range gf.Graphic.GraphicData.Tbl.Tr {
			for _, cell := range row.Tc {
				if cell.TxBody == nil {
					continue
				}
				var tb TextBox
				tb.X, tb.Y, tb.Cx, tb.Cy = absX, absY, cx, cy
				for _, p := range cell.TxBody.Para {
					for _, r := range p.R {
						run := TextRun{Text: r.T}
						if r.RPr != nil {
							run.Bold = boolAttr(r.RPr.B)
							run.Italic = boolAttr(r.RPr.I)
							run.Underline = r.RPr.U != ""
							if r.RPr.Sz != "" {
								fs, _ := strconv.Atoi(r.RPr.Sz)
								run.FontSize = fs / 100
							}
							if r.RPr.Latin != nil {
								run.Font = r.RPr.Latin.Typeface
							}
							if r.RPr.SolidFill != nil && r.RPr.SolidFill.SrgbClr != nil {
								run.Color = r.RPr.SolidFill.SrgbClr.Val
							}
						}
						if strings.TrimSpace(run.Text) != "" {
							tb.Runs = append(tb.Runs, run)
						}
					}
				}
				if len(tb.Runs) > 0 {
					out = append(out, tb)
				}
			}
		}
	}
	return out
}
