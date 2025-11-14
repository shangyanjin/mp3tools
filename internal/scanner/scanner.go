package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

var audioExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".aac":  true,
	".ogg":  true,
	".wma":  true,
}

type AudioFile struct {
	Path     string
	RelPath  string
	BasePath string
}

func ScanDirectory(rootPath string) ([]AudioFile, error) {
	var files []AudioFile
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(absRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if audioExtensions[ext] {
			relPath, err := filepath.Rel(absRoot, path)
			if err != nil {
				return err
			}
			files = append(files, AudioFile{
				Path:     path,
				RelPath:  relPath,
				BasePath: absRoot,
			})
		}
		return nil
	})

	return files, err
}

