const path = require("path");
const fs = require("fs");
const pkgJson = require("./package.json")

const versionGo = `package version

var Version string = "${pkgJson.version}"
`;

fs.writeFileSync(path.join(__dirname, "cmd", "version", "version.go"), versionGo, "utf-8")
