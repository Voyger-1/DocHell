const exec = require("child_process").exec;

const pptxFile = process.argv[2];
const num = process.argv[3];

(() => {
  if (!pptxFile || !num) {
    console.error("Usage: node index.js <pptxFile> <num>");
    process.exit(1);
  }

  const removal = `rm -rf *.html`;
  exec(removal, (error, stdout, stderr) => {
    if (error) {
      console.error(`exec error: ${error}`);
      return;
    }
    console.log(stdout);
  });

  const cmd = `./pptXhell ${pptxFile} ${num}`;
  exec(cmd, (error, stdout, stderr) => {
    if (error) {
      console.error(`exec error: ${error}`);
      return;
    }
    console.log(stdout);
  });
})();
