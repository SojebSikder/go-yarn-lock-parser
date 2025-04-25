package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type YarnLockEntry struct {
	Version      string
	Resolved     string
	Integrity    string
	Dependencies map[string]string
}

func ParseYarnLock(filePath string) (map[string]YarnLockEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Map to store package details
	packageData := make(map[string]YarnLockEntry)

	// Regex patterns
	namePattern := regexp.MustCompile(`^"([^"]+)":`)
	versionPattern := regexp.MustCompile(`^  version "([^"]+)"`)
	resolvedPattern := regexp.MustCompile(`^  resolved "([^"]+)"`)
	integrityPattern := regexp.MustCompile(`^  integrity "([^"]+)"`)

	var currentPackage string
	var currentEntry YarnLockEntry
	inDependencies := false

	// Reading the yarn.lock file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Check for package name
		if match := namePattern.FindStringSubmatch(line); match != nil {
			if currentPackage != "" {
				packageData[currentPackage] = currentEntry
			}
			currentPackage = match[1]
			currentEntry = YarnLockEntry{Dependencies: make(map[string]string)}
			inDependencies = false
			continue
		}

		// Check for version
		if match := versionPattern.FindStringSubmatch(line); match != nil {
			currentEntry.Version = match[1]
			inDependencies = false
			continue
		}

		// Check for resolved URL
		if match := resolvedPattern.FindStringSubmatch(line); match != nil {
			currentEntry.Resolved = match[1]
			inDependencies = false
			continue
		}

		// Check for integrity hash
		if match := integrityPattern.FindStringSubmatch(line); match != nil {
			currentEntry.Integrity = match[1]
			inDependencies = false
			continue
		}

		// Check for dependencies block
		if strings.TrimSpace(line) == "dependencies:" {
			inDependencies = true
			continue
		}

		// Handle dependencies lines
		if inDependencies {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || !strings.Contains(trimmed, " ") || strings.HasPrefix(trimmed, "\"") {
				inDependencies = false
				continue
			}
			parts := strings.SplitN(trimmed, " ", 2)
			if len(parts) == 2 {
				key := strings.Trim(parts[0], `"`)
				value := strings.Trim(parts[1], `"`)
				currentEntry.Dependencies[key] = value
			}
		}
	}

	// Add the last package to the map
	if currentPackage != "" {
		packageData[currentPackage] = currentEntry
	}

	return packageData, nil
}

func main() {
	packageData, err := ParseYarnLock("yarn.lock")
	if err != nil {
		fmt.Println("Error parsing yarn.lock file:", err)
		return
	}

	// Print out the parsed data
	for pkg, entry := range packageData {
		fmt.Printf("Package: %s\n", pkg)
		fmt.Printf("  Version: %s\n", entry.Version)
		fmt.Printf("  Resolved: %s\n", entry.Resolved)
		fmt.Printf("  Integrity: %s\n", entry.Integrity)
		fmt.Println("  Dependencies:")
		for dep, version := range entry.Dependencies {
			fmt.Printf("    - %s: %s\n", dep, version)
		}
		fmt.Println()
	}
}
