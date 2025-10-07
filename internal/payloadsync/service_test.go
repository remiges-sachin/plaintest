package payloadsync

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
	if service.cfg.PayloadsDir != "payloads" {
		t.Errorf("PayloadsDir = %v, want payloads", service.cfg.PayloadsDir)
	}
}

func createTestCollectionWithBody() map[string]any {
	return map[string]any{
		"info": map[string]any{"name": "Test"},
		"item": []any{
			map[string]any{
				"name": "Create User",
				"request": map[string]any{
					"method": "POST",
					"url":    "{{base_url}}/users",
					"body": map[string]any{
						"mode": "raw",
						"raw":  "{\n  \"name\": \"{{test_name}}\",\n  \"email\": \"{{test_email}}\",\n  \"age\": 25\n}",
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
	collection := createTestCollectionWithBody()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})
	err := service.Extract("test")

	if err != nil {
		t.Errorf("Extract failed: %v", err)
	}

	// Verify payload file was created
	expectedFile := "payloads/test/create-user.json"
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s not created", expectedFile)
	}

	// Verify payload content is valid JSON
	payloadData, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read payload file: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		t.Errorf("Payload file contains invalid JSON: %v", err)
	}

	// Verify Postman variables are preserved
	if !strings.Contains(string(payloadData), "{{test_name}}") {
		t.Error("Postman variables not preserved in payload")
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

func TestService_Extract_NoBody(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection with request but no body
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := map[string]any{
		"info": map[string]any{"name": "Test"},
		"item": []any{
			map[string]any{
				"name": "Get User",
				"request": map[string]any{
					"method": "GET",
					"url":    "{{base_url}}/users",
				},
			},
		},
	}
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})
	err := service.Extract("test")

	if err != nil {
		t.Errorf("Extract should not fail for requests without body: %v", err)
	}

	// Verify no payload file was created
	expectedFile := "payloads/test/get-user.json"
	if _, err := os.Stat(expectedFile); !os.IsNotExist(err) {
		t.Error("Payload file should not be created for requests without body")
	}
}

func TestService_Build(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection and extract payloads first
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := createTestCollectionWithBody()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})

	// Extract payloads first
	if err := service.Extract("test"); err != nil {
		t.Fatalf("Failed to extract: %v", err)
	}

	// Update payload file
	updatedPayload := map[string]any{
		"name":  "{{test_name}}",
		"email": "{{test_email}}",
		"age":   30, // Changed from 25 to 30
		"role":  "admin",
	}
	payloadData, _ := json.MarshalIndent(updatedPayload, "", "  ")
	if err := os.WriteFile("payloads/test/create-user.json", payloadData, 0644); err != nil {
		t.Fatalf("Failed to write payload: %v", err)
	}

	// Build with updated payload
	err := service.Build("test")
	if err != nil {
		t.Errorf("Build failed: %v", err)
	}

	// Verify collection was updated with payload content
	updatedData, err := os.ReadFile("collections/test.postman_collection.json")
	if err != nil {
		t.Fatalf("Failed to read updated collection: %v", err)
	}
	updatedStr := string(updatedData)
	if !strings.Contains(updatedStr, "role") {
		t.Errorf("Collection not updated with new payload field. Content: %s", updatedStr)
	}
	if !strings.Contains(updatedStr, "30") {
		t.Errorf("Collection not updated with changed payload value. Content: %s", updatedStr)
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
	collection := createTestCollectionWithBody()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})

	// Extract payloads
	if err := service.Extract("test"); err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Edit extracted payload
	updatedPayload := map[string]any{
		"name":     "{{test_name}}",
		"email":    "{{test_email}}",
		"age":      35,
		"verified": true,
	}
	payloadData, _ := json.MarshalIndent(updatedPayload, "", "  ")
	if err := os.WriteFile("payloads/test/create-user.json", payloadData, 0644); err != nil {
		t.Fatalf("Failed to edit payload: %v", err)
	}

	// Build collection with edited payload
	if err := service.Build("test"); err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify collection contains edited payload
	updatedData, err := os.ReadFile("collections/test.postman_collection.json")
	if err != nil {
		t.Fatalf("Failed to read updated collection: %v", err)
	}
	updatedStr := string(updatedData)
	if !strings.Contains(updatedStr, "verified") {
		t.Errorf("Collection not updated with edited payload. Content: %s", updatedStr)
	}
	if !strings.Contains(updatedStr, "{{test_name}}") {
		t.Error("Postman variables lost during round-trip")
	}
}

func TestPayloadPath(t *testing.T) {
	service := NewService(Config{})

	result := service.payloadPath("my-api", []string{"Create User"})
	expected := "payloads/my-api/create-user.json"
	if result != expected {
		t.Errorf("payloadPath() = %v, want %v", result, expected)
	}

	result = service.payloadPath("user-tests", []string{"Save KYC Draft"})
	expected = "payloads/user-tests/save-kyc-draft.json"
	if result != expected {
		t.Errorf("payloadPath() = %v, want %v", result, expected)
	}
}

func TestService_Extract_MultiplePaths(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection with nested folders
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := map[string]any{
		"info": map[string]any{"name": "Test"},
		"item": []any{
			map[string]any{
				"name": "Users",
				"item": []any{
					map[string]any{
						"name": "Create User",
						"request": map[string]any{
							"method": "POST",
							"body": map[string]any{
								"mode": "raw",
								"raw":  "{\"name\": \"test\"}",
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})
	err := service.Extract("test")

	if err != nil {
		t.Errorf("Extract failed: %v", err)
	}

	// Verify payload file created with proper naming
	expectedFile := "payloads/test/users-create-user.json"
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("Expected file %s not created", expectedFile)
	}
}

func TestService_Build_PreservesEmptyBody(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection with GET request (no body)
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := map[string]any{
		"info": map[string]any{"name": "Test"},
		"item": []any{
			map[string]any{
				"name": "Get User",
				"request": map[string]any{
					"method": "GET",
					"url":    "{{base_url}}/users",
				},
			},
		},
	}
	originalData, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", originalData, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})

	// Build should not fail even without payload files
	err := service.Build("test")
	if err != nil {
		t.Errorf("Build should not fail for requests without payload files: %v", err)
	}

	// Verify collection remains unchanged
	updatedData, err := os.ReadFile("collections/test.postman_collection.json")
	if err != nil {
		t.Fatalf("Failed to read collection: %v", err)
	}

	var original, updated map[string]any
	if err := json.Unmarshal(originalData, &original); err != nil {
		t.Fatalf("Failed to unmarshal original: %v", err)
	}
	if err := json.Unmarshal(updatedData, &updated); err != nil {
		t.Fatalf("Failed to unmarshal updated: %v", err)
	}

	if updated["item"] == nil {
		t.Error("Build corrupted collection structure")
	}
}

func TestService_Extract_InvalidJSON(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection with non-JSON body
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := map[string]any{
		"info": map[string]any{"name": "Test"},
		"item": []any{
			map[string]any{
				"name": "Create User",
				"request": map[string]any{
					"method": "POST",
					"body": map[string]any{
						"mode": "raw",
						"raw":  "not valid json",
					},
				},
			},
		},
	}
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	service := NewService(Config{})
	err := service.Extract("test")

	// Should not fail for non-JSON bodies, just skip them
	if err != nil {
		t.Errorf("Extract should not fail for non-JSON bodies: %v", err)
	}
}

func TestService_Build_InvalidPayloadJSON(t *testing.T) {
	_, cleanup := setupTest(t)
	defer cleanup()

	// Create collection
	if err := os.MkdirAll("collections", 0755); err != nil {
		t.Fatalf("Failed to create collections dir: %v", err)
	}
	collection := createTestCollectionWithBody()
	data, _ := json.Marshal(collection)
	if err := os.WriteFile("collections/test.postman_collection.json", data, 0644); err != nil {
		t.Fatalf("Failed to write collection: %v", err)
	}

	// Create invalid JSON payload file
	if err := os.MkdirAll("payloads/test", 0755); err != nil {
		t.Fatalf("Failed to create payloads dir: %v", err)
	}
	if err := os.WriteFile("payloads/test/create-user.json", []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write payload: %v", err)
	}

	service := NewService(Config{})
	err := service.Build("test")

	if err == nil {
		t.Error("Expected error for invalid JSON in payload file")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("Expected 'invalid JSON' error, got: %v", err)
	}
}

func TestSanitize_EmptyString(t *testing.T) {
	result := sanitize("")
	expected := "unnamed"
	if result != expected {
		t.Errorf("sanitize(\"\") = %v, want %v", result, expected)
	}
}
