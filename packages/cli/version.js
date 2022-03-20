const path = require("path");
const fs = require("fs");
const pkgJson = require("./package.json")

const versionGo = `package lib

var Version string = "${pkgJson.version}"
`;

fs.writeFileSync(path.join(__dirname, "lib", "version.go"), versionGo, "utf-8")
