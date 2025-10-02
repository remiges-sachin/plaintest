package payloadsync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config drives the extract and build process.
type Config struct {
	CollectionsDir string
	PayloadsDir    string
}

// Service coordinates extract and build operations between collections and payloads.
type Service struct {
	cfg Config
}

// NewService constructs a new Service with sane defaults.
func NewService(cfg Config) *Service {
	if cfg.CollectionsDir == "" {
		cfg.CollectionsDir = "collections"
	}
	if cfg.PayloadsDir == "" {
		cfg.PayloadsDir = "payloads"
	}
	return &Service{cfg: cfg}
}

// Extract extracts request bodies from a collection to editable JSON files.
func (s *Service) Extract(collectionName string) error {
	if err := s.ensureDirs(); err != nil {
		return err
	}

	collectionPath := filepath.Join(s.cfg.CollectionsDir, collectionName+".postman_collection.json")
	if _, err := os.Stat(collectionPath); err != nil {
		return fmt.Errorf("collection not found: %s", collectionPath)
	}

	return s.extractCollection(collectionPath, collectionName)
}

// Build builds a collection with updated payloads from JSON files.
func (s *Service) Build(collectionName string) error {
	if err := s.ensureDirs(); err != nil {
		return err
	}

	collectionPath := filepath.Join(s.cfg.CollectionsDir, collectionName+".postman_collection.json")
	if _, err := os.Stat(collectionPath); err != nil {
		return fmt.Errorf("collection not found: %s", collectionPath)
	}

	return s.buildCollection(collectionPath, collectionName)
}

func (s *Service) ensureDirs() error {
	if err := os.MkdirAll(s.cfg.CollectionsDir, 0755); err != nil {
		return fmt.Errorf("failed to create collections dir: %w", err)
	}
	if err := os.MkdirAll(s.cfg.PayloadsDir, 0755); err != nil {
		return fmt.Errorf("failed to create payloads dir: %w", err)
	}
	return nil
}

func (s *Service) extractCollection(collectionPath, collectionName string) error {
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return fmt.Errorf("failed to read collection: %w", err)
	}

	var collection map[string]any
	if err := json.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse collection: %w", err)
	}

	items, ok := collection["item"].([]any)
	if !ok {
		return nil
	}

	return s.extractItems(items, collectionName, []string{})
}

func (s *Service) extractItems(items []any, collectionName string, path []string) error {
	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, _ := itemMap["name"].(string)
		currentPath := append(path, name)

		// Check if this is a folder with nested items
		if nestedItems, ok := itemMap["item"].([]any); ok {
			if err := s.extractItems(nestedItems, collectionName, currentPath); err != nil {
				return err
			}
			continue
		}

		// Extract request body if present
		if request, ok := itemMap["request"].(map[string]any); ok {
			if err := s.extractRequestBody(request, collectionName, currentPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) extractRequestBody(request map[string]any, collectionName string, path []string) error {
	body, ok := request["body"].(map[string]any)
	if !ok {
		return nil
	}

	rawBody, ok := body["raw"].(string)
	if !ok || rawBody == "" {
		return nil
	}

	// Try to parse as JSON
	var payloadData any
	if err := json.Unmarshal([]byte(rawBody), &payloadData); err != nil {
		// Not valid JSON, skip
		return nil
	}

	// Write payload to file
	payloadPath := s.payloadPath(collectionName, path)
	payloadDir := filepath.Dir(payloadPath)
	if err := os.MkdirAll(payloadDir, 0755); err != nil {
		return fmt.Errorf("failed to create payload dir: %w", err)
	}

	// Pretty print JSON
	prettyJSON, err := json.MarshalIndent(payloadData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := os.WriteFile(payloadPath, prettyJSON, 0644); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	return nil
}

func (s *Service) buildCollection(collectionPath, collectionName string) error {
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return fmt.Errorf("failed to read collection: %w", err)
	}

	var collection map[string]any
	if err := json.Unmarshal(data, &collection); err != nil {
		return fmt.Errorf("failed to parse collection: %w", err)
	}

	items, ok := collection["item"].([]any)
	if !ok {
		return nil
	}

	if err := s.buildItems(items, collectionName, []string{}); err != nil {
		return err
	}

	// Write updated collection
	updatedData, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	if err := os.WriteFile(collectionPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write collection: %w", err)
	}

	return nil
}

func (s *Service) buildItems(items []any, collectionName string, path []string) error {
	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		name, _ := itemMap["name"].(string)
		currentPath := append(path, name)

		// Check if this is a folder with nested items
		if nestedItems, ok := itemMap["item"].([]any); ok {
			if err := s.buildItems(nestedItems, collectionName, currentPath); err != nil {
				return err
			}
			continue
		}

		// Update request body if payload file exists
		if request, ok := itemMap["request"].(map[string]any); ok {
			if err := s.buildRequestBody(request, collectionName, currentPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) buildRequestBody(request map[string]any, collectionName string, path []string) error {
	payloadPath := s.payloadPath(collectionName, path)

	// Check if payload file exists
	if _, err := os.Stat(payloadPath); os.IsNotExist(err) {
		return nil
	}

	// Read payload file
	payloadData, err := os.ReadFile(payloadPath)
	if err != nil {
		return fmt.Errorf("failed to read payload: %w", err)
	}

	// Validate JSON
	var payload any
	if err := json.Unmarshal(payloadData, &payload); err != nil {
		return fmt.Errorf("invalid JSON in payload file %s: %w", payloadPath, err)
	}

	// Convert to compact JSON string (what Postman expects)
	compactJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Update request body
	body, ok := request["body"].(map[string]any)
	if !ok {
		body = map[string]any{
			"mode": "raw",
		}
		request["body"] = body
	}

	body["raw"] = string(compactJSON)

	return nil
}

func (s *Service) payloadPath(collectionName string, path []string) string {
	// Sanitize path components
	sanitized := make([]string, len(path))
	for i, p := range path {
		sanitized[i] = sanitize(p)
	}

	// Join path components with dash
	filename := strings.Join(sanitized, "-") + ".json"

	return filepath.Join(s.cfg.PayloadsDir, collectionName, filename)
}

func sanitize(s string) string {
	if s == "" {
		return "unnamed"
	}
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}
