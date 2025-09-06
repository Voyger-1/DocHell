# PPTX Clean and Convert System

This system automatically removes overlapping text from PowerPoint presentations and converts them to clean PDF and PNG formats.

## 🎯 What It Does

Instead of trying to layer elements on top of each other, this system:
1. **Extracts metadata** from the PPTX (text positions, media positions)
2. **Identifies overlaps** between text and images
3. **Removes overlapping text** from the original PPTX
4. **Re-zips the cleaned PPTX**
5. **Converts to PDF** and **PNG** with clean results

## 🚀 Quick Start

### Option 1: Use the Shell Script (Recommended)
```bash
./clean_and_convert.sh your_presentation.pptx
```

### Option 2: Use the Web Interface
```bash
# Start the Node.js server
node is.js

# Open index.htm in your browser
# Enter PPTX URL and convert slides
```

### Option 3: Manual Steps
```bash
# 1. Extract metadata
go run main.go input.pptx meta.json

# 2. Clean PPTX
node clean_pptx.js input.pptx meta.json cleaned.pptx

# 3. Convert to PDF
libreoffice --headless --convert-to pdf cleaned.pptx

# 4. Convert PDF to PNG
pdftoppm -png -r 150 cleaned.pdf slide
```

## 📁 Files Overview

- **`main.go`** - Go script to extract PPTX metadata
- **`clean_pptx.js`** - Node.js script to clean PPTX files
- **`is.js`** - Node.js server for web interface
- **`index.htm`** - Web interface
- **`clean_and_convert.sh`** - Complete workflow script

## 🔧 Requirements

- **Go** (for metadata extraction)
- **Node.js** (for PPTX cleaning and web server)
- **LibreOffice** (for PPTX to PDF conversion)
- **pdftoppm** (for PDF to PNG conversion)

### Install Dependencies
```bash
# Node.js dependencies
npm install

# Go dependencies
go mod tidy
```

## 🎨 How It Works

### 1. Metadata Extraction (`main.go`)
- Parses PPTX XML structure
- Extracts text positions and content
- Extracts media positions and types
- Saves coordinates in JSON format

### 2. PPTX Cleaning (`clean_pptx.js`)
- Reads metadata JSON
- Identifies text that overlaps with media
- Removes overlapping text elements from slide XML
- Re-zips the cleaned PPTX

### 3. Conversion
- Converts cleaned PPTX to PDF
- Converts PDF to high-quality PNG slides
- Results in clean slides without text overlap

## 📊 Output

- **`meta.json`** - Metadata with text and media positions
- **`cleaned_presentation.pptx`** - PPTX with overlapping text removed
- **`cleaned_presentation.pdf`** - Clean PDF version
- **`slides/slide-*.png`** - Individual PNG slides

## 🎯 Benefits

✅ **No more text overlap issues**  
✅ **Clean, professional slides**  
✅ **Images with transparent backgrounds work perfectly**  
✅ **Maintains original layout and positioning**  
✅ **Automated workflow**  
✅ **Web interface for easy use**  

## 🔍 Example Use Cases

- **Educational content** - Remove overlapping text from math problems
- **Business presentations** - Clean up slides with image overlays
- **Design work** - Ensure images are properly visible
- **Content conversion** - Prepare slides for different formats

## 🚨 Troubleshooting

### Common Issues

1. **"LibreOffice not found"**
   ```bash
   sudo apt-get install libreoffice
   ```

2. **"pdftoppm not found"**
   ```bash
   sudo apt-get install poppler-utils
   ```

3. **"JSZip module not found"**
   ```bash
   npm install jszip
   ```

### Debug Mode
```bash
# Enable verbose logging
DEBUG=1 node is.js
```

## 🤝 Contributing

Feel free to submit issues and enhancement requests!

## 📄 License

This project is open source and available under the MIT License.
