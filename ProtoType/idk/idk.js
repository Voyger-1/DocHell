(async () => {
  const PPTXExtractor = require("./stuff");
  const extractor = new PPTXExtractor(
    "https://storage.googleapis.com/gotoschool/assess_test/content_data/questionwiseweeklytest/2025/09/01/1756708059823/4. Graphs of Quadratic Equations (1).pptx",
    "downloads/sample.pptx",
    "output_images"
  );

  
  await extractor.downloadFile();
  await extractor.load();

  const result = await extractor.saveToMetaJSON("meta.json");
  await extractor.saveImages();

  console.log("Metadata + slide data:", result);
})();
