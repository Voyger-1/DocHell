package main

import (
	"archive/zip"
	"fmt"
	"os"
	"strconv"

	"pptXhell/bgfinder"
	"pptXhell/helper"
	"pptXhell/imagefinder"
	mathfinder "pptXhell/math"
	"pptXhell/shapefinder"
	"pptXhell/textfinder"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run main.go <pptx-file><slideIndex>")
		return
	}
	pptxFile := os.Args[1]
	slideIndexStr := os.Args[2]
	l, _ := strconv.Atoi(slideIndexStr)
	// if slideIndex < 1 {
	// 	fmt.Println("slideIndex must be >= 1")
	// 	return
	// }

	r, err := zip.OpenReader(pptxFile)
	if err != nil {
		fmt.Println("open pptx:", err)
		return
	}
	defer r.Close()

	for slideIndex := 1; slideIndex < l; slideIndex++ {

		slidePath := fmt.Sprintf("ppt/slides/slide%d.xml", slideIndex)
		relsPath := fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", slideIndex)

		mediaPath := ""
		if _, mp, err := bgfinder.FindBackground(r, slidePath, relsPath, 0); err == nil {
			mediaPath = mp
		}

		bgURI := ""
		if mediaPath != "" {
			a := helper.NormalizeMediaPath(mediaPath)
			uri, err := helper.MediaDataURIFromZip(r, a)
			fmt.Println("Background media:", a, err)
			if err == nil {
				bgURI = uri
			}
		}

		boxes, err := textfinder.ExtractTextBoxes(r, slideIndex)
		if err != nil {
			fmt.Println("extract text:", err)
			return
		}

		mathBoxes, err := mathfinder.ExtractMath(r, slideIndex)
		if err != nil {
			fmt.Println("extract math:", err)
		}

		shapes, err := shapefinder.ExtractShapes(r, slideIndex)
		if err == nil {
			for _, s := range shapes {
				fmt.Printf("Shape: %s @ x=%d y=%d w=%d h=%d text=%q\n",
					s.Geom, s.X, s.Y, s.Cx, s.Cy, s.Text)
			}
		}

		//Need to Fix as Images are coming for all slides
		images, _ := imagefinder.ExtractImages(r, 6)

		html := helper.BuildSlideHTML(bgURI, boxes, mathBoxes, shapes, images)
		name := "slide" + strconv.Itoa(slideIndex+1) + ".html"
		if err := os.WriteFile(name, []byte(html), 0644); err != nil {
			fmt.Println("slide.html:", err)
			return
		}
	}
	fmt.Println("IDFK")
}
