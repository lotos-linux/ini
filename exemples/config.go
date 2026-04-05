//go:generate go run github.com/lotos-linux/ini/cmd/ini-generator -i $GOFILE -o .
package conf

// ini:app.conf
// AppConfig is the main application configuration.
type AppConfig struct {
	Server   ServerConfig   `section:"Server"`
	Database DatabaseConfig `section:"Database"`
	Features FeaturesConfig `section:"Features"`
}

type ServerConfig struct {
	// Server listen address
	Host string `def:"localhost"`
	// Server port number
	Port int `def:"8080" min:"1024" max:"65535"`
	// Enable TLS mode
	TLS bool `def:"false"`
}

type DatabaseConfig struct {
	// Database connection URL
	URL string `def:"postgres://localhost:5432/db"`
	// Maximum connection pool size
	MaxConnections int `def:"10" min:"1" max:"100"`
}

type FeaturesConfig struct {
	// Enable debug mode
	Debug bool `def:"false"`
	// Allowed origins list
	Origins []string `sep:"," def:"localhost,127.0.0.1"`
	// Rate limit per second
	RateLimit int `def:"100" min:"0"`
}

// ini:theme.conf
// ThemeConfig holds the visual theme settings.
type ThemeConfig struct {
	UI     UIConfig     `section:"UI"`
	Colors ColorsConfig `section:"Colors"`
}

type UIConfig struct {
	// Theme name (light, dark, auto)
	Theme string `def:"auto" valid:"light,dark,auto"`
	// Icon size in pixels
	IconSize int `def:"24" min:"16" max:"64"`
	// Show tooltips
	Tooltips bool `def:"true"`
}

type ColorsConfig struct {
	// Primary color in hex format
	Primary string `def:"#3b82f6"`
	// Secondary color in hex format
	Secondary string `def:"#6b7280"`
}

type IgnoredStruct struct {
	// This struct does not have an ini: marker and will be ignored
	Value string `def:"test"`
}
