package parser

import (
	"agentbridge/internal/models"
	"fmt"
	"regexp"
	"strings"
)

// CodeNodeParser handles Coze code node parsing.
type CodeNodeParser struct {
	*BaseNodeParser
}

func NewCodeNodeParser(variableRefSystem *models.VariableReferenceSystem) NodeParser {
	return &CodeNodeParser{
		BaseNodeParser: NewBaseNodeParser("5", variableRefSystem),
	}
}

// ParseNode parses Coze code node and converts code format.
func (p *CodeNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	// Validate node
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, fmt.Errorf("node validation failed: %w", err)
	}

	// Parse basic node information
	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeCode

	// Parse inputs and outputs
	node.Inputs = p.parseInputs(cozeNode)
	node.Outputs = p.parseOutputs(cozeNode)

	// Parse code runner configuration
	codeConfig, err := p.parseCodeConfiguration(cozeNode, node.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse code configuration: %w", err)
	}

	node.Config = codeConfig

	return node, nil
}

// parseCodeConfiguration extracts and converts code from Coze format to iFlytek format.
func (p *CodeNodeParser) parseCodeConfiguration(cozeNode CozeNode, inputs []models.Input) (models.CodeConfig, error) {
	// Extract code runner configuration from node inputs
	codeRunner := p.extractCodeRunner(cozeNode)
	if codeRunner == nil {
		return models.CodeConfig{}, fmt.Errorf("code runner configuration not found")
	}

	// Extract code and language from code runner
	cozeCode := p.extractCodeFromRunner(codeRunner)
	language := p.extractLanguageFromRunner(codeRunner)

	// Convert Coze code format to iFlytek format dynamically
	convertedCode, err := p.convertCozeCodeToSparkFormat(cozeCode, inputs)
	if err != nil {
		return models.CodeConfig{}, fmt.Errorf("code conversion failed: %w", err)
	}

	return models.CodeConfig{
		Language:      language,
		Code:          convertedCode,
		Dependencies:  []string{},
		IsInIteration: false,
		IterationID:   "",
	}, nil
}

// extractCodeRunner dynamically extracts code runner from node inputs.
func (p *CodeNodeParser) extractCodeRunner(cozeNode CozeNode) map[string]interface{} {
	if cozeNode.Data.Inputs == nil || cozeNode.Data.Inputs.CodeRunner == nil {
		return nil
	}

	// Convert to map for dynamic access
	if codeRunnerMap, ok := cozeNode.Data.Inputs.CodeRunner.(map[string]interface{}); ok {
		return codeRunnerMap
	}
	if codeRunnerMap, ok := cozeNode.Data.Inputs.CodeRunner.(map[interface{}]interface{}); ok {
		return p.convertInterfaceMapToStringMap(codeRunnerMap)
	}

	return nil
}

// extractCodeFromRunner dynamically extracts code content.
func (p *CodeNodeParser) extractCodeFromRunner(codeRunner map[string]interface{}) string {
	if code, exists := codeRunner["code"]; exists {
		if codeStr, ok := code.(string); ok {
			return codeStr
		}
	}
	return ""
}

// extractLanguageFromRunner dynamically extracts and converts language type.
func (p *CodeNodeParser) extractLanguageFromRunner(codeRunner map[string]interface{}) string {
	if lang, exists := codeRunner["language"]; exists {
		// Coze uses integer language codes, convert to string
		switch langCode := lang.(type) {
		case int:
			return p.convertLanguageCode(langCode)
		case float64:
			return p.convertLanguageCode(int(langCode))
		case string:
			return langCode
		}
	}
	return "python3" // Default language
}

// convertLanguageCode converts Coze language codes to standard names.
func (p *CodeNodeParser) convertLanguageCode(langCode int) string {
	languageMap := map[int]string{
		1: "javascript",
		2: "python2",
		3: "python3",
		4: "java",
		5: "go",
	}

	if language, exists := languageMap[langCode]; exists {
		return language
	}
	return "python3" // Default to python3
}

// convertCozeCodeToSparkFormat converts Coze async code format to iFlytek Spark format dynamically.
func (p *CodeNodeParser) convertCozeCodeToSparkFormat(cozeCode string, inputs []models.Input) (string, error) {
	if cozeCode == "" {
		return p.generateDefaultSparkCode(inputs), nil
	}

	// Step 1: Generate function signature dynamically from inputs
	sparkFunctionSignature := p.generateSparkFunctionSignature(inputs)

	// Step 2: Remove Coze-specific async/await and Args handling
	cleanedCode := p.removeCozeSpecificSyntax(cozeCode)

	// Step 3: Replace parameter access patterns dynamically
	convertedCode := p.replaceParameterAccess(cleanedCode, inputs)

	// Step 4: Combine signature with converted body
	return p.assembleSparkCode(sparkFunctionSignature, convertedCode), nil
}

// generateSparkFunctionSignature dynamically generates function signature from inputs.
func (p *CodeNodeParser) generateSparkFunctionSignature(inputs []models.Input) string {
	if len(inputs) == 0 {
		return "def main() -> dict:"
	}

	var params []string
	for _, input := range inputs {
		paramType := p.convertUnifiedDataTypeToString(input.Type)
		params = append(params, fmt.Sprintf("%s:%s", input.Name, paramType))
	}

	return fmt.Sprintf("def main(%s) -> dict:", strings.Join(params, ", "))
}

// convertUnifiedDataTypeToString converts unified data types to Python type annotations.
func (p *CodeNodeParser) convertUnifiedDataTypeToString(dataType models.UnifiedDataType) string {
	typeMap := map[models.UnifiedDataType]string{
		models.DataTypeString:      "str",
		models.DataTypeInteger:     "int",
		models.DataTypeFloat:       "float",
		models.DataTypeBoolean:     "bool",
		models.DataTypeArrayString: "list",
		models.DataTypeArrayObject: "list",
		models.DataTypeObject:      "dict",
	}

	if pyType, exists := typeMap[dataType]; exists {
		return pyType
	}
	return "str" // Default to string
}

// removeCozeSpecificSyntax removes Coze-specific syntax elements.
func (p *CodeNodeParser) removeCozeSpecificSyntax(code string) string {
	lines := strings.Split(code, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip Coze-specific lines
		if p.isCozeSpecificLine(trimmedLine) {
			continue
		}

		// Keep other lines
		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}

// isCozeSpecificLine identifies Coze-specific syntax lines to skip.
func (p *CodeNodeParser) isCozeSpecificLine(line string) bool {
	cozePatterns := []string{
		"async def main(args: Args)",
		"def main(args: Args)",
		"params = args.params",
		"from typing import",
		"import typing",
		"args.params.get",
		"args.text",
		"args.input_test",
	}

	// Special case: filter any line containing "async def main" regardless of format
	if strings.Contains(line, "async def main") {
		return true
	}

	// Special case: filter parameter access using args pattern
	if strings.Contains(line, "args.params") || strings.Contains(line, "args.text") || strings.Contains(line, "args.input_test") {
		return true
	}

	for _, pattern := range cozePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

// replaceParameterAccess dynamically replaces parameter access patterns.
func (p *CodeNodeParser) replaceParameterAccess(code string, inputs []models.Input) string {
	convertedCode := code

	// Build replacement map from inputs
	for _, input := range inputs {
		// Replace patterns like: str(params.get('paramName', '')) -> paramName
		patterns := []string{
			fmt.Sprintf(`str\(params\.get\(['"]%s['"], ['"]['"]?\)\)`, input.Name),
			fmt.Sprintf(`int\(params\.get\(['"]%s['"], 0\)\)`, input.Name),
			fmt.Sprintf(`float\(params\.get\(['"]%s['"], 0\.0\)\)`, input.Name),
			fmt.Sprintf(`bool\(params\.get\(['"]%s['"], False\)\)`, input.Name),
			fmt.Sprintf(`params\.get\(['"]%s['"], [^)]+\)`, input.Name),
		}

		for _, pattern := range patterns {
			re := regexp.MustCompile(pattern)
			convertedCode = re.ReplaceAllString(convertedCode, input.Name)
		}
	}

	return convertedCode
}

// assembleSparkCode combines signature with converted body, moving imports to top.
func (p *CodeNodeParser) assembleSparkCode(signature, body string) string {
	lines := strings.Split(body, "\n")
	var importLines []string
	var bodyLines []string

	// Separate import statements from other code
	inFunctionBody := false
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines at the beginning
		if trimmedLine == "" && !inFunctionBody {
			continue
		}

		// Check if this is an import statement
		if strings.HasPrefix(trimmedLine, "import ") || strings.HasPrefix(trimmedLine, "from ") {
			importLines = append(importLines, trimmedLine)
			continue
		}

		// Mark that we've entered function body content
		if trimmedLine != "" {
			inFunctionBody = true
		}

		if inFunctionBody {
			bodyLines = append(bodyLines, line)
		}
	}

	// Ensure proper indentation for function body
	var indentedBodyLines []string
	for _, line := range bodyLines {
		if strings.TrimSpace(line) == "" {
			indentedBodyLines = append(indentedBodyLines, "")
		} else if !strings.HasPrefix(line, "    ") && strings.TrimSpace(line) != "" {
			indentedBodyLines = append(indentedBodyLines, "    "+strings.TrimSpace(line))
		} else {
			indentedBodyLines = append(indentedBodyLines, line)
		}
	}

	// Assemble final code: imports first, then function definition with body
	var result []string

	// Add import statements at the top
	for _, importLine := range importLines {
		result = append(result, importLine)
	}

	// Add empty line after imports if there are any
	if len(importLines) > 0 {
		result = append(result, "")
	}

	// Add function signature
	result = append(result, signature)

	// Add function body
	result = append(result, indentedBodyLines...)

	return strings.Join(result, "\n")
}

// generateDefaultSparkCode generates default code when no code is provided.
func (p *CodeNodeParser) generateDefaultSparkCode(inputs []models.Input) string {
	signature := p.generateSparkFunctionSignature(inputs)
	defaultBody := "    return {'result': 'No implementation provided'}"
	return signature + "\n" + defaultBody
}

// convertInterfaceMapToStringMap converts map[interface{}]interface{} to map[string]interface{}.
func (p *CodeNodeParser) convertInterfaceMapToStringMap(interfaceMap map[interface{}]interface{}) map[string]interface{} {
	stringMap := make(map[string]interface{})
	for k, v := range interfaceMap {
		if key, ok := k.(string); ok {
			stringMap[key] = v
		}
	}
	return stringMap
}
