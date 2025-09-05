const fs = require("fs");
const JSZip = require("jszip");
const xml2js = require("xml2js");
const axios = require("axios");
const path = require("path");

const emuToPx = (emuValue) =>
  Math.round((parseInt(emuValue, 10) / 914400) * 96);

function extractTextFromShape(shape) {
  let text = "";
  const txBody = shape["p:txBody"]?.[0];
  if (txBody) {
    const paragraphs = txBody["a:p"] || [];
    paragraphs.forEach((para) => {
      const runs = para["a:r"] || [];
      runs.forEach((run) => {
        if (run["a:t"]) {
          text += run["a:t"].join("");
        }
      });
    });
  }
  return text.trim();
}

function collectHyperlinksRecursively(obj, relsMap, ancestors = []) {
  const results = [];

  if (Array.isArray(obj)) {
    obj.forEach((child) => {
      results.push(
        ...collectHyperlinksRecursively(child, relsMap, [...ancestors, obj])
      );
    });
    return results;
  }

  if (typeof obj !== "object" || obj === null) return results;

  if (obj["a:hlinkClick"]) {
    const hyperlinkId = obj["a:hlinkClick"]?.[0]?.$?.["r:id"];
    if (hyperlinkId && relsMap[hyperlinkId]) {
      const shapeAncestor =
        ancestors.reverse().find((a) => a["p:spPr"] || a["p:txBody"]) || obj;

      const xfrm = shapeAncestor["p:spPr"]?.[0]?.["a:xfrm"]?.[0];
      const off = xfrm?.["a:off"]?.[0]?.$ || {};
      const ext = xfrm?.["a:ext"]?.[0]?.$ || {};

      results.push({
        href: relsMap[hyperlinkId],
        text: extractTextFromShape(shapeAncestor),
        x: off.x ? emuToPx(off.x) : null,
        y: off.y ? emuToPx(off.y) : null,
        width: ext.cx ? emuToPx(ext.cx) : null,
        height: ext.cy ? emuToPx(ext.cy) : null,
      });
    }
  }

  Object.keys(obj).forEach((key) => {
    results.push(
      ...collectHyperlinksRecursively(obj[key], relsMap, [...ancestors, obj])
    );
  });

  return results;
}

function collectImagesRecursively(obj, relsMap, ancestors = []) {
  const results = [];

  if (Array.isArray(obj)) {
    obj.forEach((child) => {
      results.push(
        ...collectImagesRecursively(child, relsMap, [...ancestors, obj])
      );
    });
    return results;
  }

  if (typeof obj !== "object" || obj === null) return results;

  if (obj["p:blipFill"]) {
    const blip = obj["p:blipFill"][0]["a:blip"]?.[0];
    const embedId = blip?.$?.["r:embed"];

    if (embedId && relsMap[embedId]) {
      const shapeAncestor =
        ancestors.reverse().find((a) => a["p:spPr"]) || obj;

      const xfrm = shapeAncestor["p:spPr"]?.[0]?.["a:xfrm"]?.[0];
      const off = xfrm?.["a:off"]?.[0]?.$ || {};
      const ext = xfrm?.["a:ext"]?.[0]?.$ || {};

      results.push({
        src: relsMap[embedId], // maps to ppt/media/imageX.png
        x: off.x ? emuToPx(off.x) : null,
        y: off.y ? emuToPx(off.y) : null,
        width: ext.cx ? emuToPx(ext.cx) : null,
        height: ext.cy ? emuToPx(ext.cy) : null,
      });
    }
  }

  Object.keys(obj).forEach((key) => {
    results.push(
      ...collectImagesRecursively(obj[key], relsMap, [...ancestors, obj])
    );
  });

  return results;
}

class PPTXExtractor {
  constructor(fileUrl, downloadPath, outputImageDir = "output_images") {
    this.fileUrl = fileUrl;
    this.downloadPath = downloadPath;
    this.outputImageDir = outputImageDir;
    this.zip = null;
    this.parser = new xml2js.Parser();
  }

  async downloadFile() {
    const response = await axios.get(this.fileUrl, {
      responseType: "arraybuffer",
    });

    const dir = path.dirname(this.downloadPath);
    if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });

    fs.writeFileSync(this.downloadPath, response.data);
  }

  async load() {
    if (!fs.existsSync(this.downloadPath)) {
      throw new Error(`File not found: ${this.downloadPath}`);
    }
    const buffer = fs.readFileSync(this.downloadPath);
    this.zip = await JSZip.loadAsync(buffer);
  }

  async extractSlides() {
    if (!this.zip) {
      throw new Error("PPTX file not loaded. Please call load() first.");
    }

    const slides = Object.keys(this.zip.files).filter((name) =>
      name.match(/^ppt\/slides\/slide\d+\.xml$/)
    );

    const slideData = [];

    for (let slidePath of slides) {
      try {
        const xml = await this.zip.files[slidePath].async("text");
        const slideJson = await this.parser.parseStringPromise(xml);

        const relPath =
          slidePath.replace("slides/slide", "slides/_rels/slide") + ".rels";
        let relsMap = {};
        if (this.zip.files[relPath]) {
          const relsXml = await this.zip.files[relPath].async("text");
          const relsJson = await this.parser.parseStringPromise(relsXml);
          if (relsJson?.Relationships?.Relationship) {
            relsJson.Relationships.Relationship.forEach((rel) => {
              relsMap[rel.$.Id] = rel.$.Target;
            });
          }
        }

        const spTree = slideJson?.["p:sld"]?.["p:cSld"]?.[0]?.["p:spTree"]?.[0];
        const links = spTree ? collectHyperlinksRecursively(spTree, relsMap) : [];
        const images = spTree ? collectImagesRecursively(spTree, relsMap) : [];

        slideData.push({
          slide: slidePath.match(/slide(\d+)\.xml/)[1],
          links,
          images,
        });
      } catch (error) {
        console.error(`Error processing slide: ${slidePath}, ${error.message}`);
      }
    }

    return slideData;
  }

  async getMetadata() {
    if (!this.zip) {
      throw new Error("PPTX file not loaded. Please call load() first.");
    }

    const metadata = {};

    if (this.zip.files["docProps/core.xml"]) {
      const coreXml = await this.zip.files["docProps/core.xml"].async("text");
      const coreJson = await this.parser.parseStringPromise(coreXml);

      const core = coreJson["cp:coreProperties"] || {};
      metadata.author = core["dc:creator"]?.[0] || "";
      metadata.title = core["dc:title"]?.[0] || "";
      metadata.subject = core["dc:subject"]?.[0] || "";
      metadata.created = core["dcterms:created"]?.[0]?._ || "";
      metadata.modified = core["dcterms:modified"]?.[0]?._ || "";
    }

    if (this.zip.files["docProps/app.xml"]) {
      const appXml = await this.zip.files["docProps/app.xml"].async("text");
      const appJson = await this.parser.parseStringPromise(appXml);

      const app = appJson.Properties || {};
      metadata.slideCount = app.Slides?.[0] || 0;
      metadata.application = app.Application?.[0] || "";
      metadata.company = app.Company?.[0] || "";
    }

    return metadata;
  }

  async saveImages() {
    if (!this.zip) {
      throw new Error("PPTX file not loaded. Please call load() first.");
    }

    if (!fs.existsSync(this.outputImageDir)) {
      fs.mkdirSync(this.outputImageDir, { recursive: true });
    }

    const mediaFiles = Object.keys(this.zip.files).filter((f) =>
      f.startsWith("ppt/media/")
    );

    for (let file of mediaFiles) {
      const data = await this.zip.files[file].async("nodebuffer");
      const outputPath = path.join(this.outputImageDir, path.basename(file));
      fs.writeFileSync(outputPath, data);
    }
  }

  async saveToMetaJSON(fileName = "meta.json") {
    const [slides, metadata] = await Promise.all([
      this.extractSlides(),
      this.getMetadata(),
    ]);

    const metaJson = { metadata, slides };

    fs.writeFileSync(fileName, JSON.stringify(metaJson, null, 2));
    return metaJson;
  }

  cleanUp() {
    try {
      fs.unlinkSync(this.downloadPath);
    } catch {}
  }
}

module.exports = PPTXExtractor;
