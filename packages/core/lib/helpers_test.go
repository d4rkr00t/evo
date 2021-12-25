package lib_test

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/otiai10/copy"
)

func RestoreFixture(name string) string {
	var fixtures_folder = "__fixtures__"
	var testfs_folder = "__testfs__"
	var temp_dir, _ = ioutil.TempDir(testfs_folder, "*")
	copy.Copy(path.Join(fixtures_folder, name), temp_dir)
	var os_cwd, _ = os.Getwd()
	return path.Join(os_cwd, temp_dir)
}

func CleanFixture(name string) {
	os.RemoveAll(name)
}
