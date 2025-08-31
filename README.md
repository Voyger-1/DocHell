# PowerPoint to HTML Converter

A comprehensive Go-based converter that transforms PowerPoint presentations into pixel-perfect HTML while preserving all formatting, shapes, text, images, equations, and more.

## Features

### ✅ **Comprehensive Element Support**
- **Text Elements**: Fonts, sizes, colors, bold/italic/underline, bullet points, numbering
- **Shapes**: All PowerPoint shapes with exact geometry, fills, strokes, and effects
- **Images**: Full image support with cropping, scaling, and positioning
- **Equations**: MathML/OMML to LaTeX conversion with MathJax rendering
- **Tables**: Complex table structures with cell merging and styling
- **Charts**: Chart data extraction and visualization
- **Links**: Hyperlinks with proper styling
- **Groups**: Nested element grouping with transformations
- **Connectors**: Line connectors and arrows

### ✅ **Exact Formatting Preservation**
- **Colors**: RGB, scheme colors, preset colors, gradients
- **Positioning**: Pixel-perfect positioning using EMU to pixel conversion
- **Effects**: Shadows, glows, reflections, soft edges
- **Borders**: Line styles, widths, colors, caps, joins
- **Transforms**: Rotation, scaling, skewing
- **Layering**: Proper z-index ordering

### ✅ **Multi-Format Support**
- Microsoft PowerPoint (.pptx)
- Google Slides (exported as .pptx)
- LibreOffice Impress (.pptx)
- Any OOXML-compliant presentation format

## Usage

### Basic Usage
```bash
go run main.go presentation.pptx 1
```

### Advanced Usage
```go
package main

import (
    "fmt"
    "log"
)

func main() {
    // Convert a specific slide
    html, err := ConvertPPTXToHTMLSimple("presentation.pptx", 1)
    if err != nil {
        log.Fatal(err)
    }
    
    // Save to file
    err = os.WriteFile("slide.html", []byte(html), 0644)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Conversion completed successfully!")
}
```

## Architecture

### Core Components

1. **PPTXConverter** - Main conversion orchestrator
2. **Element Extractors** - Specialized parsers for different element types
3. **HTML Generators** - Element-specific HTML output generators
4. **Style Processors** - Formatting and styling preservation

### Element Processing Pipeline

```
PPTX File → XML Parsing → Element Extraction → Style Processing → HTML Generation
```

### Supported Element Types

| Element Type | Description | Features |
|--------------|-------------|----------|
| **Text** | Text boxes, paragraphs, runs | Fonts, colors, effects, bullets |
| **Shapes** | Geometric shapes, connectors | Fills, strokes, effects, transforms |
| **Images** | Pictures, photos, graphics | Cropping, scaling, positioning |
| **Tables** | Data tables, grids | Cell merging, borders, styling |
| **Math** | Equations, formulas | OMML to LaTeX conversion |
| **Charts** | Data visualizations | Chart types, data, styling |
| **Links** | Hyperlinks | URLs, targets, styling |
| **Groups** | Element collections | Nesting, transformations |

## Output

### HTML Structure
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>PowerPoint Slide</title>
    <script src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js"></script>
    <style>
        /* Comprehensive CSS for exact formatting */
        .slide-container {
            width: 960px;
            height: 720px;
            position: relative;
        }
        .element {
            position: absolute;
            box-sizing: border-box;
        }
        /* Element-specific styles */
    </style>
</head>
<body>
    <div class="slide-container">
        <div class="slide-background"></div>
        <div class="slide-content">
            <!-- All extracted elements with exact positioning -->
        </div>
    </div>
</body>
</html>
```

### CSS Features
- **Absolute Positioning**: Pixel-perfect element placement
- **Z-Index Layering**: Proper element stacking order
- **Responsive Design**: Maintains aspect ratios
- **Modern CSS**: Flexbox, Grid, CSS Variables
- **Math Rendering**: MathJax integration for equations

## Installation

### Prerequisites
- Go 1.19 or higher
- Git

### Setup
```bash
git clone <repository-url>
cd pptXhell
go mod tidy
go build
```

## Examples

### Simple Text Slide
```bash
# Convert slide 1 of a presentation
go run main.go presentation.pptx 1
```

### Complex Slide with All Elements
```bash
# Convert slide 5 with shapes, images, and tables
go run main.go complex_presentation.pptx 5
```

## Technical Details

### EMU to Pixel Conversion
- **EMU**: English Metric Units (1 inch = 914400 EMU)
- **Conversion**: `pixels = EMU / 914400.0 * 96.0`
- **Precision**: Maintains exact positioning and sizing

### XML Parsing Strategy
- **Streaming**: Efficient memory usage for large files
- **Namespace Handling**: Supports multiple XML namespaces
- **Error Recovery**: Graceful handling of malformed XML

### Performance Optimizations
- **Lazy Loading**: Only processes requested slides
- **Memory Management**: Efficient handling of large presentations
- **Parallel Processing**: Concurrent element extraction where possible

## Limitations

### Current Constraints
- **Single Slide**: Processes one slide at a time
- **File Size**: Large presentations may require more memory
- **Complex Effects**: Some advanced PowerPoint effects may be simplified

### Future Enhancements
- **Multi-Slide Support**: Convert entire presentations
- **Animation Support**: Slide transitions and element animations
- **Interactive Elements**: Clickable elements and forms
- **Export Formats**: PDF, SVG, PNG output options

## Contributing

### Development Setup
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Testing
```bash
go test ./...
go test -v ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- **MathJax**: For mathematical equation rendering
- **Go Standard Library**: For XML parsing and ZIP handling
- **PowerPoint Community**: For format specifications and examples

---

**Note**: This converter is designed to handle the most complex PowerPoint presentations while maintaining exact visual fidelity. It's suitable for professional use cases where pixel-perfect conversion is required.