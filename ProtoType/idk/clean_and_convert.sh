#!/bin/bash

# Clean and Convert PPTX Workflow
# This script demonstrates the complete process

echo "🚀 Starting PPTX Clean and Convert Workflow"
echo "=============================================="

# Check if required tools are installed
command -v node >/dev/null 2>&1 || { echo "❌ Node.js is required but not installed. Aborting." >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo "❌ Go is required but not installed. Aborting." >&2; exit 1; }
command -v libreoffice >/dev/null 2>&1 || { echo "❌ LibreOffice is required but not installed. Aborting." >&2; exit 1; }
command -v pdftoppm >/dev/null 2>&1 || { echo "❌ pdftoppm is required but not installed. Aborting." >&2; exit 1; }

echo "✅ All required tools are installed"

# Check if input file exists
if [ $# -eq 0 ]; then
    echo "Usage: $0 <input.pptx>"
    echo "Example: $0 presentation.pptx"
    exit 1
fi

INPUT_PPTX="$1"
OUTPUT_DIR="cleaned_output"
META_JSON="meta.json"
CLEANED_PPTX="cleaned_presentation.pptx"
OUTPUT_PDF="cleaned_presentation.pdf"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo ""
echo "📋 Step 1: Extracting metadata from PPTX..."
go run main.go "$INPUT_PPTX" "$META_JSON"

if [ $? -ne 0 ]; then
    echo "❌ Failed to extract metadata"
    exit 1
fi

echo "✅ Metadata extracted to $META_JSON"

echo ""
echo "🧹 Step 2: Cleaning PPTX (removing overlapping text)..."
node clean_pptx.js "$INPUT_PPTX" "$META_JSON" "$CLEANED_PPTX"

if [ $? -ne 0 ]; then
    echo "❌ Failed to clean PPTX"
    exit 1
fi

echo "✅ PPTX cleaned and saved to $CLEANED_PPTX"

echo ""
echo "📄 Step 3: Converting cleaned PPTX to PDF..."
libreoffice --headless --convert-to pdf --outdir "$OUTPUT_DIR" "$CLEANED_PPTX"

if [ $? -ne 0 ]; then
    echo "❌ Failed to convert to PDF"
    exit 1
fi

echo "✅ PDF created: $OUTPUT_DIR/$OUTPUT_PDF"

echo ""
echo "🖼️  Step 4: Converting PDF to PNG slides..."
mkdir -p "$OUTPUT_DIR/slides"
pdftoppm -png -r 150 "$OUTPUT_DIR/$OUTPUT_PDF" "$OUTPUT_DIR/slides/slide"

if [ $? -ne 0 ]; then
    echo "❌ Failed to convert PDF to PNG"
    exit 1
fi

echo "✅ PNG slides created in $OUTPUT_DIR/slides/"

echo ""
echo "🧹 Cleaning up temporary files..."
rm -f "$META_JSON" "$CLEANED_PPTX"

echo ""
echo "🎉 Workflow completed successfully!"
echo "📁 Output files are in: $OUTPUT_DIR/"
echo "📄 PDF: $OUTPUT_DIR/$OUTPUT_PDF"
echo "🖼️  PNG slides: $OUTPUT_DIR/slides/"
echo ""
echo "💡 You can now use the cleaned slides without text overlap issues!"
