package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CodeNodeGenerator generates Coze code nodes
type CodeNodeGenerator struct {
	idGenerator   *CozeIDGenerator
	isInIteration bool // Track if this code node is inside an iteration
}

// NewCodeNodeGenerator creates a code node generator
func NewCodeNodeGenerator() *CodeNodeGenerator {
	return &CodeNodeGenerator{
		idGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator sets the shared ID generator
func (g *CodeNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// SetIterationContext sets whether this code node is inside an iteration
func (g *CodeNodeGenerator) SetIterationContext(inIteration bool) {
	g.isInIteration = inIteration
}

// GetNodeType returns the node type this generator handles
func (g *CodeNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeCode
}

// ValidateNode validates the unified node before generation
func (g *CodeNodeGenerator) ValidateNode(unifiedNode *models.Node) error {
	if unifiedNode == nil {
		return fmt.Errorf("unified node is nil")
	}

	if unifiedNode.Type != models.NodeTypeCode {
		return fmt.Errorf("invalid node type: expected %s, got %s", models.NodeTypeCode, unifiedNode.Type)
	}

	if unifiedNode.Config == nil {
		return fmt.Errorf("node config is nil")
	}

	codeConfig, ok := common.AsCodeConfig(unifiedNode.Config)
	if !ok || codeConfig == nil {
		return fmt.Errorf("invalid config type: expected CodeConfig")
	}

	if codeConfig.Code == "" {
		return fmt.Errorf("code content is empty")
	}

	return nil
}

// GenerateNode generates a Coze workflow code node
func (g *CodeNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Extract code configuration
	codeConfig, ok := common.AsCodeConfig(unifiedNode.Config)
	if !ok || codeConfig == nil {
		return nil, fmt.Errorf("invalid code config type for node %s", unifiedNode.ID)
	}

	// Convert Python code to Coze format if needed
	convertedCode, err := g.convertPythonCodeToCozeFormat(codeConfig.Code, unifiedNode.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Python code to Coze format: %v", err)
	}

	// Generate input parameters from unified node inputs
	inputParams := g.generateInputParameters(unifiedNode)

	// Generate error handling settings
	errorSettings := g.generateErrorSettings()

	// Map language to Coze language code
	languageCode := g.mapLanguageToCozeCode(codeConfig.Language)

	// Create code node input structure based on context
	var codeInputs map[string]interface{}

	if g.isInIteration {
		// For iteration internal nodes, use clean format with only essential fields
		codeInputs = map[string]interface{}{
			"inputParameters": inputParams,   // Essential input parameters
			"settingOnError":  errorSettings, // Essential error handling
			"code":            convertedCode, // Essential code field
			"language":        languageCode,  // ✅ Essential: direct language field
		}
	} else {
		// For top-level nodes, use coderunner nesting
		codeInputs = map[string]interface{}{
			"inputparameters": inputParams,
			"settingonerror":  errorSettings,
			// Required null fields for Coze compatibility
			"nodebatchinfo":   nil,
			"llmparam":        nil,
			"outputemitter":   nil,
			"exit":            nil,
			"llm":             nil,
			"loop":            nil,
			"selector":        nil,
			"textprocessor":   nil,
			"subworkflow":     nil,
			"intentdetector":  nil,
			"databasenode":    nil,
			"httprequestnode": nil,
			"knowledge":       nil,
			"coderunner": map[string]interface{}{
				"code":     convertedCode,
				"language": languageCode,
			},
			"pluginapiparam":     nil,
			"variableaggregator": nil,
			"variableassigner":   nil,
			"qa":                 nil,
			"batch":              nil,
			"comment":            nil,
			"inputreceiver":      nil,
		}
	}

	// Generate node metadata
	nodeMeta := g.generateNodeMeta(unifiedNode)

	// Generate outputs
	outputs := g.generateOutputs(unifiedNode)

	return &CozeNode{
		ID:   cozeNodeID,
		Type: "5", // Coze code node type
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
		Data: &CozeNodeData{
			Meta: &CozeNodeMetaInfo{
				Title:       nodeMeta["title"].(string),
				Description: nodeMeta["description"].(string),
				Icon:        nodeMeta["icon"].(string),
				SubTitle:    nodeMeta["subtitle"].(string),
				MainColor:   nodeMeta["maincolor"].(string),
			},
			Outputs: g.convertToCozeNodeOutputs(outputs),
			Inputs:  codeInputs,
			Size:    nil,
		},
		Blocks:  []interface{}{},
		Edges:   []interface{}{},
		Version: "",
	}, nil
}

// GenerateSchemaNode generates a Coze schema code node
func (g *CodeNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate schema input parameters with proper RawMeta format
	schemaInputParams := g.generateSchemaInputParameters(unifiedNode)

	// Generate error handling settings for schema
	errorSettings := g.generateSchemaErrorSettings()

	// Extract code configuration
	codeConfig, ok := common.AsCodeConfig(unifiedNode.Config)
	if !ok || codeConfig == nil {
		return nil, fmt.Errorf("invalid code config type for node %s", unifiedNode.ID)
	}

	// Convert Python code to Coze format if needed
	convertedCode, err := g.convertPythonCodeToCozeFormat(codeConfig.Code, unifiedNode.Inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Python code to Coze format: %v", err)
	}

	// Map language to Coze language code
	languageCode := g.mapLanguageToCozeCode(codeConfig.Language)

	// Create schema inputs structure - use code and language for schema section
	schemaInputs := map[string]interface{}{
		"inputParameters": schemaInputParams,
		"settingOnError":  errorSettings,
		"code":            convertedCode,
		"language":        languageCode,
		// Required null fields for Coze compatibility
		"batch":              nil,
		"comment":            nil,
		"databasenode":       nil,
		"exit":               nil,
		"httprequestnode":    nil,
		"inputreceiver":      nil,
		"intentdetector":     nil,
		"knowledge":          nil,
		"llm":                nil,
		"llmparam":           nil,
		"loop":               nil,
		"nodebatchinfo":      nil,
		"outputemitter":      nil,
		"pluginapiparam":     nil,
		"qa":                 nil,
		"selector":           nil,
		"subworkflow":        nil,
		"textprocessor":      nil,
		"variableaggregator": nil,
		"variableassigner":   nil,
	}

	// Generate schema outputs
	schemaOutputs := g.generateSchemaOutputs(unifiedNode)

	return &CozeSchemaNode{
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Title:       g.getNodeTitle(unifiedNode),
				Description: g.getNodeDescription(unifiedNode),
				Icon:        g.getNodeIcon(),
				SubTitle:    "代码",
				MainColor:   "#00B2B2",
			},
			Inputs:  schemaInputs,
			Outputs: schemaOutputs,
		},
		ID:   cozeNodeID,
		Type: "5", // Coze code node type
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
	}, nil
}

// generateInputParameters generates input parameters from unified node inputs
func (g *CodeNodeGenerator) generateInputParameters(unifiedNode *models.Node) []map[string]interface{} {
	var inputParams []map[string]interface{}

	if unifiedNode.Inputs == nil {
		return inputParams
	}

	for _, input := range unifiedNode.Inputs {
		if g.isInIteration {
			// CRITICAL: For iteration context, use format matching Coze iteration examples
			param := map[string]interface{}{
				"input": map[string]interface{}{
					"type":  g.mapUnifiedTypeToCozeType(input.Type), // lowercase 'type'
					"value": g.generateIterationInputValue(input),   // lowercase 'value'
				},
				"name":      input.Name,
				"left":      nil,
				"right":     nil,
				"variables": []interface{}{},
			}
			inputParams = append(inputParams, param)
		} else {
			// For top-level context, use uppercase format
			param := map[string]interface{}{
				"input": map[string]interface{}{
					"Type":  g.mapUnifiedTypeToCozeType(input.Type),
					"Value": g.generateInputValue(input),
				},
				"name":      input.Name,
				"left":      nil,
				"right":     nil,
				"variables": []interface{}{},
			}
			inputParams = append(inputParams, param)
		}
	}

	return inputParams
}

// generateIterationInputValue generates input value configuration for iteration context
func (g *CodeNodeGenerator) generateIterationInputValue(input models.Input) map[string]interface{} {
	if input.Reference != nil {
		// Determine blockID - use directly if already a Coze ID (numeric), otherwise map it
		blockID := input.Reference.NodeID
		if !g.isCozeNodeID(blockID) {
			blockID = g.idGenerator.MapToCozeNodeID(blockID)
		}

		// Reference to another node's output - use iteration format with lowercase fields
		return map[string]interface{}{
			"type": "ref",
			"content": map[string]interface{}{
				"blockID": blockID,
				"name":    g.mapOutputFieldNameForCoze(input.Reference.NodeID, input.Reference.OutputName),
				"source":  "block-output",
			},
			"rawMeta": map[string]interface{}{ // CRITICAL: Use camelCase 'rawMeta' for iteration context
				"type": g.getCozeTypeCode(input.Type),
			},
		}
	} else if input.Default != nil {
		// Literal value
		return map[string]interface{}{
			"type":    "literal",
			"content": fmt.Sprintf("%v", input.Default),
			"rawMeta": map[string]interface{}{
				"type": g.getCozeTypeCode(input.Type),
			},
		}
	}

	// Empty value
	return map[string]interface{}{
		"type":    "literal",
		"content": "",
		"rawMeta": map[string]interface{}{
			"type": g.getCozeTypeCode(input.Type),
		},
	}
}

// generateInputValue generates input value configuration
func (g *CodeNodeGenerator) generateInputValue(input models.Input) map[string]interface{} {
	if input.Reference != nil {
		// Determine blockID - use directly if already a Coze ID (numeric), otherwise map it
		blockID := input.Reference.NodeID
		if !g.isCozeNodeID(blockID) {
			blockID = g.idGenerator.MapToCozeNodeID(blockID)
		}

		// Reference to another node's output
		return map[string]interface{}{
			"type": "ref",
			"content": map[string]interface{}{
				"blockID": blockID,
				"name":    g.mapOutputFieldNameForCoze(input.Reference.NodeID, input.Reference.OutputName),
				"source":  "block-output",
			},
			"rawMeta": map[string]interface{}{
				"type": g.getCozeTypeCode(input.Type),
			},
		}
	} else if input.Default != nil {
		// Literal value
		return map[string]interface{}{
			"type":    "literal",
			"content": fmt.Sprintf("%v", input.Default),
			"rawMeta": map[string]interface{}{
				"type": g.getCozeTypeCode(input.Type),
			},
		}
	}

	// Empty value
	return map[string]interface{}{
		"type":    "literal",
		"content": "",
		"rawmeta": map[string]interface{}{
			"type": g.getCozeTypeCode(input.Type),
		},
	}
}

// generateOutputs generates output definitions from unified node outputs
func (g *CodeNodeGenerator) generateOutputs(unifiedNode *models.Node) []map[string]interface{} {
	var outputs []map[string]interface{}

	if unifiedNode.Outputs == nil {
		return outputs
	}

	for _, output := range unifiedNode.Outputs {
		outputDef := map[string]interface{}{
			"name": output.Name,
			"type": g.mapUnifiedTypeToCozeType(output.Type),
		}

		// Add schema for array types
		if g.isArrayType(output.Type) {
			outputDef["schema"] = map[string]interface{}{
				"type": g.getArrayElementType(output.Type),
			}
		}

		outputs = append(outputs, outputDef)
	}

	return outputs
}

// generateErrorSettings generates error handling settings
func (g *CodeNodeGenerator) generateErrorSettings() map[string]interface{} {
	if g.isInIteration {
		// CRITICAL: For iteration context, use camelCase naming
		return map[string]interface{}{
			"dataonerr":   "",
			"switch":      false,
			"processType": 1,
			"retryTimes":  0,
			"timeoutMs":   60000, // 60 seconds timeout for code execution
			"ext":         nil,
		}
	} else {
		// For top-level context, use camelCase naming
		return map[string]interface{}{
			"dataonerr":   "",
			"switch":      false,
			"processType": 1,
			"retryTimes":  0,
			"timeoutMs":   60000, // 60 seconds timeout for code execution
			"ext":         nil,
		}
	}
}

// mapOutputFieldNameForCoze maps output field names from unified DSL to Coze platform format
func (g *CodeNodeGenerator) mapOutputFieldNameForCoze(nodeID, outputName string) string {
	// Map classifier output names from iFlytek format to Coze format
	if outputName == "class_name" {
		// iFlytek classifier outputs "class_name", but Coze uses "classificationId"
		return "classificationId"
	}

	// Add more mappings as needed for other node types
	// Example: if outputName == "some_other_field" { return "mappedField" }

	// Default: return original name if no mapping needed
	return outputName
}

// generateNodeMeta generates node metadata
func (g *CodeNodeGenerator) generateNodeMeta(unifiedNode *models.Node) map[string]interface{} {
	title := unifiedNode.Title
	if title == "" {
		title = "代码节点"
	}

	description := unifiedNode.Description
	if description == "" {
		description = "编写代码，处理输入变量来生成返回值"
	}

	return map[string]interface{}{
		"title":       title,
		"description": description,
		"icon":        "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Code-v2.jpg",
		"subtitle":    "代码",
		"maincolor":   "#00B2B2",
	}
}

// mapLanguageToCozeCode maps unified language to Coze language code
func (g *CodeNodeGenerator) mapLanguageToCozeCode(language string) int {
	switch language {
	case "python3", "python":
		return 3
	case "javascript", "js":
		return 1
	default:
		return 3 // Default to Python
	}
}

// mapUnifiedTypeToCozeType maps unified data type to Coze type string
func (g *CodeNodeGenerator) mapUnifiedTypeToCozeType(unifiedType models.UnifiedDataType) string {
	switch unifiedType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "float"
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeArrayString, models.DataTypeArrayInteger, models.DataTypeArrayFloat,
		models.DataTypeArrayNumber, models.DataTypeArrayBoolean, models.DataTypeArrayObject:
		return "list" // Coze uses "list" for all array types
	case models.DataTypeObject:
		return "object"
	default:
		return "string"
	}
}

// convertToCozeNodeOutputs converts map outputs to CozeNodeOutput array
func (g *CodeNodeGenerator) convertToCozeNodeOutputs(outputs []map[string]interface{}) []CozeNodeOutput {
	var cozeOutputs []CozeNodeOutput

	for _, output := range outputs {
		cozeOutput := CozeNodeOutput{
			Name:     output["name"].(string),
			Type:     output["type"].(string),
			Required: false,
		}
		cozeOutputs = append(cozeOutputs, cozeOutput)
	}

	return cozeOutputs
}

// getCozeTypeCode returns Coze type code for rawmeta
func (g *CodeNodeGenerator) getCozeTypeCode(unifiedType models.UnifiedDataType) int {
	switch unifiedType {
	case models.DataTypeString:
		return 1
	case models.DataTypeInteger:
		return 2
	case models.DataTypeBoolean:
		return 3
	case models.DataTypeFloat, models.DataTypeNumber:
		return 4
	case models.DataTypeArrayString, models.DataTypeArrayInteger, models.DataTypeArrayFloat,
		models.DataTypeArrayNumber, models.DataTypeArrayBoolean, models.DataTypeArrayObject:
		return 5 // All array types use code 5
	case models.DataTypeObject:
		return 6
	default:
		return 1 // Default to string
	}
}

// generateSchemaInputParameters generates schema input parameters with proper RawMeta format
func (g *CodeNodeGenerator) generateSchemaInputParameters(unifiedNode *models.Node) []CozeInputParameter {
	var schemaInputParams []CozeInputParameter

	if unifiedNode.Inputs == nil {
		return schemaInputParams
	}

	for _, input := range unifiedNode.Inputs {
		cozeType := g.mapUnifiedTypeToCozeType(input.Type)

		if input.Reference != nil && input.Reference.Type == models.ReferenceTypeNodeOutput {
			// Variable reference input for schema - schema uses RawMeta with uppercase M
			schemaInput := CozeInputParameter{
				Name: input.Name,
				Input: &CozeInputValue{
					Type: cozeType,
					Value: &CozeInputRef{
						Type: "ref",
						Content: &CozeRefContent{
							BlockID: g.idGenerator.MapToCozeNodeID(input.Reference.NodeID),
							Name:    input.Reference.OutputName,
							Source:  "block-output",
						},
						RawMeta: &CozeRawMeta{ // Schema section uses RawMeta with uppercase M
							Type: g.getCozeTypeCode(input.Type),
						},
					},
				},
			}
			schemaInputParams = append(schemaInputParams, schemaInput)
		}
	}

	return schemaInputParams
}

// generateSchemaErrorSettings generates error handling settings for schema
func (g *CodeNodeGenerator) generateSchemaErrorSettings() map[string]interface{} {
	return map[string]interface{}{
		"processType": 1,
		"retryTimes":  0,
		"timeoutMs":   60000,
	}
}

// generateSchemaOutputs generates schema outputs
func (g *CodeNodeGenerator) generateSchemaOutputs(unifiedNode *models.Node) []CozeNodeOutput {
	var outputs []CozeNodeOutput

	if unifiedNode.Outputs == nil {
		return outputs
	}

	for _, output := range unifiedNode.Outputs {
		cozeOutput := CozeNodeOutput{
			Name:     output.Name,
			Type:     g.mapUnifiedTypeToCozeType(output.Type),
			Required: false,
		}

		// Add schema for array types in CozeNodeOutput
		if g.isArrayType(output.Type) {
			cozeOutput.Schema = &CozeOutputSchema{
				Type: g.getArrayElementType(output.Type),
			}
		}

		outputs = append(outputs, cozeOutput)
	}

	return outputs
}

// getNodeTitle returns node title with uniqueness
func (g *CodeNodeGenerator) getNodeTitle(unifiedNode *models.Node) string {
	if unifiedNode.Title != "" {
		return unifiedNode.Title
	}
	return "代码节点"
}

// getNodeDescription returns node description
func (g *CodeNodeGenerator) getNodeDescription(unifiedNode *models.Node) string {
	if unifiedNode.Description != "" {
		return unifiedNode.Description
	}
	return "面向开发者提供代码开发能力，目前仅支持python语言，允许使用该节点已定义的变量作为参数传入，返回语句用于输出函数的结果"
}

// getNodeIcon returns code node icon
func (g *CodeNodeGenerator) getNodeIcon() string {
	return "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Code-v2.jpg"
}

// convertPythonCodeToCozeFormat converts iFlytek Python code format to Coze format
func (g *CodeNodeGenerator) convertPythonCodeToCozeFormat(originalCode string, inputParams []models.Input) (string, error) {
	// Check if this is Python code that needs conversion
	if !strings.Contains(originalCode, "def main(") {
		return originalCode, nil
	}

	// Check if already in Coze format (async def main with Args)
	if strings.Contains(originalCode, "async def main(args: Args)") {
		return originalCode, nil
	}

	// Extract parameter names from input parameters
	var paramNames []string
	for _, param := range inputParams {
		paramNames = append(paramNames, param.Name)
	}

	// Use regex to match function definition
	funcPattern := regexp.MustCompile(`def\s+main\s*\([^)]*\)\s*->\s*[^:]*:`)
	funcMatch := funcPattern.FindString(originalCode)

	if funcMatch == "" {
		// No function definition found, return as is
		return originalCode, nil
	}

	// Replace function signature
	newSignature := "async def main(args: Args) -> Output:"
	codeAfterSignature := funcPattern.ReplaceAllString(originalCode, newSignature)

	// Generate parameter extraction code
	var paramExtractions []string
	paramExtractions = append(paramExtractions, "    params = args.params")

	for _, paramName := range paramNames {
		// Use str() to ensure string type for all parameters as seen in Coze examples
		paramExtractions = append(paramExtractions, fmt.Sprintf("    %s = str(params.get('%s', ''))", paramName, paramName))
	}

	// Find the position to insert parameter extraction (after function definition)
	lines := strings.Split(codeAfterSignature, "\n")
	var resultLines []string

	for _, line := range lines {
		resultLines = append(resultLines, line)

		// Insert parameter extraction after the function definition line
		if strings.Contains(line, "async def main(args: Args) -> Output:") {
			// Add empty line and parameter extractions
			resultLines = append(resultLines, "")
			for _, paramExtraction := range paramExtractions {
				resultLines = append(resultLines, paramExtraction)
			}
			resultLines = append(resultLines, "")
		}
	}

	return strings.Join(resultLines, "\n"), nil
}

// isArrayType checks if the unified type is an array type
func (g *CodeNodeGenerator) isArrayType(unifiedType models.UnifiedDataType) bool {
	return unifiedType == models.DataTypeArrayString || unifiedType == models.DataTypeArrayInteger ||
		unifiedType == models.DataTypeArrayFloat || unifiedType == models.DataTypeArrayNumber ||
		unifiedType == models.DataTypeArrayBoolean || unifiedType == models.DataTypeArrayObject
}

// getArrayElementType returns the element type for array types
func (g *CodeNodeGenerator) getArrayElementType(unifiedType models.UnifiedDataType) string {
	switch unifiedType {
	case models.DataTypeArrayString:
		return "string"
	case models.DataTypeArrayInteger:
		return "integer"
	case models.DataTypeArrayFloat:
		return "float"
	case models.DataTypeArrayNumber:
		return "float" // Map generic number to float for Coze
	case models.DataTypeArrayBoolean:
		return "boolean"
	case models.DataTypeArrayObject:
		return "object"
	default:
		return "string" // Default fallback
	}
}

// isCozeNodeID checks if the given ID is already a Coze numeric node ID
func (g *CodeNodeGenerator) isCozeNodeID(nodeID string) bool {
	// Coze node IDs are numeric strings (e.g., "197161", "100001", "900001")
	// If it can be parsed as an integer and doesn't contain special prefixes, it's likely a Coze ID
	if _, err := strconv.Atoi(nodeID); err == nil {
		return true
	}
	return false
}
