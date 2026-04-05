// Package ini provides functionality for parsing and managing INI configuration files.
package ini

import (
	"reflect"

	"github.com/hashicorp/go-hclog"
)

// Manager handles INI configuration file operations including section retrieval
// and unmarshaling into Go structs.
type Manager struct {
	raw    map[string]map[string]string
	path   string
	logger hclog.Logger
}

// New creates a new Manager instance for the specified INI file.
//
// Parameters:
//   - file: Path to the INI configuration file
//   - logger: Logger instance for error and warning reporting
//   - globalKey: Optional. Default section name. Defaults to "General" if not provided or empty
//
// Returns:
//   - *Manager: Initialized Manager instance. If file parsing fails, creates empty configuration
func New(file string, logger hclog.Logger, globalKey ...string) *Manager {
	initialSection := "General"
	if len(globalKey) != 0 && globalKey[0] != "" {
		initialSection = globalKey[0]
	}

	raw, err := GetMap(file, initialSection)
	if err != nil {
		raw = make(map[string]map[string]string, 0)
		logger.Error("INI PARSING",
			"error", err,
		)
	}

	return &Manager{
		raw:    raw,
		path:   file,
		logger: logger,
	}
}

// GetSection retrieves a configuration section by name.
//
// Parameters:
//   - name: Section name to retrieve
//
// Returns:
//   - *Section: Section instance. If section doesn't exist, returns empty section and logs warning
func (cm *Manager) GetSection(name string) *Section {
	sraw, exist := cm.raw[name]
	if !exist {
		cm.logger.Warn("Config section not found",
			"section", name,
			"file", cm.path,
		)

		sraw = make(map[string]string, 0)
	}

	return NewSection(sraw, cm.logger)
}

// ParseSection retrieves and unmarshals a section into the provided struct.
//
// Parameters:
//   - v: Pointer to struct that will receive unmarshaled section data
//   - name: Section name to parse
//
// Returns:
//   - *Section: Section instance containing raw data
//   - error: Unmarshal error if parsing fails
func (cm *Manager) ParseSection(v interface{}, name string) (*Section, error) {
	section := cm.GetSection(name)
	err := section.Unmarshal(v)
	if err != nil {
		return nil, err
	}
	return section, nil
}

// Unmarshal parses the entire INI file into a struct.
//
// The target struct must use "section" tags to map struct fields to INI sections.
// Only struct fields with section tags are processed.
//
// Parameters:
//   - v: Pointer to struct that will receive all section data
//
// Returns:
//   - error: Nil on success, error if v is not a pointer to struct
//
// Example struct tag:
//
//	type Config struct {
//	    Database struct { ... } `section:"Database"`
//	}
func (cm *Manager) Unmarshal(v interface{}) error {
	return structForRange(v, func(field reflect.StructField, currentValue interface{}) interface{} {
		if field.Type.Kind() != reflect.Struct {
			return nil
		}

		sectionName, ok := field.Tag.Lookup("section")
		if !ok {
			return nil
		}

		newStructPtr := reflect.New(field.Type)
		newStructInterface := newStructPtr.Interface()

		_, err := cm.ParseSection(newStructInterface, sectionName)
		if err != nil {
			cm.logger.Error("Failed to parse section",
				"section", sectionName,
				"name", field.Name,
				"error", err,
			)

			return nil
		}

		return newStructPtr.Elem().Interface()
	}, cm.logger)
}
