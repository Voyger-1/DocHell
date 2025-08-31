package helper

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"pptXhell/imagefinder"
	mathfinder "pptXhell/math"
	"pptXhell/shapefinder"
	"pptXhell/textfinder"
	"strconv"
	"strings"
)

// EMU -> px : 1 inch = 914400 EMU, 96 px/inch
func emuToPx(emu int64) int {
	return int(float64(emu)/914400.0*96.0 + 0.5)
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
	return nil, os.ErrNotExist
}

func MediaDataURIFromZip(r *zip.ReadCloser, mediaPath string) (string, error) {
	if mediaPath == "" {
		return "", nil
	}
	b, err := readFileFromZip(r, mediaPath)
	if err != nil {
		return "", err
	}
	ext := strings.ToLower(filepath.Ext(mediaPath))
	mime := "image/png"
	if ext == ".jpg" || ext == ".jpeg" {
		mime = "image/jpeg"
	} else if ext == ".gif" {
		mime = "image/gif"
	} else if ext == ".svg" {
		mime = "image/svg+xml"
	}
	return "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(b), nil
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func runStyle(run textfinder.TextRun) string {
	styles := make([]string, 0, 8)

	styles = append(styles, "z-index:1")
	styles = append(styles, "position:relative")

	// Font weight
	if run.Bold {
		styles = append(styles, "font-weight:bold")
	} else {
		styles = append(styles, "font-weight:normal")
	}

	// Font style
	if run.Italic {
		styles = append(styles, "font-style:italic")
	} else {
		styles = append(styles, "font-style:normal")
	}

	// Underline
	if run.Underline {
		styles = append(styles, "text-decoration:underline")
	}

	// Font size with better conversion
	if run.FontSize > 0 {
		pt := float64(run.FontSize)
		px := int(pt*1.333 + 0.5)
		if px < 12 {
			px = 12 // Minimum readable size
		}
		styles = append(styles, "font-size:"+strconv.Itoa(px)+"px")
	} else {
		styles = append(styles, "font-size:16px") // Default size
	}

	// Font family
	if run.Font != "" {
		font := strings.ReplaceAll(run.Font, "\"", "")
		// Add fallback fonts
		styles = append(styles, "font-family:"+font+",Arial,sans-serif")
	} else {
		styles = append(styles, "font-family:Arial,sans-serif")
	}

	// Color with better handling
	if run.Color != "" {
		c := strings.TrimSpace(run.Color)
		if !strings.HasPrefix(c, "#") {
			c = "#" + c
		}
		// Ensure color is valid
		if len(c) == 7 && c[0] == '#' {
			styles = append(styles, "color:"+c)
		} else {
			styles = append(styles, "color:#000000") // Default black
		}
	} else {
		styles = append(styles, "color:#000000") // Default black
	}

	// Additional text properties
	styles = append(styles, "line-height:1.2")
	styles = append(styles, "margin:0")
	styles = append(styles, "padding:0")

	return strings.Join(styles, ";") + ";"

}

func BuildSlideHTML(bgURI string, boxes []textfinder.TextBox, mathBoxes []mathfinder.MathBox, shapes []shapefinder.ShapeBox, images []imagefinder.ImageBox) string {
	slideW := 1280
	slideH := 720

	sb := &strings.Builder{}
	sb.WriteString("<!doctype html><html><head><meta charset='utf-8'>")
	sb.WriteString("<meta name='viewport' content='width=device-width,initial-scale=1'>")
	sb.WriteString("<style>body{margin:0;background:#333}")
	sb.WriteString(".slide{position:relative;margin:16px auto;background:#fff;")
	sb.WriteString("max-width:1280px;width:100%;height:720px;")
	sb.WriteString("box-shadow:0 4px 18px rgba(0,0,0,.25);overflow:hidden;}")
	sb.WriteString(".tb{position:absolute;box-sizing:border-box;padding:8px;white-space:pre-wrap;line-height:1.4;}")
	sb.WriteString(".math{display:flex;align-items:center;justify-content:center;}")
	sb.WriteString(".math div{font-size:24px;text-align:center;}")
	sb.WriteString(".text-box{background:transparent;border:none;}")
	sb.WriteString(".shape{background:transparent;border:1px solid transparent;}")
	sb.WriteString(".slide img{position:absolute;}") // for images
	sb.WriteString("</style></head><body>")
	sb.WriteString(`<script src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"></script>`)

	if bgURI != "" {
		sb.WriteString(fmt.Sprintf(
			"<div class='slide' style=\"background-image:url('%s');background-size:cover;background-position:center;width:%dpx;height:%dpx;\">",
			bgURI, slideW, slideH))
	} else {
		sb.WriteString(fmt.Sprintf(
			"<div class='slide' style=\"background:#fff;width:%dpx;height:%dpx;\">",
			slideW, slideH))
	}

	// Render math boxes
	for _, mb := range mathBoxes {
		left := emuToPx(mb.X)
		top := emuToPx(mb.Y)
		w := emuToPx(mb.Cx)
		h := emuToPx(mb.Cy)
		if w == 0 {
			w = 300
		}
		if h == 0 {
			h = 80
		}

		sb.WriteString(fmt.Sprintf(
			"<div class='tb math' style='left:%dpx;top:%dpx;width:%dpx;height:%dpx;'>"+
				"<div>\\(%s\\)</div></div>",
			left, top, w, h, htmlEscape(mb.Latex)))
	}

	// Render text boxes
	for _, tb := range boxes {
		left := emuToPx(tb.X)
		top := emuToPx(tb.Y)
		w := emuToPx(tb.Cx)
		h := emuToPx(tb.Cy)
		if w == 0 {
			w = 300
		}
		if h == 0 {
			h = 80
		}

		className := "tb text-box"
		if tb.ShapeName != "" {
			className = "tb shape"
		}

		inner := &strings.Builder{}
		for _, run := range tb.Runs {
			style := runStyle(run)
			text := htmlEscape(run.Text)
			if style != "" {
				inner.WriteString(fmt.Sprintf("<p style=\"%s\">%s</p>", style, text))
			} else {
				inner.WriteString(text)
			}
		}

		boxStyle := fmt.Sprintf("left:%dpx;top:%dpx;width:%dpx;height:%dpx;", left, top, w, h)
		if tb.BackgroundColor != "" {
			boxStyle += fmt.Sprintf("background-color:#%s;", tb.BackgroundColor)
		}
		if tb.BorderColor != "" {
			boxStyle += fmt.Sprintf("border:2px solid #%s;", tb.BorderColor)
		}
		if tb.ShapeType == "roundRect" {
			boxStyle += "border-radius:8px;"
		}
		boxStyle += "padding:8px;box-sizing:border-box;"

		sb.WriteString(fmt.Sprintf("<div class='%s' style='%s'>%s</div>",
			className, boxStyle, inner.String()))
	}

	// Render shapes
	for _, sh := range shapes {
		left := emuToPx(sh.X)
		top := emuToPx(sh.Y)
		w := emuToPx(sh.Cx)
		h := emuToPx(sh.Cy)

		fill := "none"
		if sh.Fill != "" {
			fill = sh.Fill
		}
		stroke := "black"
		if sh.Stroke != "" {
			stroke = sh.Stroke
		}

		svg := ""
		switch sh.Geom {
		case "rect":
			svg = fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s" stroke="%s"/>`, w, h, fill, stroke)
		case "ellipse":
			svg = fmt.Sprintf(`<ellipse cx="%d" cy="%d" rx="%d" ry="%d" fill="%s" stroke="%s"/>`, w/2, h/2, w/2, h/2, fill, stroke)
		case "star4":
			svg = fmt.Sprintf(`<polygon points="%d,0 %d,%d %d,%d %d,%d" fill="%s" stroke="%s"/>`,
				w/2, w, h/2, w/2, h, 0, h/2, fill, stroke)
		default:
			svg = fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="%s" stroke="%s"/>`, w, h, fill, stroke)
		}

		sb.WriteString(fmt.Sprintf(
			`<svg class="shape" style="position:absolute;left:%dpx;top:%dpx;width:%dpx;height:%dpx;">%s</svg>`,
			left, top, w, h, svg))
	}

	/*
		Need To fix
		// Render images
	**/
	for _, img := range images {
		// left := emuToPx(img.X)
		// top := emuToPx(img.Y)
		w := emuToPx(img.Cx)
		h := emuToPx(img.Cy)
		if w == 0 {
			w = 200
		}
		if h == 0 {
			h = 200
		}

		// sb.WriteString(fmt.Sprintf(
		// 	`<img src="data:image/png;base64,%s" style="left:%dpx;top:%dpx;width:%dpx;height:%dpx;">`,
		// 	img.Base64, left, top, w, h))
	}

	sb.WriteString("</div></body></html>")
	return sb.String()
}

func NormalizeMediaPath(mediaPath string) string {
	mediaPath = filepath.ToSlash(mediaPath)

	if strings.HasPrefix(mediaPath, "ppt/media/") {
		return mediaPath
	}
	if strings.Contains(mediaPath, "media/") {
		idx := strings.Index(mediaPath, "media/")
		return "ppt/" + mediaPath[idx:]
	}
	return mediaPath
}
