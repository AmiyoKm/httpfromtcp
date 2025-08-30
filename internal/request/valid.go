package request

import "strings"

var validMethods = map[string]bool{
	"GET":     true,
	"POST":    true,
	"PUT":     true,
	"DELETE":  true,
	"HEAD":    true,
	"OPTIONS": true,
	"PATCH":   true,
}

func isValidMethod(method string) bool {
	return validMethods[method]
}

func isValidTarget(target string) bool {
	if !strings.HasPrefix(target, "/") {
		return false
	}
	if strings.Contains(target, "../") {
		return false
	}
	if len(target) > 2048 {
		return false
	}

	return true
}
