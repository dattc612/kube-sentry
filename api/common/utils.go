package common

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/dattc/go-though-storm/api/models"
	"github.com/dlclark/regexp2/v2"
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

func MatchesAllPatterns(podName string, patterns []string) (bool, error) {
	if len(patterns) == 0 {
		return true, nil
	}

	for _, pattern := range patterns {
		re, err := regexp2.Compile(pattern, regexp2.None)
		if err != nil {
			return false, err
		}

		matched, err := re.MatchString(podName)
		if err != nil {
			return false, err
		}

		if !matched {
			return false, nil
		}
	}

	return true, nil
}

func MatchesAnyPattern(podName string, patterns []string) (bool, error) {
	if len(patterns) == 0 {
		return true, nil
	}

	for _, pattern := range patterns {
		re, err := regexp2.Compile(pattern, regexp2.None)
		if err != nil {
			return false, err
		}

		matched, err := re.MatchString(podName)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}

func IsAllowedPod(podName string, patterns []string) bool {
	for _, pattern := range patterns {
		m := regexp2.MustCompile(pattern, regexp2.None)
		matched, err := m.MatchString(podName)
		if (err == nil && !matched) || err != nil {
			return false
		}
	}
	return true
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

func FindItemByID(items []models.JobConfig, id int) (*models.JobConfig, error) {
	for i := range items {
		if items[i].ID == id {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("item with ID %d not found", id)
}

func GenerateJWTSecret() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal(err)
	}
	log.Printf("Generated JWT Secret: %s", hex.EncodeToString(key))
	return []byte(hex.EncodeToString(key))
}
