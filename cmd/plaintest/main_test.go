package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test helper functions
func newLinkSpec(collection string, items ...string) LinkSpec {
	if items == nil {
		items = []string{}
	}
	return LinkSpec{Collection: collection, Items: items}
}

func withTempDir(t *testing.T, fn func(tempDir string)) {
	tempDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatal(err)
	}

	fn(tempDir)
}

func TestInitCommand(t *testing.T) {
	withTempDir(t, func(tempDir string) {
		// SETUP
		os.Args = []string{"plaintest", "init"}

		// WHEN
		main()

		// THEN - Verify all directories exist
		expectedDirs := []string{"collections", "scripts", "data", "environments", "reports"}
		for _, dir := range expectedDirs {
			_, err := os.Stat(dir)
			assert.NoError(t, err, "init command should create %s directory", dir)
		}

		// THEN - Verify template files exist
		expectedFiles := []string{
			"collections/get_auth.postman_collection.json",
			"collections/api_tests.postman_collection.json",
			"collections/smoke.postman_collection.json",
			"environments/dummyjson.postman_environment.json",
			"data/example.csv",
		}

		for _, file := range expectedFiles {
			_, err := os.Stat(file)
			assert.NoError(t, err, "init command should create template file: %s", file)
		}
	})
}

func TestDiscoverCollections(t *testing.T) {
	withTempDir(t, func(tempDir string) {
		// SETUP - Create collections directory hierarchy
		assert.NoError(t, os.MkdirAll("collections/raw", 0755))
		assert.NoError(t, os.MkdirAll("collections/build", 0755))

		// SETUP - Create test collection files
		allFiles := []string{
			"collections/raw/test1.postman_collection.json",
			"collections/raw/test2.postman_collection.json",
			"collections/build/test1.postman_collection.json",
			"collections/build/test2.postman_collection.json",
			"collections/build/get_auth.postman_collection.json",
			"collections/build/api_tests.postman_collection.json",
		}

		for _, file := range allFiles {
			f, err := os.Create(file)
			assert.NoError(t, err)
			f.Close()
		}

		// WHEN
		config := discoverAllFiles()

		// THEN - Should prioritize build directory over raw
		expected := map[string]string{
			"test1":     "collections/build/test1.postman_collection.json",
			"test2":     "collections/build/test2.postman_collection.json",
			"get_auth":  "collections/build/get_auth.postman_collection.json",
			"api_tests": "collections/build/api_tests.postman_collection.json",
		}
		assert.Equal(t, expected, config.Collections, "should discover collections from build directory first")
	})
}

func TestParseLinkSpec(t *testing.T) {
	// GIVEN
	tests := []struct {
		name     string
		linkSpec string
		want     LinkSpec
		wantErr  bool
	}{
		{"collection only", "smoke", newLinkSpec("smoke"), false},
		{"quoted collection only", "\"api tests\"", newLinkSpec("api tests"), false},
		{"collection with single item", "auth.Login", newLinkSpec("auth", "Login"), false},
		{"quoted collection with single item", "\"auth\".\"Login Request\"", newLinkSpec("auth", "Login Request"), false},
		{"collection with multiple items", "\"api tests\".\"Create User,Update User,Delete User\"", newLinkSpec("api tests", "Create User", "Update User", "Delete User"), false},
		{"collection with items having spaces", "api_tests.\"User Folder,Admin Folder\"", newLinkSpec("api_tests", "User Folder", "Admin Folder"), false},
		{"empty items after dot", "auth.", newLinkSpec("auth"), false},
		{"empty string", "", newLinkSpec(""), false},
		{"only dot", ".", newLinkSpec(""), false},
		{"special characters in collection", "api-tests_v2", newLinkSpec("api-tests_v2"), false},
		{"whitespace around collection", " smoke ", newLinkSpec(" smoke "), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			got, err := parseLinkSpec(tt.linkSpec)

			// THEN
			if tt.wantErr {
				assert.Error(t, err, "parseLinkSpec() should return error for input: %s", tt.linkSpec)
			} else {
				assert.NoError(t, err, "parseLinkSpec() should not return error for input: %s", tt.linkSpec)
				assert.Equal(t, tt.want, got, "parseLinkSpec() should parse correctly")
			}
		})
	}
}

func TestParsePhases(t *testing.T) {
	// GIVEN
	tests := []struct {
		name       string
		setupLinks []string
		testLinks  []string
		want       []ExecutionPhase
	}{
		{
			"only test phase",
			[]string{},
			[]string{"smoke", "api_tests"},
			[]ExecutionPhase{{Links: []LinkSpec{newLinkSpec("smoke"), newLinkSpec("api_tests")}, Phase: "test"}},
		},
		{
			"only setup phase",
			[]string{"auth.Login", "db.Init"},
			[]string{},
			[]ExecutionPhase{{Links: []LinkSpec{newLinkSpec("auth", "Login"), newLinkSpec("db", "Init")}, Phase: "setup"}},
		},
		{
			"both setup and test phases",
			[]string{"auth.Login"},
			[]string{"\"api tests\".\"Create User,Update User\""},
			[]ExecutionPhase{
				{Links: []LinkSpec{newLinkSpec("auth", "Login")}, Phase: "setup"},
				{Links: []LinkSpec{newLinkSpec("api tests", "Create User", "Update User")}, Phase: "test"},
			},
		},
		{
			"empty phases",
			[]string{},
			[]string{},
			[]ExecutionPhase{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			got, err := parsePhases(tt.setupLinks, tt.testLinks)

			// THEN
			assert.NoError(t, err, "parsePhases() should not return error")
			assert.Equal(t, tt.want, got, "parsePhases() should create correct execution phases")
		})
	}
}

func TestParseArguments(t *testing.T) {
	// SETUP
	config := DiscoveryConfig{
		Collections:  map[string]string{"smoke": "collections/smoke.postman_collection.json"},
		Environments: map[string]string{},
		DataFiles:    map[string]string{},
	}

	// GIVEN
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"Newman flags only", []string{"--verbose", "-d", "data.csv"}, []string{"--verbose", "-d", "data.csv"}},
		{"skip setup and test flags", []string{"--setup", "auth", "--test", "smoke", "--verbose"}, []string{"--verbose"}},
		{"skip row selection", []string{"-r", "2-5", "--verbose"}, []string{"--verbose"}},
		{"mixed flags", []string{"--setup", "auth.Login", "-d", "data.csv", "--test", "api_tests", "--verbose"}, []string{"-d", "data.csv", "--verbose"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// WHEN
			_, gotNewmanFlags, err := parseArguments(tt.args, config)

			// THEN
			assert.NoError(t, err, "parseArguments() should not return error")
			assert.Equal(t, tt.want, gotNewmanFlags, "parseArguments() should filter PlainTest flags correctly")
		})
	}
}

func TestFlagManipulation(t *testing.T) {
	t.Run("extractCSVFromFlags", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			want  string
		}{
			{"short flag", []string{"-d", "data.csv", "--verbose"}, "data.csv"},
			{"long flag", []string{"--iteration-data", "test.csv", "--bail"}, "test.csv"},
			{"no CSV flag", []string{"--verbose", "--bail"}, ""},
			{"flag without value", []string{"-d"}, ""},
			{"empty flags", []string{}, ""},
			{"CSV flag at end", []string{"--verbose", "-d", "data.csv"}, "data.csv"},
			{"special characters in filename", []string{"-d", "test-data_v2.csv"}, "test-data_v2.csv"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := extractCSVFromFlags(tt.flags)

				// THEN
				assert.Equal(t, tt.want, got, "should extract CSV filename correctly")
			})
		}
	})

	t.Run("replaceCSVInFlags", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			new   string
			want  []string
		}{
			{"replace short flag", []string{"-d", "old.csv", "--verbose"}, "new.csv", []string{"-d", "new.csv", "--verbose"}},
			{"replace long flag", []string{"--iteration-data", "old.csv", "--bail"}, "new.csv", []string{"--iteration-data", "new.csv", "--bail"}},
			{"no CSV flag", []string{"--verbose", "--bail"}, "new.csv", []string{"--verbose", "--bail"}},
			{"empty flags", []string{}, "new.csv", []string{}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := replaceCSVInFlags(tt.flags, tt.new)

				// THEN
				assert.Equal(t, tt.want, got, "should replace CSV flag value correctly")
			})
		}
	})

	t.Run("removeCsvFlags", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			want  []string
		}{
			{"remove short CSV flag", []string{"-d", "data.csv", "--verbose"}, []string{"--verbose"}},
			{"remove long CSV flag", []string{"--iteration-data", "data.csv", "--bail"}, []string{"--bail"}},
			{"no CSV flags", []string{"--verbose", "--bail"}, []string{"--verbose", "--bail"}},
			{"CSV flag without value", []string{"-d", "--verbose"}, []string{"--verbose"}},
			{"multiple CSV flags", []string{"-d", "data1.csv", "--iteration-data", "data2.csv", "--verbose"}, []string{"--verbose"}},
			{"empty flags", []string{}, []string{}},
			{"only CSV flags", []string{"-d", "data.csv"}, []string{}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := removeCsvFlags(tt.flags)

				// THEN
				assert.Equal(t, tt.want, got, "should remove CSV flags correctly")
			})
		}
	})

	t.Run("hasEnvironmentFlag", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			want  bool
		}{
			{"has short flag", []string{"-e", "localhost", "--verbose"}, true},
			{"has long flag", []string{"--environment", "uat", "--bail"}, true},
			{"no environment flag", []string{"--verbose", "--bail"}, false},
			{"empty flags", []string{}, false},
			{"env flag without value", []string{"-e"}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := hasEnvironmentFlag(tt.flags)

				// THEN
				assert.Equal(t, tt.want, got, "should detect environment flag correctly")
			})
		}
	})

	t.Run("isVerboseMode", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			want  bool
		}{
			{"detects verbose flag", []string{"--verbose", "--bail"}, true},
			{"no verbose flag", []string{"--bail", "--timeout", "5000"}, false},
			{"empty flags", []string{}, false},
			{"verbose with other flags", []string{"-d", "data.csv", "--verbose"}, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := isVerboseMode(tt.flags)

				// THEN
				assert.Equal(t, tt.want, got, "should detect verbose mode correctly")
			})
		}
	})
}

func TestEnvironmentManagement(t *testing.T) {
	t.Run("createTempEnvironmentFile", func(t *testing.T) {
		// WHEN
		tempFile, err := createTempEnvironmentFile()

		// THEN
		assert.NoError(t, err, "should create temp file without error")
		assert.NotEmpty(t, tempFile, "should return non-empty path")
		assert.Contains(t, tempFile, "plaintest_env_", "should contain plaintest_env_ in path")
	})

	t.Run("replaceEnvironmentInFlags", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			new   string
			want  []string
		}{
			{"replace short flag", []string{"-e", "old.json", "--verbose"}, "new.json", []string{"-e", "new.json", "--verbose"}},
			{"replace long flag", []string{"--environment", "old.json", "--bail"}, "new.json", []string{"--environment", "new.json", "--bail"}},
			{"no environment flag", []string{"--verbose", "--bail"}, "new.json", []string{"--verbose", "--bail"}},
			{"flag without value", []string{"-e"}, "new.json", []string{"-e"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := replaceEnvironmentInFlags(tt.flags, tt.new)

				// THEN
				assert.Equal(t, tt.want, got, "should replace environment flag correctly")
			})
		}
	})
}

func TestReportGeneration(t *testing.T) {
	t.Run("addReportFlags", func(t *testing.T) {
		// WHEN
		got := addReportFlags([]string{"--verbose"}, "smoke")

		// THEN
		result := strings.Join(got, " ")
		assert.Contains(t, result, "--reporter-json-export", "should add JSON export flag")
		assert.Contains(t, result, "--reporter-htmlextra-export", "should add HTML export flag")
	})

	t.Run("ensureJSONReporter", func(t *testing.T) {
		// GIVEN
		tests := []struct {
			name  string
			flags []string
			want  []string
		}{
			{"adds reporters when missing", []string{"--verbose"}, []string{"--verbose", "--reporters", "cli,htmlextra,json"}},
			{"adds json to existing reporters", []string{"--reporters", "cli,htmlextra"}, []string{"--reporters", "cli,htmlextra,json"}},
			{"preserves existing json reporter", []string{"--reporters", "cli,json"}, []string{"--reporters", "cli,json"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// WHEN
				got := ensureJSONReporter(tt.flags)

				// THEN
				assert.Equal(t, tt.want, got, "should handle JSON reporter correctly")
			})
		}
	})
}
