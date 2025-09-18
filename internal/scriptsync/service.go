package scriptsync

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config drives the extract and build process.
type Config struct {
	CollectionsDir string
	ScriptsDir     string
}

// Service coordinates extract and build operations between collections and scripts.
type Service struct {
	cfg Config
}

// NewService constructs a new Service with sane defaults.
func NewService(cfg Config) *Service {
	if cfg.CollectionsDir == "" {
		cfg.CollectionsDir = "collections"
	}
	if cfg.ScriptsDir == "" {
		cfg.ScriptsDir = "scripts"
	}
	return &Service{cfg: cfg}
}

// Extract extracts scripts from a collection to editable JS files.
func (s *Service) Extract(collectionName string) error {
	if err := s.ensureDirs(); err != nil {
		return err
	}

	collectionPath := filepath.Join(s.cfg.CollectionsDir, collectionName+".postman_collection.json")
	if _, err := os.Stat(collectionPath); err != nil {
		return fmt.Errorf("collection not found: %s", collectionPath)
	}

	return s.extractCollection(collectionPath)
}

// Build builds a collection with updated scripts from JS files.
func (s *Service) Build(collectionName string) error {
	if err := s.ensureDirs(); err != nil {
		return err
	}

	collectionPath := filepath.Join(s.cfg.CollectionsDir, collectionName+".postman_collection.json")
	if _, err := os.Stat(collectionPath); err != nil {
		return fmt.Errorf("collection not found: %s", collectionPath)
	}

	return s.buildCollection(collectionPath)
}

func (s *Service) ensureDirs() error {
	dirs := []string{s.cfg.CollectionsDir, s.cfg.ScriptsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) extractCollection(collectionPath string) error {
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return err
	}

	var node any
	if err := json.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("%s: %w", collectionPath, err)
	}

	collMap, ok := node.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: unexpected JSON structure", collectionPath)
	}

	collectionName := collectionID(collMap, filepath.Base(collectionPath))
	fmt.Printf("Extracting scripts from %s:\n", filepath.Base(collectionPath))

	if events, ok := collMap["event"].([]any); ok {
		if err := s.extractEvents(collectionName, []string{"collection"}, events); err != nil {
			return err
		}
	}

	if items, ok := collMap["item"].([]any); ok {
		if err := s.extractItems(collectionName, nil, items); err != nil {
			return err
		}
	}

	fmt.Printf("Extraction complete\n")
	return nil
}

func (s *Service) buildCollection(collectionPath string) error {
	data, err := os.ReadFile(collectionPath)
	if err != nil {
		return err
	}

	var node any
	if err := json.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("%s: %w", collectionPath, err)
	}

	collMap, ok := node.(map[string]any)
	if !ok {
		return fmt.Errorf("%s: unexpected JSON structure", collectionPath)
	}

	collectionName := collectionID(collMap, filepath.Base(collectionPath))
	fmt.Printf("Building %s from scripts:\n", filepath.Base(collectionPath))

	if events, ok := collMap["event"].([]any); ok {
		if err := s.buildEvents(collectionName, []string{"collection"}, events); err != nil {
			return err
		}
	}

	if items, ok := collMap["item"].([]any); ok {
		if err := s.buildItems(collectionName, nil, items); err != nil {
			return err
		}
	}

	// Write updated collection back to original location
	buildData, err := json.MarshalIndent(collMap, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(collectionPath, buildData, 0o644); err != nil {
		return err
	}

	fmt.Printf("Collection updated successfully\n")
	return nil
}

func (s *Service) extractItems(collectionName string, parents []string, items []any) error {
	for _, itemVal := range items {
		item, ok := itemVal.(map[string]any)
		if !ok {
			return errors.New("unexpected item structure")
		}

		name := stringValue(item["name"])
		currentPath := append(parents, name)

		if events, ok := item["event"].([]any); ok {
			if err := s.extractEvents(collectionName, currentPath, events); err != nil {
				return err
			}
		}

		if children, ok := item["item"].([]any); ok {
			if err := s.extractItems(collectionName, currentPath, children); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) buildItems(collectionName string, parents []string, items []any) error {
	for _, itemVal := range items {
		item, ok := itemVal.(map[string]any)
		if !ok {
			return errors.New("unexpected item structure")
		}

		name := stringValue(item["name"])
		currentPath := append(parents, name)

		if events, ok := item["event"].([]any); ok {
			if err := s.buildEvents(collectionName, currentPath, events); err != nil {
				return err
			}
		}

		if children, ok := item["item"].([]any); ok {
			if err := s.buildItems(collectionName, currentPath, children); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) extractEvents(collectionName string, parents []string, events []any) error {
	for _, eventVal := range events {
		evt, ok := eventVal.(map[string]any)
		if !ok {
			return errors.New("unexpected event structure")
		}
		listen := stringValue(evt["listen"])
		scriptMap, ok := evt["script"].(map[string]any)
		if !ok {
			continue
		}

		rawLines := extractExec(scriptMap["exec"])
		rawContent := joinLines(rawLines)
		scriptPath := s.scriptPath(collectionName, parents, listen)

		// Always overwrite script files
		if err := os.MkdirAll(filepath.Dir(scriptPath), 0o755); err != nil {
			return err
		}

		normContent := normalizeScript(rawContent)
		if err := os.WriteFile(scriptPath, withTrailingNewline(normContent), 0o644); err != nil {
			return err
		}

		fmt.Printf("✓ Extracted: %s\n", filepath.Base(scriptPath))
	}
	return nil
}

func (s *Service) buildEvents(collectionName string, parents []string, events []any) error {
	for _, eventVal := range events {
		evt, ok := eventVal.(map[string]any)
		if !ok {
			return errors.New("unexpected event structure")
		}
		listen := stringValue(evt["listen"])
		scriptMap, ok := evt["script"].(map[string]any)
		if !ok {
			continue
		}

		scriptPath := s.scriptPath(collectionName, parents, listen)

		// Read script content from file
		scriptBytes, err := os.ReadFile(scriptPath)
		if err != nil {
			return fmt.Errorf("script file not found: %s", scriptPath)
		}

		scriptContent := normalizeScript(string(scriptBytes))
		scriptMap["exec"] = splitLines(scriptContent)

		fmt.Printf("✓ Injected: %s\n", filepath.Base(scriptPath))
	}
	return nil
}

func (s *Service) scriptPath(collectionName string, parents []string, listen string) string {
	parts := []string{s.cfg.ScriptsDir, sanitize(collectionName)}
	if len(parents) == 0 {
		parents = []string{"root"}
	}
	if len(parents) > 1 {
		for _, part := range parents[:len(parents)-1] {
			parts = append(parts, sanitize(part))
		}
	}
	last := parents[len(parents)-1]
	dir := filepath.Join(parts...)

	// Collection-level scripts need "_collection" prefix to avoid naming conflicts.
	// If someone names a request "collection", both files would collide.
	// The underscore prevents this and sorts these files first.
	if last == "collection" {
		last = "_collection"
	}

	fileName := sanitize(last) + "__" + sanitize(listen) + ".js"
	return filepath.Join(dir, fileName)
}

func collectionID(coll map[string]any, fallback string) string {
	if info, ok := coll["info"].(map[string]any); ok {
		if name := stringValue(info["name"]); name != "" {
			return name
		}
	}
	return fallback
}

func extractExec(execVal any) []string {
	switch val := execVal.(type) {
	case []any:
		result := make([]string, 0, len(val))
		for _, line := range val {
			result = append(result, stringValue(line))
		}
		return result
	case []string:
		return val
	case string:
		return []string{val}
	default:
		return nil
	}
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		cleaned = append(cleaned, strings.ReplaceAll(line, "\r\n", "\n"))
	}
	return strings.Join(cleaned, "\n")
}

func splitLines(content string) []string {
	if content == "" {
		return []string{""}
	}
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	normalized = strings.TrimSuffix(normalized, "\n")
	return strings.Split(normalized, "\n")
}

func normalizeScript(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	return strings.TrimRight(content, "\n")
}

func withTrailingNewline(content string) []byte {
	if content == "" {
		return []byte("\n")
	}
	return []byte(content + "\n")
}

func stringValue(in any) string {
	switch v := in.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}

func sanitize(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	cleaned := strings.Trim(b.String(), "-")
	if cleaned == "" {
		return "unnamed"
	}
	return cleaned
}
