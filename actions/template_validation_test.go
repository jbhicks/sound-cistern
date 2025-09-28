package actions_test

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jbhicks/sound-cistern/templates"

	"github.com/stretchr/testify/require"
)

// Test_AdminTemplateStructure validates that all admin templates follow proper structure
func Test_AdminTemplateStructure(t *testing.T) {
	r := require.New(t)

	// Define required elements for admin templates
	requiredElements := []string{
		`<nav class="container-fluid dashboard-nav">`,
		`<strong>My Go SaaS</strong>`,
		`<main class="container">`,
	}

	// Track templates that should have navigation (exclude partials and special cases)
	adminTemplates := []string{}
	excludePatterns := []string{
		"_", // partials (start with underscore)
		"htmx.plush.html",
		"application.plush.html",
	}

	// Walk through templates directory to find admin templates
	templateFS := templates.FS()
	err := fs.WalkDir(templateFS, "admin", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".plush.html") {
			// Check if this template should be excluded
			shouldExclude := false
			for _, pattern := range excludePatterns {
				if strings.Contains(filepath.Base(path), pattern) {
					shouldExclude = true
					break
				}
			}

			if !shouldExclude {
				adminTemplates = append(adminTemplates, path)
			}
		}
		return nil
	})
	r.NoError(err)

	// Validate each admin template has required navigation elements
	for _, templatePath := range adminTemplates {
		t.Run(fmt.Sprintf("Template_%s", templatePath), func(t *testing.T) {
			r := require.New(t)

			// Read template content
			content, err := fs.ReadFile(templateFS, templatePath)
			r.NoError(err)

			templateContent := string(content)

			// Check for required elements
			for _, element := range requiredElements {
				r.Contains(templateContent, element,
					"Template %s is missing required element: %s", templatePath, element)
			}

			// Verify navigation comes before main content
			navIndex := strings.Index(templateContent, `<nav class="container-fluid dashboard-nav">`)
			mainIndex := strings.Index(templateContent, `<main class="container">`)

			r.True(navIndex >= 0, "Template %s missing navigation", templatePath)
			r.True(mainIndex >= 0, "Template %s missing main container", templatePath)
			r.True(navIndex < mainIndex, "Template %s: navigation should come before main content", templatePath)
		})
	}

	// Ensure we found some templates to test
	r.Greater(len(adminTemplates), 0, "No admin templates found to validate")

	t.Logf("Validated %d admin templates for proper structure", len(adminTemplates))
}
