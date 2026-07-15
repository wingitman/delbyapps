package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	dir := "/home/wing/.config/delbysoft"

	entries, err := OSReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e, ".toml") && !strings.Contains(e, "meta") {
			fmt.Println(strings.Replace(e, ".toml", "", -1))
		}
	}
}

// Source - https://stackoverflow.com/a/49196644
// Posted by manigandand, modified by community. See post 'Timeline' for change history
// Retrieved 2026-07-15, License - CC BY-SA 4.0

func OSReadDir(root string) ([]string, error) {
	var files []string
	f, err := os.Open(root)
	if err != nil {
		return files, err
	}
	fileInfo, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		return files, err
	}

	for _, file := range fileInfo {
		files = append(files, file.Name())
	}
	return files, nil
}
