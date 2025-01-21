package utils

import (
	"path"
	"strings"
)

func IsValidImageFile(filename string) bool {
	ext := strings.ToLower(path.Ext(filename))
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	return validExts[ext]
}
