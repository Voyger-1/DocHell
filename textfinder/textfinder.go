package textfinder

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
)

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
	// fmt.Printf("Processing slide %d: Found %d shapes, %d groups, %d graphic frames\n",
	// 	slideIndex, len(s.SpTree.Sp), len(s.SpTree.GrpSp), len(s.SpTree.GraphicFrame))

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
