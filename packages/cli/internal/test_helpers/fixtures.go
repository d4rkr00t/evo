package test_helpers

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/otiai10/copy"
)

func RestoreFixture(name string) (func(), string, string) {
	var fixturesFolder = "__fixtures__"
	var testfsFolder = "__testfs__"
	var osCwd, _ = os.Getwd()
	var tempDirRel, _ = ioutil.TempDir(testfsFolder, "*")
	var tempDirAbs = path.Join(osCwd, tempDirRel)
	copy.Copy(path.Join(fixturesFolder, name), tempDirAbs)

	return func() {
		os.RemoveAll(tempDirAbs)
	}, tempDirRel, tempDirAbs
}
