package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory for test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Test that init command creates all PlainTest directories
	os.Args = []string{"plaintest", "init"}

	// This should create all PlainTest directories and template files
	main()

	// Verify all directories exist
	directories := []string{
		"collections",
		"scripts",
		"data",
		"environments",
		"reports",
	}
	for _, dir := range directories {
		dirPath := filepath.Join(tempDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Fatalf("init command should create %s directory", dir)
		}
	}

	// Verify template files exist
	expectedFiles := []string{
		"collections/get_auth.postman_collection.json",
		"collections/api_tests.postman_collection.json",
		"collections/smoke.postman_collection.json",
		"environments/dummyjson.postman_environment.json",
		"data/example.csv",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(tempDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Fatalf("init command should create template file: %s", file)
		}
	}

}

func TestDiscoverCollections(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory for test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Create collections directory hierarchy and test files
	err = os.MkdirAll("collections/raw", 0755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.MkdirAll("collections/build", 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test collection files
	rawFiles := []string{
		"collections/raw/test1.postman_collection.json",
		"collections/raw/test2.postman_collection.json",
	}
	buildFiles := []string{
		"collections/build/test1.postman_collection.json",
		"collections/build/test2.postman_collection.json",
		"collections/build/get_auth.postman_collection.json",
		"collections/build/api_tests.postman_collection.json",
	}

	for _, file := range append(rawFiles, buildFiles...) {
		f, err := os.Create(file)
		if err != nil {
			t.Fatal(err)
		}
		f.Close()
	}

	// Test discovery
	config := discoverAllFiles()

	// Should discover from build directory first
	expected := map[string]string{
		"test1":     "collections/build/test1.postman_collection.json",
		"test2":     "collections/build/test2.postman_collection.json",
		"get_auth":  "collections/build/get_auth.postman_collection.json",
		"api_tests": "collections/build/api_tests.postman_collection.json",
	}

	if !reflect.DeepEqual(config.Collections, expected) {
		t.Errorf("discoverCollections() = %v, want %v", config.Collections, expected)
	}
}

func TestParseArguments(t *testing.T) {
	config := DiscoveryConfig{
		Collections: map[string]string{
			"smoke":     "collections/smoke.postman_collection.json",
			"get_auth":  "collections/get_auth.postman_collection.json",
			"api_tests": "collections/api_tests.postman_collection.json",
		},
		Environments: map[string]string{},
		DataFiles:    map[string]string{},
	}

	tests := []struct {
		name            string
		args            []string
		wantCollections []string
		wantNewmanFlags []string
	}{
		{
			name:            "single collection",
			args:            []string{"smoke"},
			wantCollections: []string{"smoke"},
			wantNewmanFlags: nil,
		},
		{
			name:            "multiple collections",
			args:            []string{"get_auth", "api_tests"},
			wantCollections: []string{"get_auth", "api_tests"},
			wantNewmanFlags: nil,
		},
		{
			name:            "collection with Newman flags",
			args:            []string{"smoke", "--verbose", "-d", "data.csv"},
			wantCollections: []string{"smoke"},
			wantNewmanFlags: []string{"--verbose", "-d", "data.csv"},
		},
		{
			name:            "collection with row selection",
			args:            []string{"api_tests", "-d", "data.csv", "-r", "2-5", "--verbose"},
			wantCollections: []string{"api_tests"},
			wantNewmanFlags: []string{"-d", "data.csv", "--verbose"},
		},
		{
			name:            "Newman flags only",
			args:            []string{"--help"},
			wantCollections: nil,
			wantNewmanFlags: []string{"--help"},
		},
		{
			name:            "collection with --once flag",
			args:            []string{"get_auth", "api_tests", "-d", "data.csv", "--once", "get_auth"},
			wantCollections: []string{"get_auth", "api_tests"},
			wantNewmanFlags: []string{"-d", "data.csv"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCollections, gotNewmanFlags, err := parseArguments(tt.args, config)

			if err != nil {
				t.Errorf("parseArguments() error = %v", err)
				return
			}

			if !reflect.DeepEqual(gotCollections, tt.wantCollections) {
				t.Errorf("parseArguments() collections = %v, want %v", gotCollections, tt.wantCollections)
			}

			if !reflect.DeepEqual(gotNewmanFlags, tt.wantNewmanFlags) {
				t.Errorf("parseArguments() newmanFlags = %v, want %v", gotNewmanFlags, tt.wantNewmanFlags)
			}
		})
	}
}

func TestExtractCSVFromFlags(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  string
	}{
		{
			name:  "short flag",
			flags: []string{"-d", "data.csv", "--verbose"},
			want:  "data.csv",
		},
		{
			name:  "long flag",
			flags: []string{"--iteration-data", "test.csv", "--bail"},
			want:  "test.csv",
		},
		{
			name:  "no CSV flag",
			flags: []string{"--verbose", "--bail"},
			want:  "",
		},
		{
			name:  "flag without value",
			flags: []string{"-d"},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCSVFromFlags(tt.flags)
			if got != tt.want {
				t.Errorf("extractCSVFromFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplaceCSVInFlags(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		newCSVFile string
		want       []string
	}{
		{
			name:       "replace short flag",
			flags:      []string{"-d", "old.csv", "--verbose"},
			newCSVFile: "new.csv",
			want:       []string{"-d", "new.csv", "--verbose"},
		},
		{
			name:       "replace long flag",
			flags:      []string{"--iteration-data", "old.csv", "--bail"},
			newCSVFile: "new.csv",
			want:       []string{"--iteration-data", "new.csv", "--bail"},
		},
		{
			name:       "no CSV flag",
			flags:      []string{"--verbose", "--bail"},
			newCSVFile: "new.csv",
			want:       []string{"--verbose", "--bail"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceCSVInFlags(tt.flags, tt.newCSVFile)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceCSVInFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasEnvironmentFlag(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  bool
	}{
		{
			name:  "has short flag",
			flags: []string{"-e", "localhost", "--verbose"},
			want:  true,
		},
		{
			name:  "has long flag",
			flags: []string{"--environment", "uat", "--bail"},
			want:  true,
		},
		{
			name:  "no environment flag",
			flags: []string{"--verbose", "--bail"},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasEnvironmentFlag(tt.flags)
			if got != tt.want {
				t.Errorf("hasEnvironmentFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateTempEnvironmentFile(t *testing.T) {
	tempFile, err := createTempEnvironmentFile()
	if err != nil {
		t.Errorf("createTempEnvironmentFile() error = %v", err)
		return
	}

	if tempFile == "" {
		t.Error("createTempEnvironmentFile() should return non-empty path")
	}

	if !strings.Contains(tempFile, "plaintest_env_") {
		t.Errorf("createTempEnvironmentFile() should contain plaintest_env_ in path, got %v", tempFile)
	}
}

func TestReplaceEnvironmentInFlags(t *testing.T) {
	tests := []struct {
		name       string
		flags      []string
		newEnvFile string
		want       []string
	}{
		{
			name:       "replace short flag",
			flags:      []string{"-e", "old.json", "--verbose"},
			newEnvFile: "new.json",
			want:       []string{"-e", "new.json", "--verbose"},
		},
		{
			name:       "replace long flag",
			flags:      []string{"--environment", "old.json", "--bail"},
			newEnvFile: "new.json",
			want:       []string{"--environment", "new.json", "--bail"},
		},
		{
			name:       "no environment flag",
			flags:      []string{"--verbose", "--bail"},
			newEnvFile: "new.json",
			want:       []string{"--verbose", "--bail"},
		},
		{
			name:       "flag without value",
			flags:      []string{"-e"},
			newEnvFile: "new.json",
			want:       []string{"-e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceEnvironmentInFlags(tt.flags, tt.newEnvFile)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replaceEnvironmentInFlags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddReportFlags(t *testing.T) {
	tests := []struct {
		name           string
		flags          []string
		collectionName string
		wantJSONExport bool
		wantHTMLExport bool
	}{
		{
			name:           "adds capture flags",
			flags:          []string{"--verbose"},
			collectionName: "smoke",
			wantJSONExport: true,
			wantHTMLExport: true,
		},
		{
			name:           "preserves existing flags",
			flags:          []string{"--reporters", "cli,htmlextra", "--bail"},
			collectionName: "api_tests",
			wantJSONExport: true,
			wantHTMLExport: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addReportFlags(tt.flags, tt.collectionName)

			result := strings.Join(got, " ")
			if tt.wantJSONExport && !strings.Contains(result, "--reporter-json-export") {
				t.Errorf("addReportFlags() should add JSON export flag")
			}
			if tt.wantHTMLExport && !strings.Contains(result, "--reporter-htmlextra-export") {
				t.Errorf("addReportFlags() should add HTML export flag")
			}
		})
	}
}

func TestEnsureJSONReporter(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  []string
	}{
		{
			name:  "adds reporters when missing",
			flags: []string{"--verbose"},
			want:  []string{"--verbose", "--reporters", "cli,htmlextra,json"},
		},
		{
			name:  "adds json to existing reporters",
			flags: []string{"--reporters", "cli,htmlextra"},
			want:  []string{"--reporters", "cli,htmlextra,json"},
		},
		{
			name:  "preserves existing json reporter",
			flags: []string{"--reporters", "cli,json"},
			want:  []string{"--reporters", "cli,json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ensureJSONReporter(tt.flags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ensureJSONReporter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsVerboseMode(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		want  bool
	}{
		{
			name:  "detects verbose flag",
			flags: []string{"--verbose", "--bail"},
			want:  true,
		},
		{
			name:  "no verbose flag",
			flags: []string{"--bail", "--timeout", "5000"},
			want:  false,
		},
		{
			name:  "empty flags",
			flags: []string{},
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isVerboseMode(tt.flags)
			if got != tt.want {
				t.Errorf("isVerboseMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironmentSharingIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory for test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize project structure
	os.Args = []string{"plaintest", "init"}
	main()

	// Test environment sharing flow with get_auth and api_tests
	config := discoverAllFiles()
	collections := []string{"get_auth", "api_tests"}
	flags := []string{"--debug"}

	// Parse arguments to simulate collection chaining
	parsedCollections, _, err := parseArguments(append(collections, flags...), config)
	if err != nil {
		t.Fatalf("parseArguments() error = %v", err)
	}

	// Verify multiple collections were parsed
	if len(parsedCollections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(parsedCollections))
	}

	if parsedCollections[0] != "get_auth" || parsedCollections[1] != "api_tests" {
		t.Errorf("Expected [get_auth, api_tests], got %v", parsedCollections)
	}

	// Test temporary environment file creation
	tempEnvFile, err := createTempEnvironmentFile()
	if err != nil {
		t.Fatalf("createTempEnvironmentFile() error = %v", err)
	}

	// Verify temp file path format
	if !strings.Contains(tempEnvFile, "plaintest_env_") {
		t.Errorf("Temp environment file should contain 'plaintest_env_', got %s", tempEnvFile)
	}

	// Test environment flag replacement
	originalFlags := []string{"-e", "environments/dummyjson.postman_environment.json", "--debug"}
	replacedFlags := replaceEnvironmentInFlags(originalFlags, tempEnvFile)

	expectedFlags := []string{"-e", tempEnvFile, "--debug"}
	if !reflect.DeepEqual(replacedFlags, expectedFlags) {
		t.Errorf("replaceEnvironmentInFlags() = %v, want %v", replacedFlags, expectedFlags)
	}

	// Verify the environment sharing logic components work together
	// This tests the integration without actually running Newman
	if len(parsedCollections) > 1 {
		// Simulate flags with environment file (as would happen with auto-detection)
		flagsWithEnv := []string{"-e", "environments/dummyjson.postman_environment.json", "--debug"}

		// Simulate second collection with shared environment
		secondFlags := replaceEnvironmentInFlags(flagsWithEnv, tempEnvFile)

		// Verify second collection would use the temp environment file
		found := false
		for i, flag := range secondFlags {
			if flag == "-e" && i+1 < len(secondFlags) {
				if secondFlags[i+1] == tempEnvFile {
					found = true
					break
				}
			}
		}

		if !found {
			t.Error("Second collection should use temporary environment file for sharing")
		}
	}
}

func TestListCommands(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory for test
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	// Initialize project to create test data
	os.Args = []string{"plaintest", "init"}
	main()

	// Test list collections
	config := discoverAllFiles()
	if len(config.Collections) == 0 {
		t.Error("Expected collections to be found after init")
	}

	// Test list data files
	if len(config.DataFiles) == 0 {
		t.Error("Expected data files to be found after init")
	}

	// Test list environments
	if len(config.Environments) == 0 {
		t.Error("Expected environments to be found after init")
	}

	// Create some script directories to test scripts listing
	scriptsDir := "scripts"
	testScriptDir := filepath.Join(scriptsDir, "test-collection")
	err = os.MkdirAll(testScriptDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test script file
	testScript := filepath.Join(testScriptDir, "test-script.js")
	err = os.WriteFile(testScript, []byte("console.log('test');"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Test countScriptFiles function
	count := countScriptFiles(testScriptDir)
	if count != 1 {
		t.Errorf("Expected 1 script file, got %d", count)
	}
}

func TestCountScriptFiles(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested directory structure with script files
	dirs := []string{
		"subdir1",
		"subdir1/nested",
		"subdir2",
	}

	for _, dir := range dirs {
		fullDir := filepath.Join(tempDir, dir)
		err := os.MkdirAll(fullDir, 0755)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Create script files
	scriptFiles := []string{
		"script1.js",
		"subdir1/script2.js",
		"subdir1/nested/script3.js",
		"subdir2/script4.js",
		"notascript.txt", // This should not be counted
	}

	for _, file := range scriptFiles {
		fullPath := filepath.Join(tempDir, file)
		err := os.WriteFile(fullPath, []byte("console.log('test');"), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Test counting
	count := countScriptFiles(tempDir)
	expectedCount := 4 // Only .js files should be counted
	if count != expectedCount {
		t.Errorf("Expected %d script files, got %d", expectedCount, count)
	}

	// Test empty directory
	emptyDir := filepath.Join(tempDir, "empty")
	err := os.MkdirAll(emptyDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	count = countScriptFiles(emptyDir)
	if count != 0 {
		t.Errorf("Expected 0 script files in empty directory, got %d", count)
	}
}

func TestIsRunOnceCollection(t *testing.T) {
	onceOnlyList := []string{"auth", "setup", "get_password"}

	tests := []struct {
		name           string
		collectionName string
		expected       bool
	}{
		{
			name:           "collection in run-once list",
			collectionName: "auth",
			expected:       true,
		},
		{
			name:           "collection not in run-once list",
			collectionName: "api_tests",
			expected:       false,
		},
		{
			name:           "empty collection name",
			collectionName: "",
			expected:       false,
		},
		{
			name:           "partial match should not work",
			collectionName: "auth_extended",
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRunOnceCollection(tt.collectionName, onceOnlyList)
			if result != tt.expected {
				t.Errorf("isRunOnceCollection(%q, %v) = %v, want %v", tt.collectionName, onceOnlyList, result, tt.expected)
			}
		})
	}
}

func TestRemoveCSVFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		expected []string
	}{
		{
			name:     "strip short CSV flag",
			flags:    []string{"-d", "data.csv", "--verbose"},
			expected: []string{"--verbose"},
		},
		{
			name:     "strip long CSV flag",
			flags:    []string{"--iteration-data", "data.csv", "--bail"},
			expected: []string{"--bail"},
		},
		{
			name:     "no CSV flags",
			flags:    []string{"--verbose", "--bail"},
			expected: []string{"--verbose", "--bail"},
		},
		{
			name:     "CSV flag without value",
			flags:    []string{"-d", "--verbose"},
			expected: []string{"--verbose"},
		},
		{
			name:     "multiple CSV flags",
			flags:    []string{"-d", "data1.csv", "--iteration-data", "data2.csv", "--verbose"},
			expected: []string{"--verbose"},
		},
		{
			name:     "empty flags",
			flags:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeCSVFlags(tt.flags)
			if len(result) != len(tt.expected) {
				t.Errorf("removeCSVFlags(%v) = %v, want %v", tt.flags, result, tt.expected)
				return
			}
			for i, flag := range result {
				if flag != tt.expected[i] {
					t.Errorf("removeCSVFlags(%v) = %v, want %v", tt.flags, result, tt.expected)
					break
				}
			}
		})
	}
}
