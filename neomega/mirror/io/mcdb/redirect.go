package mcdb

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
)

func CheckIsMCDBDir(filePath string) bool {
	if _, err := os.Stat(path.Join(filePath, "db")); err == nil {
		if _, err := os.Stat(path.Join(filePath, "level.dat")); err == nil {
			return true
		}
	}
	return false
}

func MCDBDirRedirect(filePath string) string {
	dataPath := filePath
	if !CheckIsMCDBDir(filePath) {
		filepath.Walk(filePath, func(newPath string, info fs.FileInfo, err error) error {
			if CheckIsMCDBDir(newPath) {
				fmt.Println("re-target path to " + newPath)
				dataPath = newPath
			}
			return nil
		})
	}
	return dataPath
}
