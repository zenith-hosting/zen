package zencli

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

const starterTemplateRoot = "init_template"

//go:embed all:init_template
var starterTemplateFS embed.FS

func starterFiles() map[string]string {
	files := make(map[string]string)

	err := fs.WalkDir(starterTemplateFS, starterTemplateRoot, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		contents, err := starterTemplateFS.ReadFile(filePath)
		if err != nil {
			return err
		}

		files[starterOutputPath(filePath)] = string(contents)
		return nil
	})
	if err != nil {
		panic(err)
	}

	return files
}

func starterOutputPath(filePath string) string {
	relativePath := path.Clean(strings.TrimPrefix(filePath, starterTemplateRoot+"/"))
	if before, ok := strings.CutSuffix(relativePath, ".template"); ok {
		return before
	}

	return relativePath
}
