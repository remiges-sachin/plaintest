package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/ssd532/plaintest/internal/core"
	"github.com/ssd532/plaintest/internal/csv"
	"github.com/ssd532/plaintest/internal/newman"
	"github.com/ssd532/plaintest/internal/scriptsync"
	"github.com/ssd532/plaintest/internal/templates"
)

var rootCmd = &cobra.Command{
	Use:   "plaintest",
	Short: "PlainTest CLI for API testing",
	Long:  "PlainTest provides a framework for API testing using Postman collections and CSV data.",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the tool version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(core.Version)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create PlainTest project structure",
	Long:  "Creates the basic PlainTest project structure with collections/, data/, environments/, and reports/ directories plus working template files.",
	Run: func(cmd *cobra.Command, args []string) {
		initializer := templates.NewProjectInitializer()
		if err := initializer.CreateProjectStructure(); err != nil {
			fmt.Printf("Error creating project structure: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("PlainTest project initialized successfully!")
		fmt.Println("Try: plaintest run smoke")
		fmt.Println("Or:  plaintest run get_auth api_tests")
		fmt.Println("Or:  plaintest run api_tests -d data/example.csv")
	},
}

var rowSelection string
var debugNewman bool
var generateReports bool
var collectionsToRunOnce []string
var generatedReports []string

// Flag constants for better self-documentation
const (
	envShortFlag  = "-e"
	envLongFlag   = "--environment"
	dataShortFlag = "-d"
	dataLongFlag  = "--iteration-data"
	rowsShortFlag = "-r"
	rowsLongFlag  = "--rows"
	debugFlag     = "--debug"
	reportsFlag   = "--reports"

	// Newman flag constants
	reportersFlag    = "--reporters"
	jsonReporter     = "json"
	defaultReporters = "cli,htmlextra,json"
	jsonExportFlag   = "--reporter-json-export"
	htmlExportFlag   = "--reporter-htmlextra-export"

	// File constants
	reportsDir      = "reports"
	timestampFormat = "20060102T150405"
)

// buildRunCommandLong creates the long description with available collections
func buildRunCommandLong() string {
	config := discoverAllFiles()

	availableNames := make([]string, 0, len(config.Collections))
	for name := range config.Collections {
		availableNames = append(availableNames, name)
	}

	envNames := make([]string, 0, len(config.Environments))
	for name := range config.Environments {
		envNames = append(envNames, name)
	}

	dataNames := make([]string, 0, len(config.DataFiles))
	for name := range config.DataFiles {
		dataNames = append(dataNames, name)
	}

	return fmt.Sprintf(`Execute API tests using Newman as a proxy. Specify collection names and any Newman flags.

Available collections: %v
Available environments: %v
Available data files: %v

Examples:
  plaintest run smoke
  plaintest run get_auth api_tests -d data/example.csv
  plaintest run get_auth api_tests -d data/example.csv --once get_auth
  plaintest run api_tests --verbose --reports
  plaintest run smoke -r 2-5 --debug

All Newman flags are supported. PlainTest-specific flags are listed below.`, availableNames, envNames, dataNames)
}

var runCmd = &cobra.Command{
	Use:   "run [collections...] [newman-flags...]",
	Short: "Execute API tests with Newman",
	Long:  buildRunCommandLong(),
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Clear any previously tracked reports
		generatedReports = nil

		service := newman.NewService()
		service.SetDebug(debugNewman)

		if !service.IsInstalled() {
			fmt.Println("Error: Newman is not installed. Install with: npm install -g newman newman-reporter-htmlextra")
			os.Exit(1)
		}

		// Discover available collections, environments, and data files
		config := discoverAllFiles()

		// Parse collections and Newman flags from raw args (skip "plaintest run")
		rawArgs := os.Args[2:] // Skip "plaintest run"
		collections, newmanFlags, err := parseArguments(rawArgs, config)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(collections) == 0 {
			availableCollections := make([]string, 0, len(config.Collections))
			for name := range config.Collections {
				availableCollections = append(availableCollections, name)
			}
			fmt.Printf("Error: No collections specified. Available: %v\n", availableCollections)
			os.Exit(1)
		}

		// Add default environment if not specified and only one environment exists
		if !hasEnvironmentFlag(newmanFlags) {
			if len(config.Environments) == 1 {
				for _, envPath := range config.Environments {
					newmanFlags = append(newmanFlags, "-e", envPath)
					break
				}
			}
		}

		// Run each collection in sequence with environment sharing
		var tempEnvFile string
		var exitCode int // Track exit code to show reports before exiting
		defer func() {
			if tempEnvFile != "" {
				cleanupTempFile(tempEnvFile)
			}
		}()

		for i, collectionName := range collections {
			collectionPath, exists := config.Collections[collectionName]
			if !exists {
				availableCollections := make([]string, 0, len(config.Collections))
				for name := range config.Collections {
					availableCollections = append(availableCollections, name)
				}
				fmt.Printf("Unknown collection: %s. Available: %v\n", collectionName, availableCollections)
				os.Exit(1)
			}

			currentFlags := determineCollectionFlags(collectionName, newmanFlags, collectionsToRunOnce)
			printCollectionStatus(collectionName, currentFlags)

			if rowSelection != "" {
				currentFlags = applyRowSelection(currentFlags, rowSelection)
			}

			// For subsequent collections, use shared environment from previous collection
			if i > 0 && tempEnvFile != "" {
				currentFlags = replaceEnvironmentInFlags(currentFlags, tempEnvFile)
			}

			// Add report flags if requested
			if generateReports {
				currentFlags = addReportFlags(currentFlags, collectionName)
			}

			var result *newman.Result
			var err error

			// If this is not the last collection, export environment for next collection
			if i < len(collections)-1 {
				// Create temporary environment file for sharing
				if tempEnvFile == "" {
					tempEnvFile, err = createTempEnvironmentFile()
					if err != nil {
						fmt.Printf("Error creating temporary environment file: %v\n", err)
						os.Exit(1)
					}
				}
				result, err = service.RunWithEnvironmentExport(collectionPath, currentFlags, tempEnvFile)
			} else {
				// Last collection doesn't need to export environment
				result, err = service.RunWithFlags(collectionPath, currentFlags)
			}

			if err != nil {
				fmt.Printf("Test execution failed: %v\n", err)
				if result != nil && result.Output != "" {
					fmt.Println("Newman output:")
					fmt.Println(result.Output)
				}
				exitCode = 1
				break // Exit loop but continue to show reports
			}

			if result.Success {
				if isVerboseMode(currentFlags) && result.Output != "" {
					fmt.Println(result.Output)
				} else {
					fmt.Printf("%s collection: All tests passed!\n", collectionName)
				}
			} else {
				fmt.Printf("%s collection: Tests completed with exit code: %d\n", collectionName, result.ExitCode)
				if result.Output != "" {
					fmt.Println("Newman output:")
					fmt.Println(result.Output)
				}
				exitCode = 1
				break // Exit loop but continue to show reports
			}
		}

		// Show summary of generated reports if capture was used
		if len(generatedReports) > 0 {
			fmt.Println()
			fmt.Println("Generated Reports:")
			for _, report := range generatedReports {
				if strings.HasSuffix(report, ".json") {
					fmt.Printf("   JSON: %s\n", report)
				} else if strings.HasSuffix(report, ".html") {
					fmt.Printf("   HTML: %s\n", report)
				}
			}
		}

		// Exit with error code if any tests failed
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	},
}

var scriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Manage collection scripts",
	Long:  "Pull scripts from collections to editable files or push edited scripts back to collections.",
}

var scriptsPullCmd = &cobra.Command{
	Use:   "pull [collection-name]",
	Short: "Pull scripts from collection to editable JS files",
	Long:  "Pulls all scripts from a Postman collection to individual JavaScript files for editing outside Postman.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collectionName := args[0]
		service := scriptsync.NewService(scriptsync.Config{})
		if err := service.Extract(collectionName); err != nil {
			fmt.Printf("Error pulling scripts: %v\n", err)
			os.Exit(1)
		}
	},
}

var scriptsPushCmd = &cobra.Command{
	Use:   "push [collection-name]",
	Short: "Push updated scripts from JS files to collection",
	Long:  "Pushes scripts from edited JavaScript files back to the Postman collection. Scripts are the source of truth.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		collectionName := args[0]
		service := scriptsync.NewService(scriptsync.Config{})
		if err := service.Build(collectionName); err != nil {
			fmt.Printf("Error pushing scripts: %v\n", err)
			os.Exit(1)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List project resources",
	Long:  "List available collections, data files, scripts, or environments in the current PlainTest project.",
}

var listCollectionsCmd = &cobra.Command{
	Use:   "collections",
	Short: "List available collections",
	Long:  "Lists all Postman collections found in the collections directory.",
	Run: func(cmd *cobra.Command, args []string) {
		config := discoverAllFiles()
		if len(config.Collections) == 0 {
			fmt.Println("No collections found in collections/ directory")
			return
		}

		fmt.Println("Available collections:")
		for name, path := range config.Collections {
			fmt.Printf("  %s (%s)\n", name, path)
		}
	},
}

var listDataCmd = &cobra.Command{
	Use:   "data",
	Short: "List available data files",
	Long:  "Lists all CSV data files found in the data directory.",
	Run: func(cmd *cobra.Command, args []string) {
		config := discoverAllFiles()
		if len(config.DataFiles) == 0 {
			fmt.Println("No data files found in data/ directory")
			return
		}

		fmt.Println("Available data files:")
		for name, path := range config.DataFiles {
			fmt.Printf("  %s (%s)\n", name, path)
		}
	},
}

var listEnvironmentsCmd = &cobra.Command{
	Use:   "environments",
	Short: "List available environments",
	Long:  "Lists all Postman environment files found in the environments directory.",
	Run: func(cmd *cobra.Command, args []string) {
		config := discoverAllFiles()
		if len(config.Environments) == 0 {
			fmt.Println("No environments found in environments/ directory")
			return
		}

		fmt.Println("Available environments:")
		for name, path := range config.Environments {
			fmt.Printf("  %s (%s)\n", name, path)
		}
	},
}

var listScriptsCmd = &cobra.Command{
	Use:   "scripts",
	Short: "List extracted scripts",
	Long:  "Lists all extracted script directories found in the scripts directory.",
	Run: func(cmd *cobra.Command, args []string) {
		scriptsDir := "scripts"

		entries, err := os.ReadDir(scriptsDir)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No scripts directory found")
				return
			}
			fmt.Printf("Error reading scripts directory: %v\n", err)
			return
		}

		scriptDirs := make([]string, 0)
		for _, entry := range entries {
			if entry.IsDir() {
				scriptDirs = append(scriptDirs, entry.Name())
			}
		}

		if len(scriptDirs) == 0 {
			fmt.Println("No extracted scripts found in scripts/ directory")
			fmt.Println("Use 'plaintest scripts pull [collection]' to pull scripts")
			return
		}

		fmt.Println("Available script directories:")
		for _, dir := range scriptDirs {
			scriptPath := filepath.Join(scriptsDir, dir)
			fileCount := countScriptFiles(scriptPath)
			fmt.Printf("  %s (%d script files)\n", dir, fileCount)
		}
	},
}

// countScriptFiles counts the number of .js files in a directory recursively
func countScriptFiles(dir string) int {
	count := 0
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors and continue
		}
		if !info.IsDir() && strings.HasSuffix(path, ".js") {
			count++
		}
		return nil
	})
	return count
}

// DiscoveryConfig holds all discovered files
type DiscoveryConfig struct {
	Collections  map[string]string
	Environments map[string]string
	DataFiles    map[string]string
}

// discoverFiles scans directories for files matching patterns and suffix
func discoverFiles(patterns []string, suffix string) map[string]string {
	fileMap := make(map[string]string)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Warning: Could not scan directory for %s: %v\n", pattern, err)
			continue
		}
		if len(matches) == 0 {
			continue
		}

		for _, filePath := range matches {
			filename := filepath.Base(filePath)
			name := strings.TrimSuffix(filename, suffix)
			fileMap[name] = filePath
		}

		// Stop after first pattern with matches (build/ has priority over raw collections/)
		if len(fileMap) > 0 {
			break
		}
	}

	return fileMap
}

// discoverAllFiles discovers all collections, environments, and data files
func discoverAllFiles() DiscoveryConfig {
	return DiscoveryConfig{
		Collections: discoverFiles([]string{
			filepath.Join("collections", "build", "*.postman_collection.json"),
			filepath.Join("collections", "*.postman_collection.json"),
		}, ".postman_collection.json"),
		Environments: discoverFiles([]string{
			filepath.Join("environments", "*.postman_environment.json"),
		}, ".postman_environment.json"),
		DataFiles: discoverFiles([]string{
			filepath.Join("data", "*.csv"),
		}, ".csv"),
	}
}

// resolveFilePathFromName attempts to resolve a name to a file path using the lookup map
func resolveFilePathFromName(name string, lookupMap map[string]string) string {
	if resolvedPath, exists := lookupMap[name]; exists {
		return resolvedPath
	}
	// Use as-is (might already be a full path)
	return name
}

// parseArguments separates collection names from Newman flags
func parseArguments(args []string, config DiscoveryConfig) (collections []string, newmanFlags []string, err error) {
	// Map of flags that need path resolution
	flagMaps := map[string]map[string]string{
		envShortFlag:  config.Environments,
		envLongFlag:   config.Environments,
		dataShortFlag: config.DataFiles,
		dataLongFlag:  config.DataFiles,
	}

	argIndex := 0
	for argIndex < len(args) {
		arg := args[argIndex]

		// Skip PlainTest-specific flags
		if arg == rowsShortFlag || arg == rowsLongFlag {
			argIndex += 2 // Skip flag and its value
			continue
		}
		if arg == "--once" {
			argIndex += 2
			continue
		}
		if arg == debugFlag {
			argIndex++
			continue
		}
		if arg == reportsFlag {
			argIndex++
			continue
		}

		if _, exists := config.Collections[arg]; exists {
			collections = append(collections, arg)
			argIndex++
		} else {
			// Check if this is a Newman flag (starts with -)
			if strings.HasPrefix(arg, "-") {
				newmanFlags = append(newmanFlags, arg)
				argIndex++

				// Check if this flag needs path resolution
				if lookupMap, needsResolution := flagMaps[arg]; needsResolution {
					if argIndex < len(args) && !strings.HasPrefix(args[argIndex], "-") {
						value := args[argIndex]
						resolvedPath := resolveFilePathFromName(value, lookupMap)
						newmanFlags = append(newmanFlags, resolvedPath)
						argIndex++
					}
				} else {
					// Check if this flag expects a value (next arg doesn't start with -)
					if argIndex < len(args) && !strings.HasPrefix(args[argIndex], "-") {
						// Add the flag value
						newmanFlags = append(newmanFlags, args[argIndex])
						argIndex++
					}
				}
			} else {
				// Non-flag argument that's not a known collection - treat as invalid collection
				availableNames := make([]string, 0, len(config.Collections))
				for name := range config.Collections {
					availableNames = append(availableNames, name)
				}
				return collections, newmanFlags, fmt.Errorf("unknown collection: %s. Available: %v", arg, availableNames)
			}
		}
	}
	return collections, newmanFlags, nil
}

// extractCSVFromFlags finds CSV file specified in Newman flags
func extractCSVFromFlags(flags []string) string {
	for i, flag := range flags {
		if flag == "-d" || flag == "--iteration-data" {
			if i+1 < len(flags) {
				return flags[i+1]
			}
		}
	}
	return ""
}

// replaceCSVInFlags replaces CSV file in Newman flags with new file
func replaceCSVInFlags(flags []string, newCSVFile string) []string {
	result := make([]string, len(flags))
	copy(result, flags)

	for i, flag := range result {
		if flag == "-d" || flag == "--iteration-data" {
			if i+1 < len(result) {
				result[i+1] = newCSVFile
			}
		}
	}
	return result
}

// hasEnvironmentFlag checks if environment is already specified in flags
func hasEnvironmentFlag(flags []string) bool {
	for _, flag := range flags {
		if flag == "-e" || flag == "--environment" {
			return true
		}
	}
	return false
}

func determineCollectionFlags(collectionName string, baseFlags []string, onceOnlyCollections []string) []string {
	if isRunOnceCollection(collectionName, onceOnlyCollections) {
		return removeCSVFlags(baseFlags)
	}
	return baseFlags
}

func printCollectionStatus(collectionName string, flags []string) {
	if csvFile := extractCSVFromFlags(flags); csvFile != "" {
		fmt.Printf("Running %s collection with CSV data iterations...\n", collectionName)
	} else if containsCollection(collectionsToRunOnce, collectionName) {
		fmt.Printf("Running %s collection once (no CSV iteration)...\n", collectionName)
	} else {
		fmt.Printf("Running %s collection...\n", collectionName)
	}
}

func applyRowSelection(flags []string, rowSelection string) []string {
	csvFile := extractCSVFromFlags(flags)
	if csvFile == "" {
		fmt.Println("Warning: Row selection specified but no CSV file found in flags")
		return flags
	}

	processor := csv.NewProcessor()
	tempCSVFile, err := processor.ProcessRows(csvFile, rowSelection)
	if err != nil {
		fmt.Printf("Error processing CSV rows: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Using row selection: %s from %s\n", rowSelection, csvFile)
	return replaceCSVInFlags(flags, tempCSVFile)
}

func isRunOnceCollection(collectionName string, onceOnlyCollections []string) bool {
	for _, name := range onceOnlyCollections {
		if name == collectionName {
			return true
		}
	}
	return false
}

func containsCollection(collections []string, target string) bool {
	for _, name := range collections {
		if name == target {
			return true
		}
	}
	return false
}

func removeCSVFlags(flags []string) []string {
	result := []string{}
	for i := 0; i < len(flags); i++ {
		if flags[i] == "-d" || flags[i] == "--iteration-data" {
			if i+1 < len(flags) && !strings.HasPrefix(flags[i+1], "-") {
				i++
			}
		} else {
			result = append(result, flags[i])
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(scriptsCmd)
	rootCmd.AddCommand(listCmd)

	scriptsCmd.AddCommand(scriptsPullCmd)
	scriptsCmd.AddCommand(scriptsPushCmd)

	listCmd.AddCommand(listCollectionsCmd)
	listCmd.AddCommand(listDataCmd)
	listCmd.AddCommand(listEnvironmentsCmd)
	listCmd.AddCommand(listScriptsCmd)

	// Only PlainTest-specific flags
	runCmd.Flags().StringVarP(&rowSelection, "rows", "r", "", "CSV row selection (2 | 2-5 | 1,3,5)")
	runCmd.Flags().BoolVar(&debugNewman, "debug", false, "Print the Newman command before running")
	runCmd.Flags().BoolVar(&generateReports, "reports", false, "Generate timestamped HTML and JSON report files")
	runCmd.Flags().StringSliceVar(&collectionsToRunOnce, "once", []string{}, "Collections to run once without CSV iteration (comma-separated or repeat flag)")

	// Allow unknown flags to be passed to Newman
	runCmd.FParseErrWhitelist.UnknownFlags = true
}

// createTempEnvironmentFile creates a temporary environment file for collection chaining
func createTempEnvironmentFile() (string, error) {
	tmpDir := os.TempDir()
	timestamp := time.Now().UnixNano()
	tempFile := filepath.Join(tmpDir, fmt.Sprintf("plaintest_env_%d.json", timestamp))
	return tempFile, nil
}

// replaceEnvironmentInFlags replaces environment file in Newman flags
func replaceEnvironmentInFlags(flags []string, newEnvFile string) []string {
	result := make([]string, 0, len(flags))

	for i := 0; i < len(flags); i++ {
		flag := flags[i]
		if flag == "-e" || flag == "--environment" {
			// Replace environment flag and its value
			result = append(result, flag)
			if i+1 < len(flags) {
				result = append(result, newEnvFile)
				i++ // Skip the old environment file
			}
		} else {
			result = append(result, flag)
		}
	}

	return result
}

// cleanupTempFile removes temporary environment file
func cleanupTempFile(filePath string) {
	if filePath != "" {
		os.Remove(filePath)
	}
}

func addReportFlags(flags []string, collectionName string) []string {
	flags = ensureJSONReporter(flags)
	flags = addExportPaths(flags, collectionName)
	return flags
}

func ensureJSONReporter(flags []string) []string {
	for i, flag := range flags {
		if flag == reportersFlag {
			if needsJSONReporter(flags, i) {
				flags[i+1] = flags[i+1] + "," + jsonReporter
			}
			return flags
		}
	}
	return append(flags, reportersFlag, defaultReporters)
}

func needsJSONReporter(flags []string, reportersIndex int) bool {
	reportersValueIndex := reportersIndex + 1
	return reportersValueIndex < len(flags) && !strings.Contains(flags[reportersValueIndex], jsonReporter)
}

func addExportPaths(flags []string, collectionName string) []string {
	timestamp := time.Now().Format(timestampFormat)
	jsonFile := filepath.Join(reportsDir, fmt.Sprintf("%s_%s.json", collectionName, timestamp))
	htmlFile := filepath.Join(reportsDir, fmt.Sprintf("%s_%s.html", collectionName, timestamp))

	flags = append(flags, jsonExportFlag, jsonFile)
	flags = append(flags, htmlExportFlag, htmlFile)

	// Track generated reports for summary at the end
	generatedReports = append(generatedReports, jsonFile, htmlFile)

	return flags
}

func isVerboseMode(flags []string) bool {
	for _, flag := range flags {
		if flag == "--verbose" {
			return true
		}
	}
	return false
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
