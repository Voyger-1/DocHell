package idk

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func ConvertToPDF(pptxFile, pdfFile string) error {
	outDir := filepath.Dir(pdfFile)
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(outDir, 0755); mkErr != nil {
			return fmt.Errorf("failed to create pdf output dir: %w", mkErr)
		}
	}

	if _, err := os.Stat(pdfFile); err == nil {
		log.Println("✅ PDF already exists:", pdfFile)
		return nil
	}

	cmd := exec.Command("libreoffice",
		"--headless",
		"--convert-to", "pdf",
		"--outdir", outDir,
		pptxFile,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Println("🔄 Converting PPTX to PDF...")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to convert to PDF: %w", err)
	}

	if _, err := os.Stat(pdfFile); err != nil {
		return fmt.Errorf("conversion finished but PDF not found: %s", pdfFile)
	}

	return nil
}

func ConvertPdfToPng(pdfFile, slidesDir string, slide int) (string, error) {
	if _, err := os.Stat(slidesDir); os.IsNotExist(err) {
		if mkErr := os.MkdirAll(slidesDir, 0755); mkErr != nil {
			return "", fmt.Errorf("failed to create slides dir: %w", mkErr)
		}
	}

	outPath := filepath.Join(slidesDir, fmt.Sprintf("slide%d.png", slide))

	if _, err := os.Stat(outPath); err == nil {
		log.Printf("✅ Slide %d PNG already exists: %s", slide, outPath)
		return outPath, nil
	}

	cmd := exec.Command("pdftoppm",
		"-f", fmt.Sprint(slide),
		"-l", fmt.Sprint(slide),
		"-png",
		"-singlefile",
		pdfFile,
		outPath[:len(outPath)-4],
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("🔄 Converting slide %d to PNG...", slide)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to convert slide %d: %w", slide, err)
	}

	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("conversion finished but PNG not found: %s", outPath)
	}

	return outPath, nil
}
