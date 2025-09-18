package parser

import (
	"ai-agents-transformer/internal/models"
)

// CodeNodeParser parses Dify code nodes.
type CodeNodeParser struct {
	*BaseNodeParser
}

func NewCodeNodeParser(variableRefSystem *models.VariableReferenceSystem) NodeParser {
	return &CodeNodeParser{
		BaseNodeParser: NewBaseNodeParser("code", variableRefSystem),
	}
}

// ParseNode parses code node.
func (p *CodeNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Create basic node
	node := p.parseBasicNodeInfo(difyNode)
	node.Type = models.NodeTypeCode

	// Parse code configuration
	codeConfig := models.CodeConfig{
		Language:      p.getCodeLanguage(difyNode.Data.CodeLanguage),
		Code:          difyNode.Data.Code,
		IsInIteration: difyNode.Data.IsInIteration,
		IterationID:   difyNode.Data.IterationID,
	}

	node.Config = codeConfig

	// Parse inputs (from variables field)
	node.Inputs = p.parseCodeInputs(difyNode.Data.Variables)

	// Parse outputs (from outputs field in object format)
	node.Outputs = p.parseCodeOutputs(difyNode.Data.Outputs)

	return node, nil
}

// getCodeLanguage gets code language.
func (p *CodeNodeParser) getCodeLanguage(language string) string {
	switch language {
	case "python3":
		return "python3"
	case "javascript":
		return "javascript"
	default:
		return "python3" // Default to python3
	}
}

// parseCodeInputs parses code node inputs.
func (p *CodeNodeParser) parseCodeInputs(variables []DifyVariable) []models.Input {
	var inputs []models.Input

	for _, variable := range variables {
		input := models.Input{
			Name:        variable.Variable,
			Type:        p.convertVariableType(variable.ValueType),
			Description: variable.Label,
			Required:    variable.Required,
		}

		// Parse variable selector reference
		if len(variable.ValueSelector) >= 2 {
			input.Reference = &models.VariableReference{
				Type:       models.ReferenceTypeNodeOutput,
				NodeID:     variable.ValueSelector[0],
				OutputName: variable.ValueSelector[1],
				DataType:   p.convertVariableType(variable.ValueType),
			}
		}

		inputs = append(inputs, input)
	}

	return inputs
}

// parseCodeOutputs parses code node outputs (object format).
func (p *CodeNodeParser) parseCodeOutputs(outputs interface{}) []models.Output {
	var result []models.Output

	// Handle different map types
	if stringMap := p.convertToStringMap(outputs); stringMap != nil {
		for outputName, value := range stringMap {
			if output := p.parseOutputEntry(outputName, value); output != nil {
				result = append(result, *output)
			}
		}
	}

	return result
}

// convertToStringMap converts various map types to map[string]interface{}
func (p *CodeNodeParser) convertToStringMap(data interface{}) map[string]interface{} {
	switch m := data.(type) {
	case map[string]interface{}:
		return m
	case map[interface{}]interface{}:
		return p.convertInterfaceMapToStringMap(m)
	default:
		return nil
	}
}

// convertInterfaceMapToStringMap converts map[interface{}]interface{} to map[string]interface{}
func (p *CodeNodeParser) convertInterfaceMapToStringMap(interfaceMap map[interface{}]interface{}) map[string]interface{} {
	stringMap := make(map[string]interface{})
	for k, v := range interfaceMap {
		if key, ok := k.(string); ok {
			stringMap[key] = v
		}
	}
	return stringMap
}

// parseOutputEntry parses a single output entry
func (p *CodeNodeParser) parseOutputEntry(outputName string, value interface{}) *models.Output {
	outputInfo := p.convertToStringMap(value)
	if outputInfo == nil {
		return nil
	}

	return &models.Output{
		Name:        outputName,
		Type:        p.convertOutputType(outputInfo),
		Description: "Code execution result",
	}
}

// convertVariableType converts variable type.
func (p *CodeNodeParser) convertVariableType(varType string) models.UnifiedDataType {
	switch varType {
	case "string", "text-input":
		return models.DataTypeString
	case "number":
		return models.DataTypeNumber
	case "boolean":
		return models.DataTypeBoolean
	default:
		return models.DataTypeString
	}
}

// convertOutputType converts output type.
func (p *CodeNodeParser) convertOutputType(outputInfo map[string]interface{}) models.UnifiedDataType {
	if typeStr, ok := outputInfo["type"].(string); ok {
		switch typeStr {
		case "string":
			return models.DataTypeString
		case "number":
			return models.DataTypeNumber
		case "boolean":
			return models.DataTypeBoolean
		case "array[string]":
			return models.DataTypeArrayString
		case "array[object]":
			return models.DataTypeArrayObject
		case "object":
			return models.DataTypeObject
		default:
			return models.DataTypeString
		}
	}
	return models.DataTypeString
}
