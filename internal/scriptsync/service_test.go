package scriptsync

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	service := NewService(Config{})
	if service.cfg.CollectionsDir != "collections" {
		t.Errorf("CollectionsDir = %v, want collections", service.cfg.CollectionsDir)
	}
	if service.cfg.ScriptsDir != "scripts" {
		t.Errorf("ScriptsDir = %v, want scripts", service.cfg.ScriptsDir)
	}
}

func createTestCollection() map[string]any {
	return map[string]any{
		"info": map[string]any{"name": "Test"},
		"event": []any{
			map[string]any{
				"listen": "prerequest",
				"script": map[string]any{
					"exec": []any{"console.log('collection script');"},
				},
			},
		},
		"item": []any{
			map[string]any{
				"name": "Test Request",
				"event": []any{
					map[string]any{
						"listen": "test",
						"script": map[string]any{
							"exec": []any{"pm.test('test', function() {});"},
						},
					},
				},
			},
		},
	}
}

func setupTest(t *testing.T) (string, func()) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()

	cleanup := func() {
		if err := os.Chdir(oldCwd); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}

	if err := os.Chdir(tempDir); err != nil {
		cleanup()
		t.Fatalf("Failed to change directory: %v", err)
	}

	return tempDir, cleanup
}

func TestService_Extract(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection file
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := createTestCollection()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})
	err := service.Extract("test")

	if err != nil {
		t.Errorf("Extract failed: %v", err)
	}

	// Verify script files were created
	expectedFiles := []string{
		"scripts/test/_collection__prerequest.js",
		"scripts/test/test-request__test.js",
	}
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			t.Errorf("Expected file %s not created", file)
		}
	}
}

func TestService_Extract_CollectionNotFound(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	service := NewService(Config{})
	err := service.Extract("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent collection")
	}
}

func TestService_Build(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection and extract scripts first
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := createTestCollection()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})

	// Extract scripts first to create all necessary script files
	if err := service.Extract("test"); err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Update script file
	updatedScript := "console.log('updated script');"
	if err := os.WriteFile("scripts/test/_collection__prerequest.js", []byte(updatedScript), 0644); err != nil {
		t.Fatalf("Failed to write script: %v", err)
	}

	// Build with updated script
	err := service.Build("test")
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	// Verify collection was updated with script content
	updatedData, err := os.ReadFile("collections/test.postman_collection.json")
	if err != nil {
		t.Fatalf("Failed to read updated collection: %v", err)
	}
	if !strings.Contains(string(updatedData), "updated script") {
		t.Error("Collection not updated with script content")
	}
}

func TestService_Build_CollectionNotFound(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	service := NewService(Config{})
	err := service.Build("nonexistent")

	if err == nil {
		t.Error("Expected error for nonexistent collection")
	}
}

func TestService_ExtractAndBuild_Integration(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection file
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := createTestCollection()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})

	// Extract scripts
	if err := service.Extract("test"); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Edit extracted script
	updatedScript := "console.log('edited script');"
	if err := os.WriteFile("scripts/test/_collection__prerequest.js", []byte(updatedScript), 0644); err != nil {
		t.Fatalf("Failed to edit script: %v", err)
	}

	// Build collection with edited script
	if err := service.Build("test"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify collection contains edited script
	updatedData, err := os.ReadFile("collections/test.postman_collection.json")
	if err != nil {
		t.Fatalf("Failed to read updated collection: %v", err)
	}
	if !strings.Contains(string(updatedData), "edited script") {
		t.Error("Collection not updated with edited script")
	}
}

func TestScriptPath(t *testing.T) {
	service := NewService(Config{})

	// Test collection-level script
	result := service.scriptPath("my-api", []string{"collection"}, "prerequest")
	expected := "scripts/my-api/_collection__prerequest.js"
	if result != expected {
		t.Errorf("scriptPath() = %v, want %v", result, expected)
	}

	// Test request-level script
	result = service.scriptPath("user-tests", []string{"Get User"}, "test")
	expected = "scripts/user-tests/get-user__test.js"
	if result != expected {
		t.Errorf("scriptPath() = %v, want %v", result, expected)
	}
}

func TestSanitize(t *testing.T) {
	if sanitize("Test") != "test" {
		t.Error("sanitize should lowercase")
	}
	if sanitize("Get User Data") != "get-user-data" {
		t.Error("sanitize should replace spaces with hyphens")
	}
	if sanitize("") != "unnamed" {
		t.Error("sanitize should handle empty strings")
	}
}
