package ini

import (
	"os"
	"strings"
)

// GetMap parses an INI file and returns a map of sections to key-value pairs.
//
// Parameters:
//   - path: File system path to INI file
//   - initialSection: Default section name for key-value pairs before first section declaration
//
// Returns:
//   - map[string]map[string]string: Section name -> (key -> value) mapping
//   - error: File read or parse error
//
// Behavior:
//   - Lines starting with '#' are treated as comments and ignored
//   - Inline comments after '#' are stripped
//   - Empty lines are ignored
//   - Section headers are enclosed in square brackets: [SectionName]
func GetMap(path string, initialSection string) (map[string]map[string]string, error) {
	lines, err := load(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string)
	currentSection := initialSection

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.Contains(line, "#") {
			line = strings.SplitN(line, "#", 2)[0]
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if result[currentSection] == nil {
			result[currentSection] = make(map[string]string)
		}

		result[currentSection][key] = value
	}

	return result, nil
}

func load(path string) ([]string, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(bytes), "\n")

	var output []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			output = append(output, line)
		}
	}

	return output, nil
}
