package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var inputFile string
	var outputDir string
	var debug bool

	flag.StringVar(&inputFile, "i", "", "input go file")
	flag.StringVar(&outputDir, "o", ".", "output directory")
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()

	if inputFile == "" {
		printHelp()
		os.Exit(1)
	}

	gen := NewGenerator(inputFile, outputDir, debug)
	if err := gen.Generate(); err != nil {
		os.Exit(1)
	}
}

func printHelp() {
	helpText := `INI Configuration Generator

					Usage:
					config-generator -i <input.go> [-o <output_dir>]

					Options:
					-i <file>     Input Go file with struct definitions (required)
					-o <dir>      Output directory for generated INI files (default: ./output)
					-h, --help    Show this help message

					Example:
					config-generator -i ./pkg/config/config.go -o ./generated

					Input file format:
					// ini:filename.conf
					type StructName struct {
						Field string ` + "`def:\"value\"`" + `
					}

					Supported field tags:
					def    - Default value
					section- Section name for nested structs
					valid  - Comma-separated list of valid values
					min    - Minimum value for numbers
					max    - Maximum value for numbers
					sep    - Separator for slices (default: ",")
					`
	fmt.Println(helpText)
}
