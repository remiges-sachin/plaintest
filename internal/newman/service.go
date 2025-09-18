package newman

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Service struct {
	executable string
	workingDir string
	debug      bool
}

type Options struct {
	Environment string
	CSV         string
	Debug       bool
	Reporters   []string
	OutputDir   string
	Iterations  int
}

type Result struct {
	Success    bool
	ExitCode   int
	Output     string
	ReportPath string
}

func NewService() *Service {
	return &Service{
		executable: "newman",
		workingDir: ".",
	}
}

func (s *Service) SetDebug(debug bool) {
	s.debug = debug
}

func (s *Service) Run(collection string, options Options) (*Result, error) {
	if collection == "" {
		return nil, errors.New("collection path is required")
	}

	args := []string{"run", collection}

	// Add environment file if specified
	if options.Environment != "" {
		args = append(args, "-e", options.Environment)
	}

	// Add CSV data file if specified
	if options.CSV != "" {
		args = append(args, "-d", options.CSV)
	}

	// Add debug flags if enabled
	if options.Debug {
		args = append(args, "--verbose", "--reporter-cli-show-timestamps")
	}

	if s.debug {
		fmt.Printf("[debug] newman %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command(s.executable, args...)
	cmd.Dir = s.workingDir
	output, err := cmd.CombinedOutput()

	result := &Result{
		Success:  err == nil,
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   string(output),
	}

	return result, err
}

// RunWithFlags runs Newman with custom flags passed through
func (s *Service) RunWithFlags(collection string, flags []string) (*Result, error) {
	if collection == "" {
		return nil, errors.New("collection path is required")
	}

	args := []string{"run", collection}
	args = append(args, flags...)

	if s.debug {
		fmt.Printf("[debug] newman %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command(s.executable, args...)
	cmd.Dir = s.workingDir
	output, err := cmd.CombinedOutput()

	result := &Result{
		Success:  err == nil,
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   string(output),
	}

	return result, err
}

// RunWithEnvironmentExport runs Newman and exports the environment to a file
func (s *Service) RunWithEnvironmentExport(collection string, flags []string, exportEnvPath string) (*Result, error) {
	if collection == "" {
		return nil, errors.New("collection path is required")
	}

	args := []string{"run", collection}
	args = append(args, flags...)

	// Add environment export flag
	args = append(args, "--export-environment", exportEnvPath)

	if s.debug {
		fmt.Printf("[debug] newman %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command(s.executable, args...)
	cmd.Dir = s.workingDir
	output, err := cmd.CombinedOutput()

	result := &Result{
		Success:  err == nil,
		ExitCode: cmd.ProcessState.ExitCode(),
		Output:   string(output),
	}

	return result, err
}

func (s *Service) IsInstalled() bool {
	_, err := exec.LookPath(s.executable)
	return err == nil
}
