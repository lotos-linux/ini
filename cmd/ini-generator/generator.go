package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
)

type Generator struct {
	inputFile string
	outputDir string
	logger    *log.Logger
	debug     bool
}

type FieldInfo struct {
	Name     string
	Value    string // значение по умолчанию
	Default  string //原始默认值
	Valid    []string
	Min      string
	Max      string
	Sep      string
	Section  string
	Comment  string
	Kind     string
	IsSlice  bool
	ElemKind string
	IsBool   bool
}

type SectionInfo struct {
	Name string
	Keys []FieldInfo
}

type RootConfig struct {
	FileName   string
	StructName string
	Sections   []SectionInfo
}

func NewGenerator(inputFile, outputDir string, debug bool) *Generator {
	return &Generator{
		inputFile: inputFile,
		outputDir: outputDir,
		logger:    log.New(os.Stdout, "", 0),
		debug:     debug,
	}
}

func (g *Generator) logInfo(msg string, args ...interface{}) {
	g.logger.Printf("[INFO] "+msg, args...)
}

func (g *Generator) logWarn(msg string, args ...interface{}) {
	g.logger.Printf("[WARN] "+msg, args...)
}

func (g *Generator) logError(msg string, args ...interface{}) {
	g.logger.Printf("[ERROR] "+msg, args...)
}

func (g *Generator) logDebug(msg string, args ...interface{}) {
	if g.debug {
		g.logger.Printf("[DEBUG] "+msg, args...)
	}
}

func (g *Generator) Generate() error {
	g.logInfo("Starting configuration generator")
	g.logInfo("Input file: %s", g.inputFile)
	g.logInfo("Output directory: %s", g.outputDir)

	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		g.logError("Failed to create output directory: %v", err)
		return err
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, g.inputFile, nil, parser.ParseComments)
	if err != nil {
		g.logError("Failed to parse input file: %v", err)
		return err
	}
	g.logInfo("File parsed successfully")

	// Собираем все структуры
	structs := make(map[string]*ast.StructType)
	ast.Inspect(node, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if st, ok := ts.Type.(*ast.StructType); ok {
				structs[ts.Name.Name] = st
				g.logDebug("Found struct: %s", ts.Name.Name)
			}
		}
		return true
	})
	g.logInfo("Found %d struct definitions", len(structs))

	// Собираем комментарии к типам
	// Собираем комментарии к типам через GenDecl
	typeComments := make(map[string]string)

	ast.Inspect(node, func(n ast.Node) bool {
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.TYPE && gd.Doc != nil {
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					typeComments[ts.Name.Name] = gd.Doc.Text()
					g.logDebug("Struct %s has doc from GenDecl: %q", ts.Name.Name, gd.Doc.Text())
				}
			}
		}
		// также проверяем ts.Doc на случай, если комментарий прикреплен напрямую
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Doc != nil {
			if _, exists := typeComments[ts.Name.Name]; !exists {
				typeComments[ts.Name.Name] = ts.Doc.Text()
				g.logDebug("Struct %s has doc from TypeSpec: %q", ts.Name.Name, ts.Doc.Text())
			}
		}
		return true
	})

	// Находим корневые структуры
	rootConfigs := make([]RootConfig, 0)
	iniMarkerRegex := regexp.MustCompile(`ini:(\S+\.conf)`)

	for structName, comment := range typeComments {
		matches := iniMarkerRegex.FindStringSubmatch(comment)
		if len(matches) > 0 {
			fileName := matches[1]
			g.logInfo("Found root struct: %s -> %s", structName, fileName)

			if st, ok := structs[structName]; ok {
				rootConfig := RootConfig{
					FileName:   fileName,
					StructName: structName,
					Sections:   []SectionInfo{},
				}

				err := g.processStruct(st, structs, "", &rootConfig)
				if err != nil {
					g.logError("Failed to process struct %s: %v", structName, err)
					continue
				}

				rootConfigs = append(rootConfigs, rootConfig)
			} else {
				g.logError("Struct %s has marker but not found in parsed file", structName)
			}
		}
	}

	if len(rootConfigs) == 0 {
		g.logWarn("No structs with ini: marker found")
		return nil
	}

	totalKeys := 0
	for _, rc := range rootConfigs {
		keysCount := 0
		for _, section := range rc.Sections {
			keysCount += len(section.Keys)
		}
		totalKeys += keysCount

		if err := g.writeINIFile(rc); err != nil {
			g.logError("Failed to write %s: %v", rc.FileName, err)
			continue
		}
		g.logInfo("✓ Generated: %s (%d sections, %d keys)", rc.FileName, len(rc.Sections), keysCount)
	}

	g.logInfo("========================================")
	g.logInfo("Generation completed successfully")
	g.logInfo("Total files: %d", len(rootConfigs))
	g.logInfo("Total keys: %d", totalKeys)
	g.logInfo("Output directory: %s", g.outputDir)

	return nil
}

func (g *Generator) processStruct(st *ast.StructType, structs map[string]*ast.StructType, currentSection string, rootConfig *RootConfig) error {
	for _, field := range st.Fields.List {
		// Получаем теги
		tag := reflect.StructTag("")
		if field.Tag != nil {
			tag = reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		}

		// Получаем комментарий поля
		fieldComment := ""
		if field.Doc != nil {
			fieldComment = strings.TrimSpace(field.Doc.Text())
		}

		// Определяем тип поля
		var fieldType string
		var elemKind string
		var embeddedStruct *ast.StructType = nil
		isSlice := false
		isBool := false

		switch t := field.Type.(type) {
		case *ast.Ident:
			fieldType = t.Name
			if st, ok := structs[t.Name]; ok {
				embeddedStruct = st
			}
			isBool = t.Name == "bool"
		case *ast.SelectorExpr:
			fieldType = t.Sel.Name
			g.logDebug("External type: %s.%s", t.X, fieldType)
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				fieldType = ident.Name
				if st, ok := structs[ident.Name]; ok {
					embeddedStruct = st
				}
			}
		case *ast.ArrayType:
			isSlice = true
			if ident, ok := t.Elt.(*ast.Ident); ok {
				elemKind = ident.Name
				fieldType = "slice"
				isBool = elemKind == "bool"
				g.logDebug("Slice field with element type: %s", elemKind)
			}
		default:
			g.logWarn("Unknown field type for field")
			continue
		}

		// Проверяем тег section для вложенных структур
		sectionTag := tag.Get("section")

		// Если это вложенная структура с section тегом
		if sectionTag != "" && embeddedStruct != nil {
			g.logDebug("Processing embedded struct %s with section tag: %s", fieldType, sectionTag)
			if err := g.processStruct(embeddedStruct, structs, sectionTag, rootConfig); err != nil {
				return err
			}
			continue
		}

		// Если это анонимное поле (встроенная структура) без тега - игнорируем
		if len(field.Names) == 0 && embeddedStruct != nil {
			if sectionTag == "" {
				g.logDebug("Ignoring embedded struct %s without section tag", fieldType)
				continue
			}
		}

		// Пропускаем неэкспортируемые поля
		if len(field.Names) == 0 {
			continue
		}

		fieldName := field.Names[0].Name

		// Пропускаем приватные поля
		if !ast.IsExported(fieldName) {
			g.logDebug("Skipping private field: %s", fieldName)
			continue
		}

		// Определяем имя ключа
		keyName := fieldName

		// Определяем имя секции
		sectionName := currentSection
		if sectionName == "" {
			sectionName = "General"
		}

		// Получаем значения из тегов
		def := tag.Get("def")
		valid := tag.Get("valid")
		min := tag.Get("min")
		max := tag.Get("max")
		sep := tag.Get("sep")

		if def == "" && !isSlice {
			g.logWarn("Field %s has no default value (def tag missing)", fieldName)
		}

		var validList []string
		if valid != "" {
			validList = strings.Split(valid, ",")
			for i := range validList {
				validList[i] = strings.TrimSpace(validList[i])
			}
		}

		// Создаём информацию о поле
		fieldInfo := FieldInfo{
			Name:     keyName,
			Value:    def,
			Default:  def,
			Valid:    validList,
			Min:      min,
			Max:      max,
			Sep:      sep,
			Comment:  fieldComment,
			IsBool:   isBool,
			IsSlice:  isSlice,
			ElemKind: elemKind,
			Kind:     fieldType,
		}

		// Добавляем в секцию
		found := false
		for i := range rootConfig.Sections {
			if rootConfig.Sections[i].Name == sectionName {
				rootConfig.Sections[i].Keys = append(rootConfig.Sections[i].Keys, fieldInfo)
				found = true
				g.logDebug("  Added key %s to section %s", keyName, sectionName)
				break
			}
		}
		if !found {
			rootConfig.Sections = append(rootConfig.Sections, SectionInfo{
				Name: sectionName,
				Keys: []FieldInfo{fieldInfo},
			})
			g.logDebug("  Created section %s with key %s", sectionName, keyName)
		}
	}

	return nil
}

func (g *Generator) writeINIFile(rc RootConfig) error {
	outputPath := filepath.Join(g.outputDir, rc.FileName)
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var buffer bytes.Buffer

	for i, section := range rc.Sections {
		if i > 0 {
			buffer.WriteString("\n\n")
		}

		buffer.WriteString(fmt.Sprintf("[%s]\n", section.Name))

		for _, key := range section.Keys {
			// Формируем строку валидации
			var validationParts []string
			if key.IsBool {
				validationParts = append(validationParts, "true, false")
			}
			if len(key.Valid) > 0 {
				validationParts = append(validationParts, strings.Join(key.Valid, ", "))
			}
			if key.Min != "" && key.Max != "" {
				validationParts = append(validationParts, fmt.Sprintf("min: %s, max: %s", key.Min, key.Max))
			} else if key.Min != "" {
				validationParts = append(validationParts, fmt.Sprintf("min: %s", key.Min))
			} else if key.Max != "" {
				validationParts = append(validationParts, fmt.Sprintf("max: %s", key.Max))
			}

			validationStr := ""
			if len(validationParts) > 0 {
				validationStr = "(" + strings.Join(validationParts, ", ") + ")"
			}

			// Формируем строку комментария
			var commentLine string
			if key.Comment != "" || validationStr != "" {
				commentLine = "#"
				if key.Comment != "" {
					commentLine += " " + key.Comment
				}
				if validationStr != "" {
					if key.Comment != "" {
						commentLine += " " + validationStr
					} else {
						commentLine += " " + validationStr
					}
				}
				buffer.WriteString(commentLine + "\n")
			}

			// Пишем ключ и значение
			buffer.WriteString(fmt.Sprintf("%s = %s\n", key.Name, key.Value))
		}
	}

	_, err = file.Write(buffer.Bytes())
	return err
}
