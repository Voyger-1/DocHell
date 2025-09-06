package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func removeChoiceBlocks(xml string) string {
	re := regexp.MustCompile(`(?s)<mc:Choice.*?</mc:Choice>`)

	return re.ReplaceAllString(xml, "")
}

func removeFallBackBlocks(xml string) string {
	idk := regexp.MustCompile(`(?s)<mc:Fallback.*?</mc:Fallback>`)

	return idk.ReplaceAllString(xml, "")
}

type Link struct {
	Href   string `json:"href"`
	Text   string `json:"text"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type TextElement struct {
	Text   string `json:"text"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type Media struct {
	File     string `json:"file"`
	Type     string `json:"type"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Rotation int    `json:"rotation"`
	FlipH    bool   `json:"flipH"`
	FlipV    bool   `json:"flipV"`
}

type Slide struct {
	Slide      string        `json:"slide"`
	Links      []Link        `json:"links"`
	Media      []Media       `json:"media"`
	TextBlocks []TextElement `json:"textBlocks"`
}

type Metadata struct {
	Author      string `json:"author"`
	Title       string `json:"title"`
	SlideCount  string `json:"slideCount"`
	Application string `json:"application"`
	Company     string `json:"company"`
}

type Result struct {
	Metadata Metadata `json:"metadata"`
	Slides   []Slide  `json:"slides"`
}

type Relationship struct {
	Id     string
	Target string
	Type   string
}

const (
	defaultSlideWidthEMU  = 96
	defaultSlideHeightEMU = 96
	TargetWidthPX         = 1280
	TargetHeightPX        = 720
)

var (
	slideWidthEMU  = defaultSlideWidthEMU
	slideHeightEMU = defaultSlideHeightEMU

	offRe   = regexp.MustCompile(`<a:off[^>]*x="(\d+)"[^>]*y="(\d+)"`)
	extRe   = regexp.MustCompile(`<a:ext[^>]*cx="(\d+)"[^>]*cy="(\d+)"`)
	rotRe   = regexp.MustCompile(`rot="(-?\d+)"`)
	flipHRe = regexp.MustCompile(`flipH="1"`)
	flipVRe = regexp.MustCompile(`flipV="1"`)
	rIdRe   = regexp.MustCompile(`r:(?:embed|link)\s*=\s*"(rId\d+)"`)
)

func cleanAndRepackPPTX(input, output string) error {
	r, err := zip.OpenReader(input)
	if err != nil {
		return fmt.Errorf("cannot open pptx: %v", err)
	}
	defer r.Close()

	outFile, err := os.Create(output)
	if err != nil {
		return fmt.Errorf("cannot create output pptx: %v", err)
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)
	defer zw.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		data, _ := io.ReadAll(rc)
		rc.Close()

		if strings.HasPrefix(f.Name, "ppt/") && strings.HasSuffix(f.Name, ".xml") {
			data = []byte(removeChoiceBlocks(string(data)))
			data = []byte(removeFallBackBlocks(string(data)))

		}

		w, err := zw.CreateHeader(&zip.FileHeader{
			Name:   f.Name,
			Method: f.Method,
		})
		if err != nil {
			return err
		}
		_, err = io.Copy(w, bytes.NewReader(data))
		if err != nil {
			return err
		}
	}

	return nil
}

func readZipFile(f *zip.File) string {
	rc, _ := f.Open()
	if rc == nil {
		return ""
	}
	defer rc.Close()
	buf, _ := io.ReadAll(rc)
	return string(buf)
}

func emuToPxX(emu string) int {
	val, _ := strconv.ParseInt(emu, 10, 64)
	return int(float64(val) / float64(slideWidthEMU) * float64(TargetWidthPX))
}

func emuToPxY(emu string) int {
	val, _ := strconv.ParseInt(emu, 10, 64)
	return int(float64(val) / float64(slideHeightEMU) * float64(TargetHeightPX))
}

func getCoordsNear(xmlStr string, matchPos int) (x, y, w, h, rotation int, flipH, flipV bool) {
	parentStart := strings.LastIndex(xmlStr[:matchPos], "<p:")
	if parentStart < 0 {
		parentStart = matchPos - 2000
		if parentStart < 0 {
			parentStart = 0
		}
	}

	windowBefore := xmlStr[parentStart:matchPos]
	windowAfterEnd := matchPos + 2000
	if windowAfterEnd > len(xmlStr) {
		windowAfterEnd = len(xmlStr)
	}
	windowAfter := xmlStr[matchPos:windowAfterEnd]

	if offMatches := offRe.FindAllStringSubmatch(windowBefore, -1); len(offMatches) > 0 {
		last := offMatches[len(offMatches)-1]
		x = emuToPxX(last[1])
		y = emuToPxY(last[2])
	}
	if extMatches := extRe.FindAllStringSubmatch(windowBefore, -1); len(extMatches) > 0 {
		last := extMatches[len(extMatches)-1]
		w = emuToPxX(last[1])
		h = emuToPxY(last[2])
	}

	if rotMatches := rotRe.FindStringSubmatch(windowBefore); len(rotMatches) > 1 {
		raw, _ := strconv.Atoi(rotMatches[1])
		rotation = raw / 60000
	}

	flipH = flipHRe.MatchString(windowBefore)
	flipV = flipVRe.MatchString(windowBefore)

	if x == 0 && y == 0 {
		if offMatches := offRe.FindAllStringSubmatch(windowAfter, -1); len(offMatches) > 0 {
			first := offMatches[0]
			x = emuToPxX(first[1])
			y = emuToPxY(first[2])
		}
	}
	if w == 0 && h == 0 {
		if extMatches := extRe.FindAllStringSubmatch(windowAfter, -1); len(extMatches) > 0 {
			first := extMatches[0]
			w = emuToPxX(first[1])
			h = emuToPxY(first[2])
		}
	}
	if rotation == 0 {
		if rotMatches := rotRe.FindStringSubmatch(windowAfter); len(rotMatches) > 1 {
			raw, _ := strconv.Atoi(rotMatches[1])
			rotation = raw / 60000
		}
	}
	if !flipH {
		flipH = flipHRe.MatchString(windowAfter)
	}
	if !flipV {
		flipV = flipVRe.MatchString(windowAfter)
	}

	return
}

func exportFile(target string, files map[string]*zip.File, outDir string) string {
	target = strings.TrimPrefix(target, "../")
	key := "ppt/" + target
	if f, ok := files[key]; ok {
		os.MkdirAll(outDir, 0755)
		outPath := filepath.Join(outDir, filepath.Base(target))
		rc, _ := f.Open()
		if rc == nil {
			return ""
		}
		defer rc.Close()
		out, _ := os.Create(outPath)
		defer out.Close()
		io.Copy(out, rc)
		return outPath
	}
	if f, ok := files[target]; ok {
		os.MkdirAll(outDir, 0755)
		outPath := filepath.Join(outDir, filepath.Base(target))
		rc, _ := f.Open()
		if rc == nil {
			return ""
		}
		defer rc.Close()
		out, _ := os.Create(outPath)
		defer out.Close()
		io.Copy(out, rc)
		return outPath
	}
	return ""
}

func extractPPTX(pptxPath string, outJSON string) error {
	r, err := zip.OpenReader(pptxPath)
	if err != nil {
		return fmt.Errorf("cannot open pptx: %v", err)
	}
	defer r.Close()

	files := map[string]*zip.File{}
	for _, f := range r.File {
		files[f.Name] = f
	}

	var result Result
	result.Metadata = Metadata{}

	if presF, ok := files["ppt/presentation.xml"]; ok {
		presXML := readZipFile(presF)
		if m := regexp.MustCompile(`<p:sldSz[^>]*cx="(\d+)"[^>]*cy="(\d+)"`).FindStringSubmatch(presXML); m != nil {
			if v, err := strconv.Atoi(m[1]); err == nil && v > 0 {
				slideWidthEMU = v
			}
			if v, err := strconv.Atoi(m[2]); err == nil && v > 0 {
				slideHeightEMU = v
			}
		}
	}

	if f, ok := files["docProps/core.xml"]; ok {
		data := readZipFile(f)
		if m := regexp.MustCompile(`<dc:creator>(.*?)</dc:creator>`).FindStringSubmatch(data); m != nil {
			result.Metadata.Author = m[1]
		}
		if m := regexp.MustCompile(`<dc:title>(.*?)</dc:title>`).FindStringSubmatch(data); m != nil {
			result.Metadata.Title = m[1]
		}
	}
	if f, ok := files["docProps/app.xml"]; ok {
		data := readZipFile(f)
		if m := regexp.MustCompile(`<Slides>(\d+)</Slides>`).FindStringSubmatch(data); m != nil {
			result.Metadata.SlideCount = m[1]
		}
		if m := regexp.MustCompile(`<Application>(.*?)</Application>`).FindStringSubmatch(data); m != nil {
			result.Metadata.Application = m[1]
		}
		if m := regexp.MustCompile(`<Company>(.*?)</Company>`).FindStringSubmatch(data); m != nil {
			result.Metadata.Company = m[1]
		}
	}

	os.MkdirAll("output_media", 0755)

	for i := 1; ; i++ {
		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", i)
		f, ok := files[slidePath]
		if !ok {
			break
		}
		xmlStr := readZipFile(f)

		relsPath := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", i)
		rels := map[string]Relationship{}
		if rf, ok := files[relsPath]; ok {
			relsXML := readZipFile(rf)
			relRe := regexp.MustCompile(`<Relationship[^>]*Id="([^"]+)"[^>]*Type="([^"]+)"[^>]*Target="([^"]+)"[^>]*\/?>`)
			for _, m := range relRe.FindAllStringSubmatch(relsXML, -1) {
				rels[m[1]] = Relationship{Id: m[1], Type: m[2], Target: m[3]}
			}
		}

		var links []Link
		var mediaList []Media
		var textBlocks []TextElement

		spRe := regexp.MustCompile(`(?s)<p:sp.*?>.*?</p:sp>`)
		for _, sp := range spRe.FindAllString(xmlStr, -1) {
			text := ""
			if m := regexp.MustCompile(`<a:t>(.*?)</a:t>`).FindAllStringSubmatch(sp, -1); m != nil {
				for _, mm := range m {
					text += mm[1]
				}
			}
			href := ""
			if m := regexp.MustCompile(`a:hlinkClick[^>]*r:id="(rId\d+)"`).FindStringSubmatch(sp); m != nil {
				if rel, ok := rels[m[1]]; ok {
					href = rel.Target
				}
			}
			x, y, w, h, _, _, _ := getCoordsNear(xmlStr, strings.Index(xmlStr, sp)+len(sp)/2)
			if href != "" || text != "" {
				links = append(links, Link{Href: href, Text: text, X: x, Y: y, Width: w, Height: h})
			}

			if text != "" {
				textBlocks = append(textBlocks, TextElement{Text: text, X: x, Y: y, Width: w, Height: h})
			}
		}

		for _, match := range rIdRe.FindAllStringSubmatchIndex(xmlStr, -1) {
			if len(match) < 4 {
				continue
			}
			rid := xmlStr[match[2]:match[3]]
			matchPos := match[0]
			x, y, w, h, rotation, flipH, flipV := getCoordsNear(xmlStr, matchPos)

			if rel, ok := rels[rid]; ok {
				ltype := strings.ToLower(rel.Type)
				ext := strings.ToLower(filepath.Ext(rel.Target))
				mediaKind := ""
				if strings.Contains(ltype, "image") || strings.Contains(ext, "png") || strings.Contains(ext, "jpg") {
					mediaKind = "image"
				} else if strings.Contains(ltype, "video") || strings.Contains(ext, "mp4") {
					mediaKind = "video"
				} else if strings.Contains(ltype, "audio") || strings.Contains(ext, "mp3") {
					mediaKind = "audio"
				}

				if mediaKind != "" {
					local := exportFile(rel.Target, files, "output_media")
					if local != "" {
						mediaList = append(mediaList, Media{
							File:     local,
							Type:     mediaKind,
							X:        x,
							Y:        y,
							Width:    w,
							Height:   h,
							Rotation: rotation,
							FlipH:    flipH,
							FlipV:    flipV,
						})
					}
				}
			}
		}

		result.Slides = append(result.Slides, Slide{
			Slide:      fmt.Sprint(i),
			Links:      links,
			Media:      mediaList,
			TextBlocks: textBlocks,
		})
	}

	out, _ := json.MarshalIndent(result, "", "  ")
	return os.WriteFile(outJSON, out, 0644)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run pptx_extractor.go <input.pptx> <output.json>")
		return
	}
	pptxPath := os.Args[1]
	outJSON := os.Args[2]

	cleanedPPTX := "cleaned_" + filepath.Base(pptxPath)
	if err := cleanAndRepackPPTX(pptxPath, cleanedPPTX); err != nil {
		fmt.Println("Error cleaning PPTX:", err)
		return
	}

	if err := extractPPTX(pptxPath, outJSON); err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Extracted metadata written to", outJSON)
	}

}
