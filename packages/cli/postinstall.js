const path = require("path");
const fs = require("fs");
const mkdirp = require('mkdirp');
const exec = require('child_process').execSync;

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
const ARCH_MAPPING = {
  "ia32": "386",
  "x64": "amd64",
  "arm": "arm"
};

// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING = {
  "darwin": "darwin",
  "linux": "linux",
  "win32": "windows",
  "freebsd": "freebsd"
};

const BIN_NAME = "evo";

(async function install() {
  const binSrcPath = path.join(__dirname, `./dist/evo_${PLATFORM_MAPPING[process.platform]}_${ARCH_MAPPING[process.arch]}/${BIN_NAME}`);
  const localBinPath = path.join(__dirname, "bin");
  const localBinDestPath = path.join(localBinPath, BIN_NAME);

  if (!fs.existsSync(binSrcPath)) {
    console.error(`Installation is not supported for this architecture ${process.arch} or platform ${process.platform}.`);
    console.error(`Installation source ${binSrcPath} not found.`);
    process.exit(1);
  }

  await mkdirp(localBinPath);
  fs.copyFileSync(binSrcPath, localBinDestPath);
})()

