package ini

import (
	"reflect"
	"slices"
	"strconv"

	"github.com/hashicorp/go-hclog"
)

// Section represents a single INI configuration section containing key-value pairs.
type Section struct {
	raw    map[string]string
	logger hclog.Logger
}

// NewSection creates a new Section instance.
//
// Parameters:
//   - raw: Map of key-value pairs for the section
//   - logger: Logger instance for error and warning reporting
//
// Returns:
//   - *Section: Initialized Section instance
func NewSection(raw map[string]string, logger hclog.Logger) *Section {
	return &Section{
		raw:    raw,
		logger: logger,
	}
}

// String retrieves a string value by key with optional validation.
//
// Parameters:
//   - key: Configuration key
//   - def: Default value returned if key doesn't exist or validation fails
//   - validateList: Optional. Allowed values. Nil disables validation
//
// Returns:
//   - string: Value if key exists and passes validation, otherwise def
func (b *Section) String(key string, def string, validateList []string) string {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	if validateList != nil {
		if !slices.Contains(validateList, val) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return val
}

// Strings retrieves a comma-separated string value as a slice.
//
// Parameters:
//   - key: Configuration key
//   - def: Default slice returned if key doesn't exist
//   - sep: Optional. Custom separator. Defaults to ","
//
// Returns:
//   - []string: Split and trimmed values, or def if key doesn't exist
func (b *Section) Strings(key string, def []string, sep ...string) []string {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	separator := ","
	if len(sep) > 0 && sep[0] != "" {
		separator = sep[0]
	}

	list := Split(val, separator)
	return list
}

// Bool retrieves a boolean value by key.
//
// Parameters:
//   - key: Configuration key
//   - def: Default value returned if key doesn't exist
//
// Returns:
//   - bool: true if value equals "true", false otherwise
func (b *Section) Bool(key string, def bool) bool {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	return val == "true"
}

// Int retrieves an integer value by key with optional validation.
//
// Parameters:
//   - key: Configuration key
//   - def: Default value returned if key doesn't exist or validation fails
//   - validator: Optional. Function to validate parsed value. Nil disables validation
//
// Returns:
//   - int: Parsed integer value, or def on parse/validation failure
func (b *Section) Int(key string, def int, validator func(int) bool) int {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	res, err := strconv.Atoi(val)
	if err != nil {
		b.logger.Warn("Config value not a number", key, val)
		return def
	}

	if validator != nil {
		if !validator(res) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return res
}

// Ints retrieves a comma-separated string of integers as a slice.
//
// Parameters:
//   - key: Configuration key
//   - def: Default slice returned if key doesn't exist or parse fails
//   - sep: Optional. Custom separator. Defaults to ","
//
// Returns:
//   - []int: Parsed integer slice, or def on parse failure
func (b *Section) Ints(key string, def []int, sep ...string) []int {
	sliceStr, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	separator := ","
	if len(sep) > 0 && sep[0] != "" {
		separator = sep[0]
	}

	slice := Split(sliceStr, separator)

	res := make([]int, 0)
	for _, numStr := range slice {
		num, err := strconv.Atoi(numStr)
		if err != nil {
			b.logger.Warn("Config value not a number", key, sliceStr)
			return def
		}

		res = append(res, num)
	}

	return res
}

// Float32 retrieves a 32-bit float value by key with optional validation.
//
// Parameters:
//   - key: Configuration key
//   - def: Default value returned if key doesn't exist or validation fails
//   - validator: Optional. Function to validate parsed value. Nil disables validation
//
// Returns:
//   - float32: Parsed float32 value, or def on parse/validation failure
func (b *Section) Float32(key string, def float32, validator func(float32) bool) float32 {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	res, err := strconv.ParseFloat(val, 32)
	if err != nil {
		b.logger.Warn("Config value not a number", key, val)
		return def
	}

	res32 := float32(res)

	if validator != nil {
		if !validator(res32) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return res32
}

// Float64 retrieves a 64-bit float value by key with optional validation.
//
// Parameters:
//   - key: Configuration key
//   - def: Default value returned if key doesn't exist or validation fails
//   - validator: Optional. Function to validate parsed value. Nil disables validation
//
// Returns:
//   - float64: Parsed float64 value, or def on parse/validation failure
func (b *Section) Float64(key string, def float64, validator func(float64) bool) float64 {
	val, exist := b.raw[key]
	if !exist {
		b.logger.Warn("Config key not found", "key", key)
		return def
	}

	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		b.logger.Warn("Config value not a number", key, val)
		return def
	}

	if validator != nil {
		if !validator(res) {
			b.logger.Warn("Config value not valid", key, val)
			return def
		}
	}

	return res
}

func (b *Section) _string(key string, def string, valid string) string {
	validList := Split(valid, ",")
	return b.String(key, def, validList)
}

func (b *Section) _strings(key string, def string, sep ...string) []string {
	defList := []string{}
	if def != "" {
		defList = Split(def, ",")
	}

	return b.Strings(key, defList, sep...)
}

func (b *Section) _int(key string, def string, max string, min string) int {
	defI, err := strconv.Atoi(def)
	if err != nil {
		defI = 0
	}

	minI, err := strconv.Atoi(min)
	if err != nil {
		minI = 0
	}

	maxI, err := strconv.Atoi(max)
	ismax := err == nil

	return b.Int(key, defI, func(i int) bool {
		if ismax && maxI < i {
			return false
		}

		return i >= minI
	})
}

func (b *Section) _ints(key string, def string, sep ...string) []int {
	defListStr := []string{}
	if def != "" {
		defListStr = Split(def, ",")
	}

	defList := make([]int, 0)
	for _, defStr := range defListStr {
		defI, err := strconv.Atoi(defStr)
		if err != nil {
			defI = 0
		}

		defList = append(defList, defI)
	}

	return b.Ints(key, defList, sep...)
}

func (b *Section) _float64(key string, def string, max string, min string) float64 {
	def64, err := strconv.ParseFloat(def, 64)
	if err != nil {
		def64 = 0
	}

	min64, err := strconv.ParseFloat(min, 64)
	if err != nil {
		min64 = 0
	}

	max64, err := strconv.ParseFloat(max, 64)
	ismax := err == nil

	return b.Float64(key, def64, func(f float64) bool {
		if ismax && max64 < f {
			return false
		}

		return f >= min64
	})
}

func (b *Section) _float32(key string, def string, max string, min string) float32 {
	def64, err := strconv.ParseFloat(def, 64)
	def32 := float32(def64)
	if err != nil {
		def64 = 0
	}

	min64, err := strconv.ParseFloat(min, 64)
	min32 := float32(min64)
	if err != nil {
		min64 = 0
	}

	max64, err := strconv.ParseFloat(max, 64)
	max32 := float32(max64)
	ismax := err == nil

	return b.Float32(key, def32, func(f float32) bool {
		if ismax && max32 < f {
			return false
		}

		return f >= min32
	})
}

// Unmarshal parses section key-value pairs into a struct using struct tags.
//
// Supported struct tags:
//   - key: Map struct field to different key name (defaults to field name)
//   - def: Default value when key doesn't exist
//   - valid: Comma-separated list of allowed values (string fields only)
//   - min: Minimum allowed value (numeric fields)
//   - max: Maximum allowed value (numeric fields)
//   - sep: Custom separator for slice fields (defaults to ",")
//
// Supported field types:
//   - string, bool, int, float32, float64
//   - []string, []int
//
// Parameters:
//   - s: Pointer to struct that will receive parsed values
//
// Returns:
//   - error: Nil on success, error if s is not a pointer to struct
func (b *Section) Unmarshal(s interface{}) error {
	return structForRange(s, func(field reflect.StructField, fieldValue interface{}) interface{} {
		key, ok := field.Tag.Lookup("key")
		if !ok {
			key = field.Name
		}

		def := field.Tag.Get("def")

		switch field.Type.Kind() {

		// STRING
		case reflect.String:
			valid, ok := field.Tag.Lookup("valid")
			if !ok {
				return b.String(key, def, nil)
			}
			return b._string(key, def, valid)

		// SLICE
		case reflect.Slice:
			elemKind := field.Type.Elem().Kind()
			sep := field.Tag.Get("sep")

			switch elemKind {
			case reflect.String:
				return b._strings(key, def, sep)
			case reflect.Int:
				return b._ints(key, def, sep)
			}

		// BOOL
		case reflect.Bool:
			return b.Bool(key, def == "true")

		// INT
		case reflect.Int:
			max := field.Tag.Get("max")
			min := field.Tag.Get("min")

			return b._int(key, def, max, min)

		// FLOAT32
		case reflect.Float32:
			max := field.Tag.Get("max")
			min := field.Tag.Get("min")

			return b._float32(key, def, max, min)

		// FLOAT64
		case reflect.Float64:
			max := field.Tag.Get("max")
			min := field.Tag.Get("min")

			return b._float64(key, def, max, min)
		}

		return nil
	}, b.logger)
}
