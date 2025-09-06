const express = require("express");
const fs = require("fs");
const path = require("path");
const util = require("util");
const { exec } = require("child_process");
const axios = require("axios");
const cors = require("cors");

const run = util.promisify(exec);
const app = express();
app.use(cors());
app.use(express.static(__dirname));

app.use(express.json());
const PORT = 5000;

// SSE middleware
function sseMiddleware(req, res, next) {
  res.setHeader("Content-Type", "text/event-stream");
  res.setHeader("Cache-Control", "no-cache");
  res.setHeader("Connection", "keep-alive");
  res.flushHeaders();

  res.sseSend = (data) => {
    res.write(`data: ${JSON.stringify(data)}\n\n`);
  };

  next();
}

// Steps
async function downloadPptx(pptxUrl, pptxFile, res) {
  if (fs.existsSync(pptxFile)) {
    res.sseSend({ step: "PPTX already cached" });
    return;
  }
  res.sseSend({ step: "Downloading PPTX..." });
  const response = await axios.get(pptxUrl, { responseType: "arraybuffer" });
  fs.writeFileSync(pptxFile, response.data);
}

async function extractMeta(pptxFile, jsonFile, res) {
  if (fs.existsSync(jsonFile)) {
    res.sseSend({ step: "Meta already extracted" });
    return;
  }
  res.sseSend({ step: "Extracting meta.json..." });
  await run(`go run main.go 4. Graphs of Quadratic Equations (1) "${jsonFile}"`);
}

async function convertToPdf(pptxFile, pdfFile, res) {
  // if (fs.existsSync(pdfFile)) {
  //   res.sseSend({ step: "PDF already exists" });
  //   return;
  // }
  res.sseSend({ step: "Converting to PDF..." });
  await run(
    `libreoffice --headless --convert-to pdf --outdir "${__dirname}" "${pptxFile}"`
  );
}

async function convertPdfToPng(pdfFile, slidesDir, slide, res) {
  res.sseSend({ step: `Converting slide ${slide} to PNG...` });
  if (!fs.existsSync(slidesDir)) {
    fs.mkdirSync(slidesDir, { recursive: true });
  }
  const outPath = path.join(slidesDir, `slide${slide}.png`);
  if (!fs.existsSync(outPath)) {
    await run(
      `pdftoppm -f ${slide} -l ${slide} -r 150 -png -singlefile "${pdfFile}" "${outPath.replace(
        /\.png$/,
        ""
      )}"`
    );
  }
  return outPath;
}

function generateSlideDiv(jsonFile, slidesDir, slide, res) {
  res.sseSend({ step: `Generating HTML div for slide ${slide}` });
  const meta = JSON.parse(fs.readFileSync(jsonFile, "utf-8"));

  const slideMeta = meta.slides.find((s) => s.slide == slide);
  if (!slideMeta) throw new Error(`Slide ${slide} not found in meta.json`);

  const imgPath = `slides/slide${slide}.png`;

  let overlays = "";

  if (slideMeta.links) {
    slideMeta.links.forEach((link) => {
      overlays += `<a href="${link.href}" target="_blank" style="position:absolute;left:${link.x}px;top:${link.y}px;width:${link.width}px;height:${link.height}px;opacity:0.3"></a>`;
    });
  }

  if (slideMeta.media) {
    slideMeta.media.forEach((m) => {
      let transform = "";
      if (m.rotation) {
        transform = `transform: rotate(${m.rotation}deg);`;
      }
      
 
      
      if (m.file.endsWith(".mp4")) {
        overlays += `
          <video src="${m.file}" controls
            style="position:absolute;left:${m.x}px;top:${m.y}px;width:${m.width}px;height:${m.height}px;opacity:1;z-index:25">
          </video>`;
      } else {
        overlays += `
          <img src="${m.file}"
            style="position:absolute;left:${m.x}px;top:${Number(m.y)}px;width:${m.width}px;height:${m.height}px;opacity:1;${transform};z-index:25;transform-origin: center center;scale:${(m.flipH || m.flipV) && Boolean(m.rotation) ? 0: 1};${m.flipH ? 'scaleX(-1);' : ''}${m.flipV ? 'scaleY(-1);' : ''}">`;
      }
    });
  }

  // Add link overlays
  if (slideMeta.links) {
    slideMeta.links.forEach((link) => {
      if (link.href) {
        overlays += `<a href="${link.href}" target="_blank" style="position:absolute;left:${link.x}px;top:${link.y}px;width:${link.width}px;height:${link.height}px;opacity:0.3;z-index:20;pointer-events:auto;background:rgba(255,255,255,0.1);border:1px solid rgba(255,255,255,0.3)"></a>`;
      }
    });
  }

  return `
    <div style="position:relative;display:inline-block;margin:10px;">
      <img src="${imgPath}" style="display:block;height:720px;width:1280px">
      ${overlays}
    </div>
  `;
}

// API Route (range of slides, SSE)
app.get("/convert", sseMiddleware, async (req, res) => {
  const pptxUrl = req.query.url;
  const start = parseInt(req.query.start || req.query.slide || "1", 10);
  const end = parseInt(req.query.end || start, 10);

  if (!pptxUrl) {
    res.sseSend({ error: "Missing ?url=" });
    return res.end();
  }

  const pptxFile = path.join(__dirname, "example.pptx");
  const pdfFile = path.join(__dirname, "cleaned_4. Graphs of Quadratic Equations (1).pdf");
  const jsonFile = path.join(__dirname, "meta.json");
  const slidesDir = path.join(__dirname, "slides");
  const reconstructed = path.join(__dirname, "cleaned_4. Graphs of Quadratic Equations (1).pptx");

  try {
    await downloadPptx(pptxUrl, pptxFile, res);
    await extractMeta(pptxFile, jsonFile, res);
    await convertToPdf(reconstructed, pdfFile, res);

    for (let slide = start; slide <= end; slide++) {
      try {
        const pngPath = await convertPdfToPng(pdfFile, slidesDir, slide, res);
        const divHtml = generateSlideDiv(jsonFile, slidesDir, slide, res);

        res.sseSend({
          step: `Slide ${slide} done`,
          success: true,
          slide,
          div: divHtml,
          img: pngPath,
        });
      } catch (err) {
        res.sseSend({ error: `Slide ${slide} failed: ${err.message}` });
      }
    }

    res.sseSend({ step: "✅ All slides processed", done: true });
    
    // Clean up files safely
    try {
      // if (fs.existsSync(pptxFile)) fs.unlinkSync(pptxFile);
      // if (fs.existsSync(pdfFile)) fs.unlinkSync(pdfFile);
      // if (fs.existsSync(jsonFile)) fs.unlinkSync(jsonFile);
    } catch (cleanupErr) {
      console.log("Cleanup warning:", cleanupErr.message);
    }

    res.end();
  } catch (err) {
    console.error(err);
    res.sseSend({ error: err.message });
    res.end();
  }
});

app.listen(PORT, () => {
});
