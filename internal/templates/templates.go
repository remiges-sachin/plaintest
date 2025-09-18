package templates

import (
	"embed"
	"os"
)

//go:embed collections/* environments/* data/*
var templateFS embed.FS

type ProjectInitializer struct{}

func NewProjectInitializer() *ProjectInitializer {
	return &ProjectInitializer{}
}

func (p *ProjectInitializer) CreateProjectStructure() error {
	// Create directories
	directories := []string{
		"collections",
		"scripts",
		"data",
		"environments",
		"reports",
	}
	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Create template files
	if err := p.createCollectionTemplates(); err != nil {
		return err
	}

	if err := p.createEnvironmentTemplates(); err != nil {
		return err
	}

	if err := p.createDataTemplates(); err != nil {
		return err
	}

	return nil
}

func (p *ProjectInitializer) createCollectionTemplates() error {
	// Copy auth collection template
	authData, err := templateFS.ReadFile("collections/auth.postman_collection.json")
	if err != nil {
		return err
	}
	err = os.WriteFile("collections/get_auth.postman_collection.json", authData, 0644)
	if err != nil {
		return err
	}

	// Copy API tests collection template
	apiData, err := templateFS.ReadFile("collections/api_tests.postman_collection.json")
	if err != nil {
		return err
	}
	err = os.WriteFile("collections/api_tests.postman_collection.json", apiData, 0644)
	if err != nil {
		return err
	}

	// Copy smoke collection template
	smokeData, err := templateFS.ReadFile("collections/smoke.postman_collection.json")
	if err != nil {
		return err
	}
	return os.WriteFile("collections/smoke.postman_collection.json", smokeData, 0644)
}

func (p *ProjectInitializer) createEnvironmentTemplates() error {
	envData, err := templateFS.ReadFile("environments/dummyjson.postman_environment.json")
	if err != nil {
		return err
	}
	return os.WriteFile("environments/dummyjson.postman_environment.json", envData, 0644)
}

func (p *ProjectInitializer) createDataTemplates() error {
	csvData, err := templateFS.ReadFile("data/example.csv")
	if err != nil {
		return err
	}
	return os.WriteFile("data/example.csv", csvData, 0644)
}
