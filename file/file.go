package file

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/fatih/color"

	"github.com/nomad-software/vend/output"
)

// CopyDependencies copies dependencies listed in the module file into your
// vendor folder.
func CopyDependencies(mod GoMod, deps []Dep) {
	deleteVendorDir()
	var report string
dep:
	for _, r := range mod.Require {
		for _, d := range deps {
			if r.Path == d.Path && r.Version == d.Version {
				report += r.String() + "\n"

				if r.Indirect {
					fmt.Fprintf(output.Stdout, "%s %s\n", color.MagentaString("vend:"), r.String())
				} else {
					fmt.Fprintf(output.Stdout, "%s %s\n", color.GreenString("vend:"), r.String())
				}

				dest := path.Join(vendorDir(), d.Path)
				copy(d.Dir, dest)

				continue dep
			}
		}

		output.Error("No dependency available for %s (%s)", r.Path, r.Version)
	}

	SaveReport(report)
}

// SaveReport saves the report into the vendor directory.
func SaveReport(report string) {
	file := path.Join(vendorDir(), "vend.info")
	err := ioutil.WriteFile(file, []byte(report), 0644)
	output.OnError(err, "Error saving report")
}

// VendorDir returns the vendor directory in the current directory.
func vendorDir() string {
	wd, err := os.Getwd()
	output.OnError(err, "Error getting the current directory")
	return path.Join(wd, "vendor")
}

// deleteVendorDir deletes the vendor directory.
func deleteVendorDir() {
	err := os.RemoveAll(vendorDir())
	output.OnError(err, "Error removing vendor directory")
}

// Copy will copy files to the vendor directory.
func copy(src string, dest string) {
	info, err := os.Lstat(src)
	output.OnError(err, "Error getting information about source")

	if info.Mode()&os.ModeSymlink != 0 {
		return // Completely ignore symlinks.
	}

	if info.IsDir() {
		copyDirectory(src, dest)
	} else {
		copyFile(src, dest)
	}
}

// CopyDirectory will copy directories.
func copyDirectory(src string, dest string) {
	err := os.MkdirAll(dest, os.ModePerm)
	output.OnError(err, "Error creating directories")

	contents, err := ioutil.ReadDir(src)
	output.OnError(err, "Error reading source directory")

	for _, content := range contents {
		s := filepath.Join(src, content.Name())
		d := filepath.Join(dest, content.Name())
		copy(s, d)
	}
}

// CopyFile will copy files.
func copyFile(src string, dest string) {
	err := os.MkdirAll(filepath.Dir(dest), os.ModePerm)
	output.OnError(err, "Error creating directories")

	d, err := os.Create(dest)
	output.OnError(err, "Error creating file")
	defer d.Close()

	s, err := os.Open(src)
	output.OnError(err, "Error opening file")
	defer s.Close()

	_, err = io.Copy(d, s)
	output.OnError(err, "Error copying file")
}
