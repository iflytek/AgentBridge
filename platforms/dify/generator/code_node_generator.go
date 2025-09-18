package generator

import (
	"ai-agents-transformer/internal/models"
	"fmt"
)

// CodeNodeGenerator generates code nodes
type CodeNodeGenerator struct {
	*BaseNodeGenerator
	variableSelectorConverter *VariableSelectorConverter
}

func NewCodeNodeGenerator() *CodeNodeGenerator {
	return &CodeNodeGenerator{
		BaseNodeGenerator:         NewBaseNodeGenerator(models.NodeTypeCode),
		variableSelectorConverter: NewVariableSelectorConverter(),
	}
}

// GenerateNode generates a code node
func (g *CodeNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeCode {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeCode)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Set code node specific data
	g.setCodeNodeData(&difyNode, node)

	// Restore Dify-specific fields from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// SetNodeMapping sets node mapping for variable selector converter
func (g *CodeNodeGenerator) SetNodeMapping(nodes []models.Node) {
	g.variableSelectorConverter.SetNodeMapping(nodes)
}

// setCodeNodeData sets code node data
func (g *CodeNodeGenerator) setCodeNodeData(difyNode *DifyNode, node models.Node) {
	// Extract code content from unified DSL configuration
	var code string
	var language string = "python3" // Default language

	if codeConfig, ok := node.Config.(models.CodeConfig); ok {
		if codeConfig.Code != "" {
			code = codeConfig.Code
		}
		if codeConfig.Language != "" {
			language = codeConfig.Language
		}
	}

	// If no code content, generate default code
	if code == "" {
		code = g.generateDefaultCode(node)
	}

	// Set code node specific fields
	difyNode.Data.Code = code
	difyNode.Data.CodeLanguage = language
	difyNode.Data.Dependencies = ""

	// Set output configuration
	difyNode.Data.Outputs = g.generateCodeOutputs(node.Outputs)

	// Set variable configuration
	difyNode.Data.Variables = g.generateCodeVariables(node.Inputs)
}

// generateCodeVariables generates variable configuration for code nodes
func (g *CodeNodeGenerator) generateCodeVariables(inputs []models.Input) []map[string]interface{} {
	variables := make([]map[string]interface{}, 0, len(inputs))

	for i, input := range inputs {
		// Only add variables when there are valid variable references
		if input.Reference != nil && input.Reference.NodeID != "" {
			// Use variable selector converter to handle field mapping
			valueSelector, err := g.variableSelectorConverter.ConvertVariableReference(input.Reference)
			if err != nil {
				// Fallback to original logic if conversion fails
				sourceNodeID := input.Reference.NodeID
				sourceVariableName := input.Reference.OutputName
				if sourceVariableName == "" {
					sourceVariableName = input.Name
				}
				valueSelector = []string{sourceNodeID, sourceVariableName}
			}

			// Dynamically generate variable names, avoid hardcoding
			variableName := input.Name
			if variableName == "" {
				variableName = fmt.Sprintf("arg%d", i+1)
			}

			variable := map[string]interface{}{
				"variable":       variableName,
				"value_type":     g.mapToDifyStandardType(string(input.Type)),
				"value_selector": valueSelector,
			}
			variables = append(variables, variable)
		}
	}

	// If no valid input variables, return empty array, avoid hardcoded default values
	return variables
}


// generateDefaultCode generates default code
func (g *CodeNodeGenerator) generateDefaultCode(node models.Node) string {
	defaultCode := `def main(arg1: str) -> dict:
    """
    ` + node.Description + `
    
    Args:
        arg1 (str): Input parameter
    
    Returns:
        dict: Return result
    """
    # Write your code logic here
    result = {
        "output": f"Processing result: {arg1}"
    }
    
    return result`

	return defaultCode
}

// generateCodeOutputs generates code output configuration
func (g *CodeNodeGenerator) generateCodeOutputs(outputs []models.Output) map[string]interface{} {
	outputsConfig := make(map[string]interface{})

	for _, output := range outputs {
		// Strictly map to Dify standard types
		outputType := g.mapToDifyStandardType(string(output.Type))

		outputsConfig[output.Name] = map[string]interface{}{
			"type":        outputType,
			"description": output.Description,
		}
	}

	// If no outputs defined, add default output
	if len(outputs) == 0 {
		outputsConfig["output"] = map[string]interface{}{
			"type":        "string",
			"description": "Code execution result",
		}
	}

	return outputsConfig
}

// mapToDifyStandardType strictly maps to Dify standard types
func (g *CodeNodeGenerator) mapToDifyStandardType(inputType string) string {
	// Use unified mapping system, supports alias handling
	mapping := models.GetDefaultDataTypeMapping()
	return mapping.MapToDifyTypeWithAliases(inputType)
}

// restoreDifyPlatformConfig restores Dify platform-specific configuration
func (g *CodeNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	// Restore code configuration - directly set at data level
	if codeLanguage, exists := config["code_language"].(string); exists {
		node.Data.CodeLanguage = codeLanguage
	}

	if code, exists := config["code"].(string); exists {
		node.Data.Code = code
	}

	if outputs, exists := config["outputs"].(map[string]interface{}); exists {
		node.Data.Outputs = outputs
	}

	if dependencies, exists := config["dependencies"].(string); exists {
		node.Data.Dependencies = dependencies
	}

	// Restore variable configuration
	if variables, exists := config["variables"].([]interface{}); exists {
		variablesSlice := make([]map[string]interface{}, len(variables))
		for i, varInterface := range variables {
			if varMap, ok := varInterface.(map[string]interface{}); ok {
				variablesSlice[i] = varMap
			}
		}
		node.Data.Variables = variablesSlice
	}

	// Restore other node-specific configuration
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}
