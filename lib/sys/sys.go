package sys

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
)

// MakeDir a directory out of a given unless it already exists.
func MakeDir(dir string) error {
	if strings.HasPrefix(dir, "~/") || strings.HasPrefix(dir, "~\\") {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		dir = filepath.Join(home, dir[2:])
	}
	path, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	if path.Mode().IsRegular() {
		return fmt.Errorf("'%s' is a file, not a directory", dir)
	}
	return nil
}

// FileExists checks if a file exists, and that it is a regular file.
func FileExists(f string) bool {
	path, err := os.Stat(f)
	if os.IsNotExist(err) {
		return false
	}
	if !path.Mode().IsRegular() {
		log.Fatal(fmt.Errorf("'%s' is not a regular file, "+
			"delete or move it and try again.", f))
	}
	return true
}

// CleanDir removes all files from a directory.
func CleanDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
