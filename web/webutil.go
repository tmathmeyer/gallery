package web

import (
	"net/http"
	"os"
	"strings"
)

// Test if file exists
func FExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func GetPathParts(r *http.Request, str string, splitct int) []string {
	restPath := r.URL.Path[len(str):]
	return strings.SplitN(restPath, "/", splitct)
}
