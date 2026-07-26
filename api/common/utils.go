package common

import (
	"errors"
	"os"
	"regexp"

	"github.com/dattc/go-though-storm/api/models"
)

func Intersect(list1, list2 []string) []string {
	seen := make(map[string]struct{})
	for _, item := range list1 {
		seen[item] = struct{}{}
	}

	var result []string
	for _, item := range list2 {
		if _, exists := seen[item]; exists {
			result = append(result, item)
			delete(seen, item)
		}
	}

	return result
}

func IsFileEditable(filePath string, config models.EditableFilesConfig) bool {
	if !config.Enabled {
		return false
	}

	for _, pattern := range config.Patterns {
		matched, err := regexp.MatchString(pattern, filePath)
		if err == nil && matched {
			return true
		}
	}

	return false
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
	}
	return false
}
