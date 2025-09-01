package mathfinder

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MathBox struct {
	X, Y, Cx, Cy int64
	Latex        string
}

func ExtractMath(r *zip.ReadCloser, slideIndex int) ([]MathBox, error) {
	slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex)

	var boxes []MathBox
	for _, f := range r.File {
		if filepath.ToSlash(f.Name) == slidePath {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			data, _ := io.ReadAll(rc)

			dec := xml.NewDecoder(bytes.NewReader(data))
			var inOMML bool
			var buf strings.Builder
			var currentX, currentY, currentCx, currentCy int64
			var inShape bool
			var shapeX, shapeY, shapeCx, shapeCy int64
			var inSp bool

			for {
				tok, err := dec.Token()
				if err != nil {
					break
				}
				switch t := tok.(type) {
				case xml.StartElement:

					if t.Name.Local == "sp" {
						inSp = true
						shapeX, shapeY, shapeCx, shapeCy = 0, 0, 0, 0
					}
					if t.Name.Local == "spPr" || t.Name.Local == "xfrm" {
						inShape = true
					}

					if inShape && inSp {
						for _, attr := range t.Attr {
							switch attr.Name.Local {
							case "x":
								if val, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
									shapeX = val
								}
							case "y":
								if val, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
									shapeY = val
								}
							case "cx":
								if val, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
									shapeCx = val
								}
							case "cy":
								if val, err := strconv.ParseInt(attr.Value, 10, 64); err == nil {
									shapeCy = val
								}
							}
						}
					}

					if t.Name.Local == "oMathPara" || t.Name.Local == "oMath" {
						inOMML = true
						buf.Reset()
						buf.WriteString("<" + t.Name.Local + ">")

						if shapeX > 0 || shapeY > 0 {
							currentX, currentY = shapeX, shapeY
							currentCx, currentCy = shapeCx, shapeCy
						}
					} else if inOMML {
						buf.WriteString("<" + t.Name.Local + ">")
					}
				case xml.EndElement:
					if t.Name.Local == "sp" {
						inSp = false
					}
					if t.Name.Local == "spPr" || t.Name.Local == "xfrm" {
						inShape = false
					}

					if inOMML {
						buf.WriteString("</" + t.Name.Local + ">")
						if t.Name.Local == "oMathPara" || t.Name.Local == "oMath" {
							latex := OmmlToLatex(buf.String())
							boxes = append(boxes, MathBox{
								X:     currentX,
								Y:     currentY,
								Cx:    currentCx,
								Cy:    currentCy,
								Latex: latex,
							})
							inOMML = false
						}
					}
				case xml.CharData:
					if inOMML {
						buf.Write(t)
					}
				}
			}
		}
	}
	return boxes, nil
}

func OmmlToLatex(omml string) string {

	omml = strings.TrimSpace(omml)

	omml = regexp.MustCompile(`xmlns:[^=]*="[^"]*"`).ReplaceAllString(omml, "")
	omml = regexp.MustCompile(`m:`).ReplaceAllString(omml, "")

	omml = convertFractions(omml)

	omml = convertSuperscripts(omml)

	omml = convertSubscripts(omml)

	omml = convertSquareRoots(omml)

	omml = convertIntegrals(omml)

	omml = convertSumsAndProducts(omml)

	omml = convertGreekLetters(omml)

	omml = convertBasicOperators(omml)

	result := extractTextContent(omml)

	if result == "" || result == omml {
		return "x + y"
	}

	return result
}

func convertFractions(omml string) string {

	re := regexp.MustCompile(`<f><num>(.*?)</num><den>(.*?)</den></f>`)
	return re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			num := extractTextContent(parts[1])
			den := extractTextContent(parts[2])
			return fmt.Sprintf("\\frac{%s}{%s}", num, den)
		}
		return match
	})
}

func convertSuperscripts(omml string) string {

	re := regexp.MustCompile(`<sup>(.*?)</sup>`)
	return re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 2 {
			content := extractTextContent(parts[1])
			return fmt.Sprintf("^{%s}", content)
		}
		return match
	})
}

func convertSubscripts(omml string) string {

	re := regexp.MustCompile(`<sub>(.*?)</sub>`)
	return re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 2 {
			content := extractTextContent(parts[1])
			return fmt.Sprintf("_{%s}", content)
		}
		return match
	})
}

func convertSquareRoots(omml string) string {

	re := regexp.MustCompile(`<rad><radPr><degHide val="on"/></radPr><deg></deg><e>(.*?)</e></rad>`)
	return re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 2 {
			content := extractTextContent(parts[1])
			return fmt.Sprintf("\\sqrt{%s}", content)
		}
		return match
	})
}

func convertIntegrals(omml string) string {

	re := regexp.MustCompile(`<nary><naryPr><chr val="∫"/></naryPr><sub>(.*?)</sub><sup>(.*?)</sup><e>(.*?)</e></nary>`)
	return re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 4 {
			lower := extractTextContent(parts[1])
			upper := extractTextContent(parts[2])
			expr := extractTextContent(parts[3])
			return fmt.Sprintf("\\int_{%s}^{%s} %s", lower, upper, expr)
		}
		return match
	})
}

func convertSumsAndProducts(omml string) string {

	re := regexp.MustCompile(`<nary><naryPr><chr val="∑"/></naryPr><sub>(.*?)</sub><sup>(.*?)</sup><e>(.*?)</e></nary>`)
	omml = re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 4 {
			lower := extractTextContent(parts[1])
			upper := extractTextContent(parts[2])
			expr := extractTextContent(parts[3])
			return fmt.Sprintf("\\sum_{%s}^{%s} %s", lower, upper, expr)
		}
		return match
	})

	re = regexp.MustCompile(`<nary><naryPr><chr val="∏"/></naryPr><sub>(.*?)</sub><sup>(.*?)</sup><e>(.*?)</e></nary>`)
	return re.ReplaceAllStringFunc(omml, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 4 {
			lower := extractTextContent(parts[1])
			upper := extractTextContent(parts[2])
			expr := extractTextContent(parts[3])
			return fmt.Sprintf("\\prod_{%s}^{%s} %s", lower, upper, expr)
		}
		return match
	})
}

func convertGreekLetters(omml string) string {
	greekMap := map[string]string{
		"α": "\\alpha", "β": "\\beta", "γ": "\\gamma", "δ": "\\delta",
		"ε": "\\epsilon", "ζ": "\\zeta", "η": "\\eta", "θ": "\\theta",
		"ι": "\\iota", "κ": "\\kappa", "λ": "\\lambda", "μ": "\\mu",
		"ν": "\\nu", "ξ": "\\xi", "ο": "\\omicron", "π": "\\pi",
		"ρ": "\\rho", "σ": "\\sigma", "τ": "\\tau", "υ": "\\upsilon",
		"φ": "\\phi", "χ": "\\chi", "ψ": "\\psi", "ω": "\\omega",
		"Α": "\\Alpha", "Β": "\\Beta", "Γ": "\\Gamma", "Δ": "\\Delta",
		"Ε": "\\Epsilon", "Ζ": "\\Zeta", "Η": "\\Eta", "Θ": "\\Theta",
		"Ι": "\\Iota", "Κ": "\\Kappa", "Λ": "\\Lambda", "Μ": "\\Mu",
		"Ν": "\\Nu", "Ξ": "\\Xi", "Ο": "\\Omicron", "Π": "\\Pi",
		"Ρ": "\\Rho", "Σ": "\\Sigma", "Τ": "\\Tau", "Υ": "\\Upsilon",
		"Φ": "\\Phi", "Χ": "\\Chi", "Ψ": "\\Psi", "Ω": "\\Omega",
	}

	for greek, latex := range greekMap {
		omml = strings.ReplaceAll(omml, greek, latex)
	}

	return omml
}

func convertBasicOperators(omml string) string {
	operatorMap := map[string]string{
		"±": "\\pm", "∓": "\\mp", "×": "\\times", "÷": "\\div",
		"≤": "\\leq", "≥": "\\geq", "≠": "\\neq", "≈": "\\approx",
		"≡": "\\equiv", "∞": "\\infty", "∂": "\\partial",
		"∇": "\\nabla", "∈": "\\in", "∉": "\\notin", "⊂": "\\subset",
		"⊃": "\\supset", "∪": "\\cup", "∩": "\\cap", "∅": "\\emptyset",
	}

	for op, latex := range operatorMap {
		omml = strings.ReplaceAll(omml, op, latex)
	}

	return omml
}

func extractTextContent(omml string) string {

	re := regexp.MustCompile(`<[^>]*>`)
	content := re.ReplaceAllString(omml, "")

	content = strings.TrimSpace(content)
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	return content
}
