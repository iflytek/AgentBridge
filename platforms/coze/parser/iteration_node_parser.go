package parser

import (
	"agentbridge/internal/models"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IterationNodeParser parses Coze iteration nodes.
type IterationNodeParser struct {
	*BaseNodeParser
}

func NewIterationNodeParser(variableRefSystem *models.VariableReferenceSystem) *IterationNodeParser {
	return &IterationNodeParser{
		BaseNodeParser: NewBaseNodeParser("21", variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *IterationNodeParser) GetSupportedType() string {
	return "21"
}

// ParseNode parses a Coze iteration node into unified DSL.
func (p *IterationNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, err
	}

	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeIteration

	// Parse iteration configuration
	config, err := p.parseIterationConfig(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse iteration config: %w", err)
	}
	node.Config = config

	// Parse inputs from inputParameters - use specialized iteration input parsing
	node.Inputs = p.parseIterationInputs(cozeNode)

	// Parse outputs
	node.Outputs = p.parseIterationOutputs(cozeNode)

	// Register output name mapping for variable reference resolution
	// This ensures other nodes referencing this iteration node use the correct output names
	p.registerIterationOutputMapping(cozeNode, node)

	return node, nil
}

// parseIterationConfig extracts iteration configuration from Coze node
func (p *IterationNodeParser) parseIterationConfig(cozeNode CozeNode) (models.IterationConfig, error) {
	config := models.IterationConfig{}

	// Extract loop configuration from inputs
	if cozeNode.Data.Inputs != nil && cozeNode.Data.Inputs.Loop != nil {
		loopConfig := p.extractLoopParams(cozeNode.Data.Inputs.Loop)

		// Set iterator configuration
		config.Iterator = models.IteratorConfig{
			InputType: p.getStringParam(loopConfig, "loopType", "array"),
		}

		// Find source node and output from inputParameters
		if len(cozeNode.Data.Inputs.InputParameters) > 0 {
			for _, param := range cozeNode.Data.Inputs.InputParameters {
				if param.Input.Value.Type == "ref" {
					config.Iterator.SourceNode = param.Input.Value.Content.BlockID
					config.Iterator.SourceOutput = param.Input.Value.Content.Name
					break
				}
			}
		}

		// Set execution configuration
		config.Execution = models.ExecutionConfig{
			IsParallel:      false, // Default for Coze iteration (sequential)
			ParallelNums:    1,
			ErrorHandleMode: "stop", // Default error handling
		}

		// Set output type
		config.OutputType = "array"
	}

	// Parse sub-workflow from blocks
	if len(cozeNode.Blocks) > 0 {
		subWorkflow, err := p.parseSubWorkflow(cozeNode.Blocks, cozeNode.ID)
		if err != nil {
			return config, fmt.Errorf("failed to parse sub-workflow: %w", err)
		}
		config.SubWorkflow = subWorkflow
	}

	// Parse output selector from outputs
	if len(cozeNode.Data.Outputs) > 0 {
		config.OutputSelector = p.parseOutputSelector(cozeNode.Data.Outputs)
	}

	return config, nil
}

// extractLoopParams converts Loop data to map for easier access
func (p *IterationNodeParser) extractLoopParams(loop interface{}) map[string]interface{} {
	params := make(map[string]interface{})

	if loopMap, ok := loop.(map[string]interface{}); ok {
		for key, value := range loopMap {
			params[key] = value
		}
	}

	return params
}

// parseSubWorkflow parses internal blocks to create sub-workflow
func (p *IterationNodeParser) parseSubWorkflow(blocks []interface{}, iterationID string) (models.SubWorkflowConfig, error) {
	subWorkflow := models.SubWorkflowConfig{
		Nodes: []models.Node{},
		Edges: []models.Edge{},
	}

	// Parse each block as a node
	for _, block := range blocks {
		if blockMap, ok := block.(map[string]interface{}); ok {
			// Create a CozeNode from the block data
			blockNode := p.convertBlockToCozeNode(blockMap)

			// Parse the block node based on its type - use actual type-specific parsing
			parsedNode, err := p.parseSpecificBlockType(blockNode, iterationID)
			if err != nil {
				return subWorkflow, fmt.Errorf("failed to parse block node: %w", err)
			}

			if parsedNode != nil {
				// Set iteration configuration for the sub-node using actual iteration ID
				p.setIterationNodeConfig(parsedNode, iterationID)
				subWorkflow.Nodes = append(subWorkflow.Nodes, *parsedNode)
			}
			// If parsedNode is nil, the node was skipped - no action needed
		}
	}

	return subWorkflow, nil
}

// parseSpecificBlockType parses block nodes based on their actual type using main layer parsers
func (p *IterationNodeParser) parseSpecificBlockType(cozeNode CozeNode, iterationID string) (*models.Node, error) {
	switch cozeNode.Type {
	case "3": // LLM node - use main layer LLM parser
		return p.parseLLMBlock(cozeNode, iterationID)
	case "5": // Code node - use detailed code parsing
		return p.parseCodeBlockDetailed(cozeNode, iterationID)
	case "8": // Selector/Branch node - use main layer selector parser
		return p.parseSelectorBlock(cozeNode, iterationID)
	case "22": // Classifier node - use main layer classifier parser
		return p.parseClassifierBlock(cozeNode, iterationID)
	default:
		// For unsupported types, skip the node instead of creating basic code node
		fmt.Printf("⚠️  Skipping unsupported iteration block type '%s' (ID: %s, Title: %s)\n",
			cozeNode.Type, cozeNode.ID, cozeNode.Data.Meta.Title)
		return nil, nil // Return nil to indicate the node should be skipped
	}
}

// parseCodeBlockDetailed parses code blocks with full code extraction and conversion
func (p *IterationNodeParser) parseCodeBlockDetailed(cozeNode CozeNode, iterationID string) (*models.Node, error) {
	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeCode

	// Extract actual code from the block data
	codeConfig, err := p.extractCodeFromBlockData(cozeNode, iterationID)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code from block: %w", err)
	}

	node.Config = codeConfig

	// Parse inputs and outputs from the actual block structure
	node.Inputs = p.parseCodeBlockInputs(cozeNode, iterationID)
	node.Outputs = p.parseCodeBlockOutputs(cozeNode)

	return node, nil
}

// extractCodeFromBlockData extracts code configuration from iteration block data
func (p *IterationNodeParser) extractCodeFromBlockData(cozeNode CozeNode, iterationID string) (models.CodeConfig, error) {
	var code string
	var language string = "python3" // Default

	// Try to extract code from CodeRunner (similar to main layer logic)
	if cozeNode.Data.Inputs != nil && cozeNode.Data.Inputs.CodeRunner != nil {
		if codeRunnerMap, ok := cozeNode.Data.Inputs.CodeRunner.(map[string]interface{}); ok {
			// Extract code
			if codeField, exists := codeRunnerMap["code"]; exists {
				if codeStr, ok := codeField.(string); ok {
					code = codeStr
				}
			}
			// Extract language
			if langField, exists := codeRunnerMap["language"]; exists {
				switch langCode := langField.(type) {
				case int:
					language = p.convertLanguageCode(langCode)
				case float64:
					language = p.convertLanguageCode(int(langCode))
				case string:
					language = langCode
				default:
					language = "python3"
				}
			}
		}
	}

	// Fallback: extract from direct inputs structure
	if code == "" && cozeNode.Data.Inputs != nil {
		if codeField, ok := p.extractFieldFromInputs(cozeNode.Data.Inputs, "code"); ok {
			if codeStr, ok := codeField.(string); ok {
				code = codeStr
			}
		}
		if langField, ok := p.extractFieldFromInputs(cozeNode.Data.Inputs, "language"); ok {
			if langInt, ok := langField.(float64); ok {
				language = p.convertLanguageCode(int(langInt))
			} else if langInt, ok := langField.(int); ok {
				language = p.convertLanguageCode(langInt)
			}
		}
	}

	if code == "" {
		return models.CodeConfig{}, fmt.Errorf("no code found in iteration block data for node: %s", cozeNode.ID)
	}

	// Convert code format properly using main layer conversion logic
	inputs := p.parseCodeBlockInputs(cozeNode, iterationID)
	convertedCode, err := p.convertIterationCodeFormat(code, inputs)
	if err != nil {
		return models.CodeConfig{}, fmt.Errorf("code conversion failed: %w", err)
	}

	return models.CodeConfig{
		Language:      language,
		Code:          convertedCode,
		Dependencies:  []string{},
		IsInIteration: true,
		IterationID:   "",
	}, nil
}

// parseCodeBlockInputs parses inputs from code block in iteration
func (p *IterationNodeParser) parseCodeBlockInputs(cozeNode CozeNode, iterationID string) []models.Input {
	var inputs []models.Input

	// Use the same input parsing logic as the main layer code node parser
	// This dynamically extracts inputs from the node's inputParameters
	if cozeNode.Data.Inputs != nil && cozeNode.Data.Inputs.InputParameters != nil {
		for _, param := range cozeNode.Data.Inputs.InputParameters {
			input := models.Input{
				Name:        param.Name,
				Label:       param.Name,
				Type:        p.convertDataType(param.Input.Type),
				Required:    true,
				Description: "",
			}

			// Parse reference if exists
			if param.Input.Value.Type == "ref" {
				// Map main workflow start node references to iteration start node
				mappedNodeID, mappedOutputName := p.mapIterationVariableReferenceWithOutput(
					param.Input.Value.Content.BlockID,
					param.Input.Value.Content.Name,
					iterationID,
				)

				input.Reference = &models.VariableReference{
					Type:       models.ReferenceTypeNodeOutput,
					NodeID:     mappedNodeID,
					OutputName: mappedOutputName,
					DataType:   input.Type,
				}
			}

			inputs = append(inputs, input)
		}
	}

	// Fallback: try to extract from CodeRunner structure if InputParameters is not available
	if len(inputs) == 0 && cozeNode.Data.Inputs != nil && cozeNode.Data.Inputs.CodeRunner != nil {
		if codeRunnerMap, ok := cozeNode.Data.Inputs.CodeRunner.(map[string]interface{}); ok {
			if inputParams, exists := codeRunnerMap["inputParameters"]; exists {
				if paramArray, ok := inputParams.([]interface{}); ok {
					for _, param := range paramArray {
						if paramMap, ok := param.(map[string]interface{}); ok {
							input := p.parseCodeBlockInputParam(paramMap, iterationID)
							if input.Name != "" {
								// Ensure string type for iteration items
								input.Type = models.DataTypeString
								inputs = append(inputs, input)
							}
						}
					}
				}
			}
		}
	}

	return inputs
}

// parseCodeBlockInputParam parses a single input parameter from code block
func (p *IterationNodeParser) parseCodeBlockInputParam(paramMap map[string]interface{}, iterationID string) models.Input {
	input := models.Input{}

	if name, ok := paramMap["name"].(string); ok {
		input.Name = name
		input.Label = name
	}

	if inputData, ok := paramMap["input"].(map[string]interface{}); ok {
		// Parse type (check both cases)
		if inputType, ok := inputData["Type"].(string); ok {
			input.Type = p.convertDataType(inputType)
		} else if inputType, ok := inputData["type"].(string); ok {
			input.Type = p.convertDataType(inputType)
		}

		// Parse reference if exists (check both cases)
		var valueMap map[string]interface{}
		if value, ok := inputData["Value"].(map[string]interface{}); ok {
			valueMap = value
		} else if value, ok := inputData["value"].(map[string]interface{}); ok {
			valueMap = value
		}

		if valueMap != nil {
			if valueType, ok := valueMap["type"].(string); ok && valueType == "ref" {
				if content, ok := valueMap["content"].(map[string]interface{}); ok {
					blockID := p.getStringFromMap(content, "blockID", "")
					outputName := p.getStringFromMap(content, "name", "")

					// Map main workflow start node references to iteration start node
					// In Coze iteration, references to main workflow start node should be mapped to iteration start node
					mappedNodeID, mappedOutputName := p.mapIterationVariableReferenceWithOutput(blockID, outputName, iterationID)

					input.Reference = &models.VariableReference{
						Type:       models.ReferenceTypeNodeOutput,
						NodeID:     mappedNodeID,
						OutputName: mappedOutputName,
						DataType:   input.Type,
					}
				}
			}
		}
	}

	input.Required = true
	input.Description = ""

	return input
}

// parseCodeBlockOutputs parses outputs from code block
func (p *IterationNodeParser) parseCodeBlockOutputs(cozeNode CozeNode) []models.Output {
	var outputs []models.Output

	if cozeNode.Data.Outputs != nil {
		for _, output := range cozeNode.Data.Outputs {
			outputs = append(outputs, models.Output{
				Name:        output.Name,
				Type:        p.convertDataType(output.Type),
				Description: "",
				Required:    true,
			})
		}
	}

	// Ensure at least one output exists
	if len(outputs) == 0 {
		outputs = append(outputs, models.Output{
			Name:        "result",
			Type:        models.DataTypeString,
			Description: "Code execution result",
			Required:    true,
		})
	}

	return outputs
}

// setIterationNodeConfig sets iteration configuration for sub-nodes
func (p *IterationNodeParser) setIterationNodeConfig(node *models.Node, iterationID string) {
	switch config := node.Config.(type) {
	case models.CodeConfig:
		config.IsInIteration = true
		config.IterationID = iterationID
		node.Config = config
	case models.LLMConfig:
		config.IsInIteration = true
		config.IterationID = iterationID
		node.Config = config
	case models.StartConfig:
		config.IsInIteration = true
		config.ParentID = iterationID
		node.Config = config
	case models.ClassifierConfig:
		config.IsInIteration = true
		config.IterationID = iterationID
		node.Config = config
	case models.ConditionConfig:
		config.IsInIteration = true
		config.IterationID = iterationID
		node.Config = config
	}
}

// convertBlockToCozeNode converts a block map to CozeNode structure
func (p *IterationNodeParser) convertBlockToCozeNode(blockMap map[string]interface{}) CozeNode {
	node := CozeNode{}

	if id, ok := blockMap["id"].(string); ok {
		node.ID = id
	}
	if nodeType, ok := blockMap["type"].(string); ok {
		node.Type = nodeType
	}

	// Parse data if exists
	if data, ok := blockMap["data"].(map[string]interface{}); ok {
		node.Data = p.parseCozeNodeDataFromBlock(data)
	}

	// Parse meta if exists
	if meta, ok := blockMap["meta"].(map[string]interface{}); ok {
		node.Meta = p.parseCozeNodeMeta(meta)
	}

	return node
}

// parseCozeNodeDataFromBlock parses node data from iteration block data
func (p *IterationNodeParser) parseCozeNodeDataFromBlock(data map[string]interface{}) CozeNodeData {
	nodeData := CozeNodeData{}

	// Parse meta - check both "nodeMeta" and "meta" keys
	if meta, ok := data["meta"].(map[string]interface{}); ok {
		nodeData.Meta = CozeDataMeta{
			Title:       p.getStringFromMap(meta, "title", ""),
			Description: p.getStringFromMap(meta, "description", ""),
			Icon:        p.getStringFromMap(meta, "icon", ""),
		}
	} else if meta, ok := data["nodeMeta"].(map[string]interface{}); ok {
		nodeData.Meta = CozeDataMeta{
			Title:       p.getStringFromMap(meta, "title", ""),
			Description: p.getStringFromMap(meta, "description", ""),
			Icon:        p.getStringFromMap(meta, "icon", ""),
		}
	}

	// Parse outputs
	if outputs, ok := data["outputs"].([]interface{}); ok {
		for _, output := range outputs {
			if outputMap, ok := output.(map[string]interface{}); ok {
				nodeData.Outputs = append(nodeData.Outputs, CozeOutput{
					Name: p.getStringFromMap(outputMap, "name", ""),
					Type: p.getStringFromMap(outputMap, "type", "string"),
				})
			}
		}
	}

	// Parse inputs - store the actual inputs data for code extraction
	if inputs, ok := data["inputs"].(map[string]interface{}); ok {
		nodeData.Inputs = &CozeNodeInputs{}

		// Initialize CodeRunner structure to store both code and inputParameters
		nodeData.Inputs.CodeRunner = make(map[string]interface{})

		// Store code field if it exists
		if codeField, exists := inputs["code"]; exists {
			nodeData.Inputs.CodeRunner.(map[string]interface{})["code"] = codeField
		}

		// Check if there's a coderunner field with code inside
		if coderunner, exists := inputs["coderunner"]; exists {
			if coderunnerMap, ok := coderunner.(map[string]interface{}); ok {
				// Extract code from coderunner
				if code, codeExists := coderunnerMap["code"]; codeExists {
					nodeData.Inputs.CodeRunner.(map[string]interface{})["code"] = code
				}
				// Extract language from coderunner
				if lang, langExists := coderunnerMap["language"]; langExists {
					nodeData.Inputs.CodeRunner.(map[string]interface{})["language"] = lang
				}
			}
		}

		// Store language field if it exists
		if langField, exists := inputs["language"]; exists {
			nodeData.Inputs.CodeRunner.(map[string]interface{})["language"] = langField
		}

		// Store inputParameters for input parsing (check both cases)
		if inputParams, exists := inputs["inputParameters"]; exists {
			nodeData.Inputs.CodeRunner.(map[string]interface{})["inputParameters"] = inputParams
		} else if inputParams, exists := inputs["inputparameters"]; exists {
			nodeData.Inputs.CodeRunner.(map[string]interface{})["inputParameters"] = inputParams
		}

		// Store LLM parameters for LLM node parsing
		if llmParam, exists := inputs["llmparam"]; exists {
			nodeData.Inputs.LLMParam = llmParam
		}

		// Store all other relevant input fields for different node types
		nodeData.Inputs.InputParameters = parseInputParametersFromMap(inputs)

		// Type assertion for branches
		if branches, exists := inputs["branches"]; exists {
			if branchesArray, ok := branches.([]interface{}); ok {
				nodeData.Inputs.Branches = branchesArray
			}
		}

		// Store other fields as interface{}
		nodeData.Inputs.IntentDetector = inputs["intentdetector"]
		nodeData.Inputs.Selector = inputs["selector"]
	}

	return nodeData
}

// parseCozeNodeMeta parses node meta from block
func (p *IterationNodeParser) parseCozeNodeMeta(meta map[string]interface{}) CozeNodeMeta {
	nodeMeta := CozeNodeMeta{}

	if position, ok := meta["position"].(map[string]interface{}); ok {
		nodeMeta.Position = CozePosition{
			X: p.getFloatFromMap(position, "x", 0),
			Y: p.getFloatFromMap(position, "y", 0),
		}
	}

	return nodeMeta
}

// extractFieldFromInputs extracts field from block inputs structure
func (p *IterationNodeParser) extractFieldFromInputs(inputs interface{}, fieldName string) (interface{}, bool) {
	if inputsMap, ok := inputs.(map[string]interface{}); ok {
		if value, exists := inputsMap[fieldName]; exists {
			return value, true
		}
	}
	return nil, false
}

// convertIterationCodeFormat converts iteration code to iFlytek format
func (p *IterationNodeParser) convertIterationCodeFormat(code string, inputs []models.Input) (string, error) {
	if code == "" {
		return p.generateDefaultIterationCode(inputs), nil
	}

	// Use similar logic to main layer code parser
	// Step 1: Generate function signature dynamically from inputs
	signature := p.generateIterationFunctionSignature(inputs)

	// Step 2: Remove Coze-specific async/await and Args handling
	cleanedCode := p.removeCozeSpecificSyntaxImproved(code)

	// Step 3: Replace parameter access patterns for iteration context
	convertedCode := p.replaceIterationParameterAccessForSingleInput(cleanedCode, inputs)

	// Step 4: Combine signature with converted body
	return p.assembleIterationCodeImproved(signature, convertedCode), nil
}

// removeCozeSpecificSyntaxImproved removes Coze-specific syntax elements
func (p *IterationNodeParser) removeCozeSpecificSyntaxImproved(code string) string {
	lines := strings.Split(code, "\n")
	var cleanedLines []string

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip Coze-specific lines
		if p.isCozeSpecificLineImproved(trimmedLine) {
			continue
		}

		// Keep other lines
		cleanedLines = append(cleanedLines, line)
	}

	return strings.Join(cleanedLines, "\n")
}

// isCozeSpecificLineImproved identifies Coze-specific syntax lines
func (p *IterationNodeParser) isCozeSpecificLineImproved(line string) bool {
	cozePatterns := []string{
		"async def main(args)",
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

	// General pattern matching
	for _, pattern := range cozePatterns {
		if strings.Contains(line, pattern) {
			return true
		}
	}

	return false
}

// assembleIterationCodeImproved combines signature with body, moving imports to top.
func (p *IterationNodeParser) assembleIterationCodeImproved(signature, body string) string {
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

	// Ensure proper indentation for function body - use same logic as main code parser
	var indentedBodyLines []string
	for _, line := range bodyLines {
		if strings.TrimSpace(line) == "" {
			indentedBodyLines = append(indentedBodyLines, "")
		} else if !strings.HasPrefix(line, "    ") && strings.TrimSpace(line) != "" {
			// Add base 4-space indentation if not already present
			indentedBodyLines = append(indentedBodyLines, "    "+strings.TrimSpace(line))
		} else {
			// Keep existing indentation
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

// generateIterationFunctionSignature generates function signature for iteration code
func (p *IterationNodeParser) generateIterationFunctionSignature(inputs []models.Input) string {
	if len(inputs) == 0 {
		return "def main() -> dict:"
	}

	// Generate parameters for all inputs (not consolidating to single iteration variable)
	var params []string
	for _, input := range inputs {
		paramType := p.convertUnifiedDataTypeToString(input.Type)
		params = append(params, fmt.Sprintf("%s: %s", input.Name, paramType))
	}

	return fmt.Sprintf("def main(%s) -> dict:", strings.Join(params, ", "))
}

// convertUnifiedDataTypeToString converts unified data types to Python type annotations
func (p *IterationNodeParser) convertUnifiedDataTypeToString(dataType models.UnifiedDataType) string {
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
	return "str"
}

// replaceIterationParameterAccessForSingleInput replaces parameter access patterns and adds null handling
func (p *IterationNodeParser) replaceIterationParameterAccessForSingleInput(code string, inputs []models.Input) string {
	if len(inputs) == 0 {
		return code
	}

	convertedCode := code

	// Replace all parameter access patterns (params.get, args.params.get) with direct variable names
	// For each input parameter, replace its access patterns with the parameter name
	for _, input := range inputs {
		paramName := input.Name

		// Replace str(args.params.get('paramName', ...)) with paramName
		re := regexp.MustCompile(fmt.Sprintf(`str\((?:args\.)?params\.get\(['"]%s['"], [^)]*\)\)`, regexp.QuoteMeta(paramName)))
		convertedCode = re.ReplaceAllString(convertedCode, paramName)

		// Replace int(args.params.get('paramName', ...)) with paramName
		re = regexp.MustCompile(fmt.Sprintf(`int\((?:args\.)?params\.get\(['"]%s['"], [^)]*\)\)`, regexp.QuoteMeta(paramName)))
		convertedCode = re.ReplaceAllString(convertedCode, paramName)

		// Replace float(args.params.get('paramName', ...)) with paramName
		re = regexp.MustCompile(fmt.Sprintf(`float\((?:args\.)?params\.get\(['"]%s['"], [^)]*\)\)`, regexp.QuoteMeta(paramName)))
		convertedCode = re.ReplaceAllString(convertedCode, paramName)

		// Replace args.params.get('paramName', ...) with paramName
		re = regexp.MustCompile(fmt.Sprintf(`(?:args\.)?params\.get\(['"]%s['"], [^)]*\)`, regexp.QuoteMeta(paramName)))
		convertedCode = re.ReplaceAllString(convertedCode, paramName)

		// Replace args.params['paramName'] with paramName
		re = regexp.MustCompile(fmt.Sprintf(`(?:args\.)?params\[['"]%s['"]\]`, regexp.QuoteMeta(paramName)))
		convertedCode = re.ReplaceAllString(convertedCode, paramName)
	}

	// Add null handling code at the beginning of the function body
	// This ensures that parameters can handle None values
	var nullHandlingLines []string
	nullHandlingLines = append(nullHandlingLines, "    # 处理可能为None的参数，转换为空字符串")
	for _, input := range inputs {
		nullHandlingLines = append(nullHandlingLines, fmt.Sprintf("    %s = %s or \"\"", input.Name, input.Name))
	}
	nullHandlingLines = append(nullHandlingLines, "")

	// Insert null handling after any import statements and before the main function body
	lines := strings.Split(convertedCode, "\n")
	var resultLines []string
	insertedNullHandling := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Add line to results first
		resultLines = append(resultLines, line)

		// Insert null handling after the function definition line
		if !insertedNullHandling && strings.Contains(trimmedLine, "def main(") && strings.Contains(trimmedLine, ") -> dict:") {
			// Insert null handling lines
			for _, nullLine := range nullHandlingLines {
				resultLines = append(resultLines, nullLine)
			}
			insertedNullHandling = true
		}
	}

	return strings.Join(resultLines, "\n")
}

// generateDefaultIterationCode generates default code for iteration blocks
func (p *IterationNodeParser) generateDefaultIterationCode(inputs []models.Input) string {
	signature := p.generateIterationFunctionSignature(inputs)
	defaultBody := "    return {'result': 'No implementation provided'}"
	return signature + "\n" + defaultBody
}

// convertLanguageCode converts Coze language codes to standard names
func (p *IterationNodeParser) convertLanguageCode(langCode int) string {
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
	return "python3"
}

// parseOutputSelector parses output selector configuration
func (p *IterationNodeParser) parseOutputSelector(outputs []CozeOutput) models.OutputSelectorConfig {
	selector := models.OutputSelectorConfig{}

	// Always use standard "output" name for iFlytek compatibility
	// This ensures consistent output mapping regardless of Coze output names
	selector.NodeID = ""           // Will be set by the framework
	selector.OutputName = "output" // Force standard "output" name to match iFlytek template

	return selector
}

// parseIterationInputs processes iteration node inputs with correct type mapping
func (p *IterationNodeParser) parseIterationInputs(cozeNode CozeNode) []models.Input {
	var inputs []models.Input

	// Parse inputs from node inputs structure - check both formats
	var inputParams []CozeNodeInputParam
	if cozeNode.Data.Inputs != nil {
		if cozeNode.Data.Inputs.InputParameters != nil {
			inputParams = cozeNode.Data.Inputs.InputParameters
		} else if cozeNode.Data.Inputs.InputParametersAlt != nil {
			inputParams = cozeNode.Data.Inputs.InputParametersAlt
		}
	}

	if len(inputParams) > 0 {
		for _, param := range inputParams {
			input := models.Input{
				Name:        param.Name,
				Label:       param.Name,
				Type:        models.DataTypeArrayString, // iFlytek iteration inputs must be array-string type
				Required:    true,
				Description: "",
			}

			// Parse reference if it exists
			if param.Input.Value.Type == "ref" {
				input.Reference = &models.VariableReference{
					Type:       models.ReferenceTypeNodeOutput,
					NodeID:     param.Input.Value.Content.BlockID,
					OutputName: param.Input.Value.Content.Name,
					DataType:   input.Type,
				}
			}

			inputs = append(inputs, input)
		}
	}

	// iFlytek requires iteration nodes to have at least one input parameter
	// If no inputs are specified, create a default input parameter
	if len(inputs) == 0 {
		// Check if there's a source node that feeds into this iteration
		sourceNodeRef := p.findIterationInputSource(cozeNode)

		defaultInput := models.Input{
			Name:        "input",
			Label:       "input",
			Type:        models.DataTypeArrayString, // iFlytek iteration inputs must be array-string type
			Required:    true,
			Description: "Default iteration input parameter",
		}

		// Set reference to source node if found
		if sourceNodeRef != nil {
			defaultInput.Reference = sourceNodeRef
		}

		inputs = append(inputs, defaultInput)
	}

	return inputs
}

// findIterationInputSource finds the source node that provides input to this iteration
func (p *IterationNodeParser) findIterationInputSource(cozeNode CozeNode) *models.VariableReference {
	// Check for input parameters - support both formats
	var inputParams []CozeNodeInputParam
	if cozeNode.Data.Inputs != nil {
		if cozeNode.Data.Inputs.InputParameters != nil {
			inputParams = cozeNode.Data.Inputs.InputParameters
		} else if cozeNode.Data.Inputs.InputParametersAlt != nil {
			inputParams = cozeNode.Data.Inputs.InputParametersAlt
		}
	}

	// Look for variable references in input parameters
	for _, param := range inputParams {
		if param.Input.Value.Type == "ref" {
			return &models.VariableReference{
				Type:       models.ReferenceTypeNodeOutput,
				NodeID:     param.Input.Value.Content.BlockID,
				OutputName: param.Input.Value.Content.Name,
				DataType:   models.DataTypeArrayString,
			}
		}
	}

	// If no explicit input found, return nil
	return nil
}

// parseIterationOutputs processes iteration node outputs
func (p *IterationNodeParser) parseIterationOutputs(cozeNode CozeNode) []models.Output {
	var outputs []models.Output

	if cozeNode.Data.Outputs != nil {
		for range cozeNode.Data.Outputs {
			// Force all iteration outputs to use standard "output" name for iFlytek compatibility
			// This ensures the iteration main node can properly obtain iteration results
			outputs = append(outputs, models.Output{
				Name:        "output",                   // Always use "output" name to match iFlytek template
				Type:        models.DataTypeArrayString, // iFlytek iteration outputs use array-string type
				Description: "",
				Required:    true,
			})
		}
	}

	// Ensure at least one output exists with standard name to match iFlytek template
	if len(outputs) == 0 {
		outputs = append(outputs, models.Output{
			Name:        "output",                   // Use standard "output" name to match iFlytek template
			Type:        models.DataTypeArrayString, // iFlytek iteration outputs use array-string type
			Description: "Iteration result list",
			Required:    true,
		})
	}

	return outputs
}

// Helper methods
func (p *IterationNodeParser) getStringParam(params map[string]interface{}, key string, defaultValue string) string {
	if value, exists := params[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (p *IterationNodeParser) getStringFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if value, exists := m[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func (p *IterationNodeParser) getFloatFromMap(m map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := m[key]; exists {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
				return floatVal
			}
		}
	}
	return defaultValue
}

// mapIterationVariableReferenceWithOutput maps variable references within iteration context
// Returns both mapped node ID and output name
func (p *IterationNodeParser) mapIterationVariableReferenceWithOutput(blockID, outputName string, iterationID string) (string, string) {
	// Check if this is a reference to main workflow start node
	if p.isMainWorkflowStartNode(blockID) {
		// Map to the original iteration ID - let the iFlytek generator handle the final mapping
		return iterationID, outputName
	}

	// For other node types, return original values
	return blockID, outputName
}

// isMainWorkflowStartNode checks if the node ID corresponds to main workflow start node
func (p *IterationNodeParser) isMainWorkflowStartNode(nodeID string) bool {
	// Common patterns for main workflow start nodes in Coze
	commonStartNodeIDs := []string{"100001", "node-start", "start"}

	for _, startID := range commonStartNodeIDs {
		if nodeID == startID {
			return true
		}
	}

	return false
}

// registerIterationOutputMapping registers output name mappings for iteration nodes
// This ensures other nodes referencing iteration outputs use the correct standardized names
func (p *IterationNodeParser) registerIterationOutputMapping(cozeNode CozeNode, parsedNode *models.Node) {
	// If the original Coze node had different output names, register mappings
	if cozeNode.Data.Outputs != nil {
		for _, originalOutput := range cozeNode.Data.Outputs {
			// Register mapping from original Coze output name to standardized iFlytek name
			if originalOutput.Name != "output" {
				p.registerNodeOutputMapping(cozeNode.ID, originalOutput.Name, "output")
			}
		}
	}
}

// parseLLMBlock parses LLM nodes within iteration using main layer LLM parser
func (p *IterationNodeParser) parseLLMBlock(cozeNode CozeNode, iterationID string) (*models.Node, error) {
	// Convert iteration LLM configuration format to main layer format
	convertedNode, err := p.convertIterationLLMFormat(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to convert iteration LLM format: %w", err)
	}

	// Create main layer LLM parser instance
	llmParser := NewLLMNodeParser(p.variableRefSystem)

	// Use main layer parsing logic with converted node
	node, err := llmParser.ParseNode(convertedNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM block using main parser: %w", err)
	}

	// Set iteration-specific configuration
	p.setIterationNodeConfig(node, iterationID)

	// Process iteration-specific variable references
	p.processIterationVariableReferences(node, iterationID)

	return node, nil
}

// parseSelectorBlock parses selector/branch nodes within iteration using main layer selector parser
func (p *IterationNodeParser) parseSelectorBlock(cozeNode CozeNode, iterationID string) (*models.Node, error) {
	// Create main layer selector parser instance
	selectorParser := NewSelectorNodeParser(p.variableRefSystem)

	// Use main layer parsing logic
	node, err := selectorParser.ParseNode(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse selector block using main parser: %w", err)
	}

	// Set iteration-specific configuration
	p.setIterationNodeConfig(node, iterationID)

	// Process iteration-specific variable references
	p.processIterationVariableReferences(node, iterationID)

	return node, nil
}

// parseClassifierBlock parses classifier nodes within iteration using main layer classifier parser
func (p *IterationNodeParser) parseClassifierBlock(cozeNode CozeNode, iterationID string) (*models.Node, error) {
	// Create main layer classifier parser instance
	classifierParser := NewClassifierNodeParser(p.variableRefSystem)

	// Use main layer parsing logic
	node, err := classifierParser.ParseNode(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse classifier block using main parser: %w", err)
	}

	// Set iteration-specific configuration
	p.setIterationNodeConfig(node, iterationID)

	// Process iteration-specific variable references
	p.processIterationVariableReferences(node, iterationID)

	return node, nil
}

// processIterationVariableReferences processes variable references for iteration context
func (p *IterationNodeParser) processIterationVariableReferences(node *models.Node, iterationID string) {
	// Process input variable references
	for i := range node.Inputs {
		if node.Inputs[i].Reference != nil {
			// Map main workflow references to iteration context
			originalNodeID := node.Inputs[i].Reference.NodeID
			originalOutputName := node.Inputs[i].Reference.OutputName

			mappedNodeID, mappedOutputName := p.mapIterationVariableReferenceWithOutput(
				originalNodeID, originalOutputName, iterationID)

			node.Inputs[i].Reference.NodeID = mappedNodeID
			node.Inputs[i].Reference.OutputName = mappedOutputName
		}
	}

	// Additional reference processing can be added here for specific config types
	switch config := node.Config.(type) {
	case models.ClassifierConfig:
		// Process classifier query variable references if needed
		if config.QueryVariable != "" {
			// Could map query variable references here if needed
		}
	case models.ConditionConfig:
		// Process condition variable references in cases
		for i := range config.Cases {
			for j := range config.Cases[i].Conditions {
				// Process variable selectors in conditions
				if len(config.Cases[i].Conditions[j].VariableSelector) > 0 {

				}
			}
		}
	}
}

// registerNodeOutputMapping registers an output name mapping for a specific node
// This tells the variable reference system to map references from oldOutputName to newOutputName
func (p *IterationNodeParser) registerNodeOutputMapping(nodeID, oldOutputName, newOutputName string) {
	// Store mapping in BaseNodeParser's variable reference system
	// This will be used when parsing variable references in other nodes
	if p.variableRefSystem != nil {
		p.variableRefSystem.RegisterOutputMapping(nodeID, oldOutputName, newOutputName)
	}
}

// convertIterationLLMFormat converts iteration LLM node format to main layer format
func (p *IterationNodeParser) convertIterationLLMFormat(cozeNode CozeNode) (CozeNode, error) {
	// Make a copy of the original node
	convertedNode := cozeNode

	// Initialize inputs if nil
	if convertedNode.Data.Inputs == nil {
		convertedNode.Data.Inputs = &CozeNodeInputs{}
	}

	// Check if this LLM node uses array format for llmparam (iteration format)
	if convertedNode.Data.Inputs.LLMParam != nil {
		// Try to convert array format to object format
		if llmParamArray, ok := convertedNode.Data.Inputs.LLMParam.([]interface{}); ok {
			// Convert array format to object format expected by main layer parser
			llmParamObject := make(map[string]interface{})

			for _, paramItem := range llmParamArray {
				if paramMap, ok := paramItem.(map[string]interface{}); ok {
					paramName := p.getStringFromMap(paramMap, "name", "")
					if paramName != "" && paramMap["input"] != nil {
						if inputMap, ok := paramMap["input"].(map[string]interface{}); ok {
							// Extract the actual value from the input structure
							if valueMap, ok := inputMap["value"].(map[string]interface{}); ok {
								if content, exists := valueMap["content"]; exists {
									// Convert content based on type
									inputType := p.getStringFromMap(inputMap, "type", "string")
									switch inputType {
									case "integer":
										if contentStr, ok := content.(string); ok {
											if intVal, err := strconv.Atoi(contentStr); err == nil {
												llmParamObject[paramName] = intVal
											}
										} else if intVal, ok := content.(int); ok {
											llmParamObject[paramName] = intVal
										}
									case "float":
										if contentStr, ok := content.(string); ok {
											if floatVal, err := strconv.ParseFloat(contentStr, 64); err == nil {
												llmParamObject[paramName] = floatVal
											}
										} else if floatVal, ok := content.(float64); ok {
											llmParamObject[paramName] = floatVal
										}
									default: // string and others
										llmParamObject[paramName] = content
									}
								}
							} else {
								// Handle cases where value is directly under input
								if content, exists := inputMap["content"]; exists {
									llmParamObject[paramName] = content
								} else if valueStr, exists := inputMap["value"]; exists {
									// Handle direct value in input (not nested in value object)
									llmParamObject[paramName] = valueStr
								}
							}
						}
					}
				}
			}

			// Replace the array format with object format
			convertedNode.Data.Inputs.LLMParam = llmParamObject
		}
	}

	return convertedNode, nil
}

// parseInputParametersFromMap parses input parameters from inputs map
func parseInputParametersFromMap(inputs map[string]interface{}) []CozeNodeInputParam {
	var inputParams []CozeNodeInputParam

	// Check for inputParameters or inputparameters
	var paramArray []interface{}
	if params, exists := inputs["inputParameters"]; exists {
		if arr, ok := params.([]interface{}); ok {
			paramArray = arr
		}
	} else if params, exists := inputs["inputparameters"]; exists {
		if arr, ok := params.([]interface{}); ok {
			paramArray = arr
		}
	}

	// Convert to CozeNodeInputParam structs
	for _, param := range paramArray {
		if paramMap, ok := param.(map[string]interface{}); ok {
			inputParam := CozeNodeInputParam{}

			if name, ok := paramMap["name"].(string); ok {
				inputParam.Name = name
			}

			if input, ok := paramMap["input"].(map[string]interface{}); ok {
				inputParam.Input = CozeNodeInput{
					Type: getStringFromMapHelper(input, "Type", "string"),
				}

				// Parse Value
				if value, ok := input["Value"].(map[string]interface{}); ok {
					inputParam.Input.Value = CozeNodeInputValue{
						Type: getStringFromMapHelper(value, "type", ""),
					}

					// Parse Content if exists
					if content, ok := value["content"].(map[string]interface{}); ok {
						inputParam.Input.Value.Content = CozeNodeInputContent{
							BlockID: getStringFromMapHelper(content, "blockID", ""),
							Name:    getStringFromMapHelper(content, "name", ""),
							Source:  getStringFromMapHelper(content, "source", ""),
						}
					}

					// Parse RawMeta if exists
					if rawMeta, ok := value["rawmeta"].(map[string]interface{}); ok {
						if typeVal, ok := rawMeta["type"].(int); ok {
							inputParam.Input.Value.RawMeta = CozeNodeInputRawMeta{
								Type: typeVal,
							}
						}
					}
				}
			}

			inputParams = append(inputParams, inputParam)
		}
	}

	return inputParams
}

// Helper function for string extraction
func getStringFromMapHelper(m map[string]interface{}, key, defaultValue string) string {
	if value, exists := m[key]; exists {
		if strVal, ok := value.(string); ok {
			return strVal
		}
	}
	return defaultValue
}
