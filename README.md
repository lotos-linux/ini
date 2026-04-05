# INI Configuration Library

A powerful Go library for parsing, managing, and generating INI configuration files with struct-based mapping, validation, and automatic documentation generation.

## Features

- **Dual functionality**: Parse existing INI files AND generate INI files from Go structs
- **Type-safe configuration**: Map INI sections/keys to Go structs with reflection
- **Rich validation**: Built-in support for min/max ranges, allowed values, and custom validators
- **Multiple data types**: Strings, integers, floats, booleans, and slices
- **Automatic comment generation**: Generates documentation comments from struct field comments
- **Go generate integration**: One-command configuration file generation
- **Hierarchical sections**: Support for nested structs mapped to INI sections
- **Graceful defaults**: Built-in default values with fallback support
- **Comprehensive logging**: Integration with go-hclog for debugging and warnings

## Installation

```bash
go get github.com/lotos-linux/ini
```

For the configuration generator:

```bash
go install github.com/lotos-linux/ini/cmd/ini-generator@latest
```

## Quick Start

### Reading Configuration Files

```go
package main

import (
    "github.com/lotos-linux/ini"
    "github.com/hashicorp/go-hclog"
)

type ServerConfig struct {
    Host string `def:"localhost"`
    Port int    `def:"8080" min:"1024" max:"65535"`
}

type AppConfig struct {
    Server ServerConfig `section:"Server"`
}

func main() {
    logger := hclog.Default()
    
    // Load configuration
    mgr := ini.New("app.conf", logger)
    
    // Parse specific section
    var config AppConfig
    if err := mgr.Unmarshal(&config); err != nil {
        logger.Error("Failed to parse config", "error", err)
    }
    
    // Or access sections directly
    section := mgr.GetSection("Server")
    host := section.String("Host", "localhost", nil)
    port := section.Int("Port", 8080, nil)
}
```

### Generating Configuration Files

Add the generate directive to your config struct file:

```go
//go:generate go run github.com/lotos-linux/ini/cmd/ini-generator -i $GOFILE -o .

package conf

// ini:app.conf
type AppConfig struct {
    Server ServerConfig `section:"Server"`
    Database DatabaseConfig `section:"Database"`
}

type ServerConfig struct {
    // Server listen address
    Host string `def:"localhost"`
    // Server port number
    Port int `def:"8080" min:"1024" max:"65535"`
}
```

Run:

```bash
go generate ./...
```

## API Reference

### Manager API

The `Manager` handles INI file operations and section management.

| Method | Description |
|--------|-------------|
| `New(file string, logger hclog.Logger, globalKey ...string) *Manager` | Creates a new manager instance |
| `GetSection(name string) *Section` | Retrieves a configuration section |
| `ParseSection(v interface{}, name string) (*Section, error)` | Retrieves and unmarshals a section |
| `Unmarshal(v interface{}) error` | Unmarshals entire file into a struct |

### Section API

The `Section` provides type-safe value retrieval methods.

| Method | Description |
|--------|-------------|
| `String(key, def string, validateList []string) string` | Gets string value with optional validation |
| `Strings(key string, def []string, sep ...string) []string` | Gets comma-separated string as slice |
| `Bool(key string, def bool) bool` | Gets boolean value |
| `Int(key string, def int, validator func(int) bool) int` | Gets integer with optional validation |
| `Ints(key string, def []int, sep ...string) []int` | Gets comma-separated integers as slice |
| `Float32(key string, def float32, validator func(float32) bool) float32` | Gets float32 with optional validation |
| `Float64(key string, def float64, validator func(float64) bool) float64` | Gets float64 with optional validation |
| `Unmarshal(s interface{}) error` | Unmarshals section into a struct |

### Helper Functions

| Function | Description |
|----------|-------------|
| `GetMap(path string, initialSection string) (map[string]map[string]string, error)` | Parses INI file into a map |
| `Split(s string, sep string) []string` | Splits and trims string by separator |

## Struct Tags Reference

### Reading Configuration (Unmarshal)

When using `Unmarshal` to read INI files into structs, the following tags are supported:

| Tag | Applies To | Description | Example |
|-----|-----------|-------------|---------|
| `section` | Struct fields | Maps struct field to INI section name | `section:"Database"` |
| `key` | All types | Maps struct field to different key name | `key:"server_host"` |
| `def` | All types | Default value if key is missing | `def:"localhost"` |
| `valid` | Strings | Comma-separated allowed values | `valid:"light,dark,auto"` |
| `min` | Numeric types | Minimum allowed value | `min:"1"` |
| `max` | Numeric types | Maximum allowed value | `max:"100"` |
| `sep` | Slices | Custom separator for splitting strings | `sep:";"` |

### Writing Configuration (Generator)

When using the generator to create INI files from structs, the same tags are used with the following additional behavior:

| Tag | Behavior |
|-----|----------|
| `def` | Default value written to generated INI |
| `min`/`max` | Written as comments in generated INI |
| `valid` | Written as allowed values in comments |
| Field comments | Written as documentation comments in generated INI |

## Generator Usage

### Command Line

```bash
ini-generator -i <input.go> -o <output-dir> [flags]
```

#### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-i` | Input Go file path | Required |
| `-o` | Output directory | `.` (current) |

### Struct Markers

The generator processes structs with special comments:

```go
// ini:output_filename.conf
type MyConfig struct {
    // This struct will be written to output_filename.conf
}
```

### Example: Complete Configuration

```go
//go:generate go run github.com/lotos-linux/ini/cmd/ini-generator -i $GOFILE -o .

package config

// ini:application.conf
type ApplicationConfig struct {
    Server   ServerConfig   `section:"Server"`
    Database DatabaseConfig `section:"Database"`
    Cache    CacheConfig    `section:"Cache"`
}

type ServerConfig struct {
    // HTTP server listen address
    Host string `def:"0.0.0.0"`
    // HTTP server port (1024-65535)
    Port int `def:"8080" min:"1024" max:"65535"`
    // Enable TLS/HTTPS
    TLS bool `def:"false"`
}

type DatabaseConfig struct {
    // Database connection string
    URL string `def:"postgres://localhost:5432/app"`
    // Maximum connections in pool
    MaxConnections int `def:"10" min:"1" max:"50"`
    // Connection timeout in seconds
    Timeout int `def:"30" min:"5" max:"120"`
}

type CacheConfig struct {
    // Redis server addresses
    Addresses []string `sep:"," def:"localhost:6379"`
    // Cache password (optional)
    Password string `def:""`
    // Default TTL in seconds
    TTL int `def:"3600" min:"60"`
}
```

Generated output (`application.conf`):

```ini
[Server]
# HTTP server listen address
Host = 0.0.0.0
# HTTP server port (1024-65535) (min: 1024, max: 65535)
Port = 8080
# Enable TLS/HTTPS (true, false)
TLS = false

[Database]
# Database connection string
URL = postgres://localhost:5432/app
# Maximum connections in pool (min: 1, max: 50)
MaxConnections = 10
# Connection timeout in seconds (min: 5, max: 120)
Timeout = 30

[Cache]
# Redis server addresses
Addresses = localhost:6379
# Cache password (optional)
Password = 
# Default TTL in seconds (min: 60)
TTL = 3600
```

## Advanced Usage

### Custom Validators

```go
// Custom integer validation
section.Int("port", 8080, func(port int) bool {
    return port > 1024 && port < 65535
})

// Custom float validation
section.Float64("ratio", 0.5, func(ratio float64) bool {
    return ratio >= 0.0 && ratio <= 1.0
})
```

### Working with Multiple Configuration Files

```go
type Config struct {
    App  AppConfig  `section:"App"`
    Log  LogConfig  `section:"Log"`
    Auth AuthConfig `section:"Auth"`
}

func LoadConfigs(files ...string) (*Config, error) {
    var config Config
    logger := hclog.Default()
    
    for _, file := range files {
        mgr := ini.New(file, logger)
        if err := mgr.Unmarshal(&config); err != nil {
            return nil, err
        }
    }
    
    return &config, nil
}
```

### Manual Section Parsing

```go
mgr := ini.New("config.ini", logger)

// Parse with explicit error handling
var dbConfig DatabaseConfig
section, err := mgr.ParseSection(&dbConfig, "Database")
if err != nil {
    log.Fatal("Failed to parse database section", err)
}

// Access raw values if needed
rawURL := section.String("URL", "", nil)
```

### Nested Configuration Structures

```go
type ElasticConfig struct {
    Hosts    []string `sep:"," def:"localhost:9200"`
    Username string   `def:"elastic"`
    Password string   `def:""`
}

type MonitoringConfig struct {
    Elastic ElasticConfig `section:"Elasticsearch"`
    Metrics MetricsConfig `section:"Metrics"`
}

type MetricsConfig struct {
    Enabled   bool   `def:"true"`
    Endpoint  string `def:"/metrics"`
    Interval  int    `def:"15" min:"5"`
}

type AppConfig struct {
    Monitoring MonitoringConfig `section:"Monitoring"`
}
```

## Best Practices

1. **Always provide defaults**: Use `def` tags to ensure your application works without configuration files
2. **Document your fields**: Field comments are automatically written to generated INI files
3. **Use validation tags**: `min`, `max`, and `valid` prevent invalid configurations
4. **Organize with sections**: Group related settings using `section` tags
5. **Log appropriately**: Pass a logger to Manager for debugging configuration issues
6. **Generate once, edit manually**: Generate base configuration files, then manually adjust for each environment

## Error Handling

The library uses the provided logger for warnings and non-fatal errors:

- Missing sections: Logs warning, returns empty section
- Missing keys: Logs warning, returns default value
- Parse errors: Logs warning, returns default value
- Validation failures: Logs warning, returns default value

Always check for errors only on struct unmarshaling operations.