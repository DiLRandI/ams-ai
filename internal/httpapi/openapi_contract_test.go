package httpapi

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v3"
)

type openAPIDoc struct {
	Paths map[string]map[string]any `yaml:"paths"`
}

func TestOpenAPIContractMatchesRouter(t *testing.T) {
	implemented := implementedRoutes(t)
	documented := documentedRoutes(t)

	for route := range implemented {
		if !documented[route] {
			t.Errorf("implemented route missing from docs/openapi.yaml: %s", route)
		}
	}
	for route := range documented {
		if !implemented[route] {
			t.Errorf("docs/openapi.yaml documents unimplemented route: %s", route)
		}
	}
}

func implementedRoutes(t *testing.T) map[string]bool {
	t.Helper()
	files := []string{
		"router.go",
		"../assets/http.go",
		"../auth/http.go",
		"../categories/http.go",
		"../documents/http.go",
		"../reminders/http.go",
		"../reports/http.go",
		"../vehicles/http.go",
	}
	routes := make(map[string]bool)
	pattern := regexp.MustCompile(`mux\.HandleFunc\("([A-Z]+) ([^"]+)"`)
	for _, file := range files {
		source, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		for _, match := range pattern.FindAllSubmatch(source, -1) {
			routes[string(match[1])+" "+string(match[2])] = true
		}
	}
	if len(routes) == 0 {
		t.Fatal("no backend routes found")
	}
	return routes
}

func documentedRoutes(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile("../../docs/openapi.yaml")
	if err != nil {
		t.Fatalf("read docs/openapi.yaml: %v", err)
	}
	var doc openAPIDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("parse docs/openapi.yaml: %v", err)
	}
	if len(doc.Paths) == 0 {
		t.Fatal("docs/openapi.yaml contains no paths")
	}
	routes := make(map[string]bool)
	for path, methods := range doc.Paths {
		for method := range methods {
			method = strings.ToUpper(method)
			switch method {
			case "GET", "POST", "PUT", "DELETE":
				routes[method+" "+path] = true
			}
		}
	}
	return routes
}
