package shapefinder

// import (
// 	"fmt"
// 	"math"
// 	"strings"
// )

// // SVGFor returns an SVG element string (path/polygon/rect/ellipse) sized to w x h px.
// func SVGFor(s Shape, w, h int) string {
// 	if w <= 0 || h <= 0 {
// 		return ""
// 	}
// 	switch strings.ToLower(s.Geom) {
// 	case "rect", "rectangle":
// 		return fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" rx="0" ry="0"/>`, w, h)
// 	case "roundrect":
// 		rx := int(float64(min(w, h)) * 0.1)
// 		return fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" rx="%d" ry="%d"/>`, w, h, rx, rx)
// 	case "ellipse", "oval":
// 		return fmt.Sprintf(`<ellipse cx="%d" cy="%d" rx="%d" ry="%d"/>`, w/2, h/2, w/2, h/2)
// 	case "diamond":
// 		return poly([][2]float64{{float64(w) / 2, 0}, {float64(w), float64(h) / 2}, {float64(w) / 2, float64(h)}, {0, float64(h) / 2}})
// 	case "triangle", "isoscelestriangle":
// 		return poly([][2]float64{{float64(w) / 2, 0}, {float64(w), float64(h)}, {0, float64(h)}})
// 	case "righttriangle":
// 		return poly([][2]float64{{0, 0}, {float64(w), float64(h)}, {0, float64(h)}})
// 	case "parallelogram":
// 		off := float64(w) * 0.2
// 		return poly([][2]float64{{off, 0}, {float64(w), 0}, {float64(w) - off, float64(h)}, {0, float64(h)}})
// 	case "star4":
// 		return starPoints(w, h, 4, 0.45)
// 	case "star5":
// 		return starPoints(w, h, 5, 0.5)
// 	case "leftarrow":
// 		return arrowLeft(w, h)
// 	case "rightarrow":
// 		return arrowRight(w, h)
// 	case "line":
// 		return `<line x1="0" y1="0" x2="100%" y2="100%" />`
// 	default:
// 		// fallback rectangle so something is visible
// 		return fmt.Sprintf(`<rect x="0" y="0" width="%d" height="%d" fill="none"/>`, w, h)
// 	}
// }

// func min(a, b int) int {
// 	if a < b {
// 		return a
// 	}
// 	return b
// }

// func poly(pts [][2]float64) string {
// 	sb := &strings.Builder{}
// 	for i, p := range pts {
// 		if i > 0 {
// 			sb.WriteByte(' ')
// 		}
// 		sb.WriteString(fmt.Sprintf("%.2f,%.2f", p[0], p[1]))
// 	}
// 	return fmt.Sprintf(`<polygon points="%s"/>`, sb.String())
// }

// func starPoints(w, h, arms int, innerRatio float64) string {
// 	cx, cy := float64(w)/2.0, float64(h)/2.0
// 	R := 0.5 * math.Min(float64(w), float64(h))
// 	r := R * innerRatio
// 	n := arms * 2
// 	pts := make([][2]float64, 0, n)
// 	startAngle := -math.Pi / 2 // top
// 	for i := 0; i < n; i++ {
// 		useR := R
// 		if i%2 == 1 {
// 			useR = r
// 		}
// 		ang := startAngle + float64(i)*2*math.Pi/float64(n)
// 		pts = append(pts, [2]float64{cx + useR*math.Cos(ang), cy + useR*math.Sin(ang)})
// 	}
// 	return poly(pts)
// }

// func arrowRight(w, h int) string {
// 	head := float64(w) * 0.35
// 	shaftH := float64(h) * 0.4
// 	y0 := (float64(h) - shaftH) / 2
// 	return poly([][2]float64{
// 		{0, y0}, {float64(w) - head, y0}, {float64(w) - head, 0}, {float64(w), float64(h) / 2},
// 		{float64(w) - head, float64(h)}, {float64(w) - head, y0 + shaftH}, {0, y0 + shaftH},
// 	})
// }

// func arrowLeft(w, h int) string {
// 	head := float64(w) * 0.35
// 	shaftH := float64(h) * 0.4
// 	y0 := (float64(h) - shaftH) / 2
// 	return poly([][2]float64{
// 		{head, 0}, {head, y0}, {float64(w), y0}, {float64(w), y0 + shaftH}, {head, y0 + shaftH},
// 		{head, float64(h)}, {0, float64(h) / 2},
// 	})
// }
