package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// ConditionNodeParser parses conditional branch nodes.
type ConditionNodeParser struct {
	*BaseNodeParser
	typeProvider TypeProvider // Optional type provider for advanced type inference
}

func NewConditionNodeParser(variableRefSystem *models.VariableReferenceSystem) *ConditionNodeParser {
	return &ConditionNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
		typeProvider:   nil, // Maintains backward compatibility
	}
}

// NewConditionNodeParserWithTypeProvider creates a condition node parser with TypeProvider support.
func NewConditionNodeParserWithTypeProvider(variableRefSystem *models.VariableReferenceSystem, typeProvider TypeProvider) *ConditionNodeParser {
	return &ConditionNodeParser{
		BaseNodeParser: NewBaseNodeParser(variableRefSystem),
		typeProvider:   typeProvider,
	}
}

// GetSupportedType returns the supported node type.
func (p *ConditionNodeParser) GetSupportedType() string {
	return IFlytekNodeTypeCondition
}

// ValidateNode validates node data.
func (p *ConditionNodeParser) ValidateNode(iflytekNode IFlytekNode) error {
	if iflytekNode.ID == "" {
		return fmt.Errorf("node ID is empty")
	}

	if iflytekNode.Type != p.GetSupportedType() {
		return fmt.Errorf("invalid node type: expected %s, got %s", p.GetSupportedType(), iflytekNode.Type)
	}

	return nil
}

// ParseNode parses a node.
func (p *ConditionNodeParser) ParseNode(iflytekNode IFlytekNode) (*models.Node, error) {
	if err := p.ValidateNode(iflytekNode); err != nil {
		return nil, err
	}

	// Parse basic information
	node := p.ParseBasicNodeInfo(iflytekNode, models.NodeTypeCondition)

	// Parse inputs and outputs
	if err := p.parseInputsOutputs(node, iflytekNode.Data); err != nil {
		return nil, err
	}

	// Parse configuration
	config, err := p.parseConditionConfig(iflytekNode.Data, iflytekNode.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse condition config: %w", err)
	}
	node.Config = config

	// Save platform-specific configuration
	p.SavePlatformConfig(node, iflytekNode)

	return node, nil
}

// parseInputsOutputs parses inputs and outputs.
func (p *ConditionNodeParser) parseInputsOutputs(node *models.Node, data map[string]interface{}) error {
	// Parse inputs
	if inputs, ok := data["inputs"].([]interface{}); ok {
		nodeInputs, err := p.ParseNodeInputs(inputs)
		if err != nil {
			return fmt.Errorf("failed to parse condition node inputs: %w", err)
		}
		node.Inputs = nodeInputs
	}

	// Conditional branch nodes usually have no outputs, but may have conditional branch output ports
	if outputs, ok := data["outputs"].([]interface{}); ok {
		nodeOutputs, err := p.ParseNodeOutputs(outputs)
		if err != nil {
			return fmt.Errorf("failed to parse condition node outputs: %w", err)
		}
		node.Outputs = nodeOutputs
	}

	return nil
}

// parseConditionConfig parses conditional branch configuration.
func (p *ConditionNodeParser) parseConditionConfig(data map[string]interface{}, nodeData map[string]interface{}) (*models.ConditionConfig, error) {
	config := &models.ConditionConfig{
		Cases: make([]models.ConditionCase, 0),
	}

	// Get conditional configuration from nodeParam
	if nodeParam, ok := data["nodeParam"].(map[string]interface{}); ok {
		// Parse cases
		if cases, ok := nodeParam["cases"].([]interface{}); ok {
			for _, caseData := range cases {
				if caseMap, ok := caseData.(map[string]interface{}); ok {
					conditionCase, err := p.parseConditionCase(caseMap, data) // Pass original data instead of inputs
					if err != nil {
						return nil, fmt.Errorf("failed to parse condition case: %w", err)
					}

					// Set branch level information
					if levelRaw, exists := caseMap["level"]; exists {
						if level, ok := levelRaw.(float64); ok {
							conditionCase.Level = int(level)
							// Default branch (level 999) is set as default case
							if level == 999 {
								config.DefaultCase = conditionCase.CaseID
							}
						} else if level, ok := levelRaw.(int); ok {
							conditionCase.Level = level
							// Default branch (level 999) is set as default case
							if level == 999 {
								config.DefaultCase = conditionCase.CaseID
							}
						}
					}

					config.Cases = append(config.Cases, *conditionCase)
				}
			}
		}
	}

	return config, nil
}

// parseConditionCase parses a single conditional branch.
func (p *ConditionNodeParser) parseConditionCase(caseData map[string]interface{}, nodeData map[string]interface{}) (*models.ConditionCase, error) {
	conditionCase := &models.ConditionCase{
		Conditions: make([]models.Condition, 0),
	}

	// Parse branch ID
	if id, ok := caseData["id"].(string); ok {
		conditionCase.CaseID = id
	}

	// Parse logical operator
	if logicalOperator, ok := caseData["logicalOperator"].(string); ok {
		conditionCase.LogicalOperator = logicalOperator
	}

	// Parse condition list
	if conditions, ok := caseData["conditions"].([]interface{}); ok {
		for _, condData := range conditions {
			if condMap, ok := condData.(map[string]interface{}); ok {
				condition, err := p.parseCondition(condMap, nodeData)
				if err != nil {
					return nil, fmt.Errorf("failed to parse condition: %w", err)
				}
				conditionCase.Conditions = append(conditionCase.Conditions, *condition)
			}
		}
	}

	return conditionCase, nil
}

// parseCondition parses a single condition.
func (p *ConditionNodeParser) parseCondition(condData map[string]interface{}, nodeData map[string]interface{}) (*models.Condition, error) {
	condition := &models.Condition{}

	// Parse comparison operator
	if operator, ok := condData["compareOperator"].(string); ok {
		condition.ComparisonOperator = p.mapComparisonOperator(operator)
	}

	// Parse left variable index (convert to variable selector)
	if leftVarIndex, ok := condData["leftVarIndex"].(string); ok {
		// Find corresponding variable selector based on left variable index
		variableSelector := p.getVariableSelectorByID(leftVarIndex, nodeData)
		condition.VariableSelector = variableSelector
	}

	// Parse right variable value (comparison value)
	if rightVarIndex, ok := condData["rightVarIndex"].(string); ok {
		// Get actual comparison value from node's original data
		actualValue := p.getInputValueByID(rightVarIndex, nodeData)
		if actualValue != "" {
			condition.Value = actualValue
		} else {
			// If corresponding input variable is not found, use original rightVarIndex
			condition.Value = rightVarIndex
		}
	}

	// Set variable type based on input variable type
	if leftVarIndex, ok := condData["leftVarIndex"].(string); ok {
		condition.VarType = p.getVariableTypeByID(leftVarIndex, nodeData)
	} else {
		condition.VarType = models.DataTypeString
	}

	return condition, nil
}

// getInputValueByID gets actual value from original data based on input variable ID.
func (p *ConditionNodeParser) getInputValueByID(varID string, nodeData map[string]interface{}) string {
	// Get input variable configuration from node's original data
	if inputs, ok := nodeData["inputs"].([]interface{}); ok {
		for _, inputInterface := range inputs {
			if inputMap, ok := inputInterface.(map[string]interface{}); ok {
				// Check if input variable ID matches
				if id, exists := inputMap["id"].(string); exists && id == varID {
					// Get actual value from schema.value.content
					if schema, exists := inputMap["schema"].(map[string]interface{}); exists {
						if value, exists := schema["value"].(map[string]interface{}); exists {
							if content, exists := value["content"].(string); exists {
								return content
							}
						}
					}
				}
			}
		}
	}
	return ""
}

// getVariableSelectorByID gets variable selector based on input variable ID.
func (p *ConditionNodeParser) getVariableSelectorByID(varID string, nodeData map[string]interface{}) []string {
	inputs, ok := nodeData["inputs"].([]interface{})
	if !ok {
		return []string{varID}
	}

	for _, inputInterface := range inputs {
		if selector := p.processInputForVariableSelector(inputInterface, varID); selector != nil {
			return selector
		}
	}

	// If no reference is found, return variable ID itself
	return []string{varID}
}

// processInputForVariableSelector processes a single input for variable selector extraction
func (p *ConditionNodeParser) processInputForVariableSelector(inputInterface interface{}, varID string) []string {
	inputMap, ok := inputInterface.(map[string]interface{})
	if !ok {
		return nil
	}

	id, exists := inputMap["id"].(string)
	if !exists || id != varID {
		return nil
	}

	return p.extractVariableSelectorFromSchema(inputMap)
}

// extractVariableSelectorFromSchema extracts variable selector from input schema
func (p *ConditionNodeParser) extractVariableSelectorFromSchema(inputMap map[string]interface{}) []string {
	schema, ok := inputMap["schema"].(map[string]interface{})
	if !ok {
		return nil
	}

	value, ok := schema["value"].(map[string]interface{})
	if !ok {
		return nil
	}

	return p.extractNodeOutputReference(value)
}

// extractNodeOutputReference extracts node and output reference from value
func (p *ConditionNodeParser) extractNodeOutputReference(value map[string]interface{}) []string {
	valueType, ok := value["type"].(string)
	if !ok || valueType != "ref" {
		return nil
	}

	content, ok := value["content"].(map[string]interface{})
	if !ok {
		return nil
	}

	nodeID, hasNodeID := content["nodeId"].(string)
	outputName, hasOutputName := content["name"].(string)

	if hasNodeID && hasOutputName {
		return []string{nodeID, outputName}
	}

	return nil
}

// getVariableTypeByID gets variable type based on input variable ID.
func (p *ConditionNodeParser) getVariableTypeByID(varID string, nodeData map[string]interface{}) models.UnifiedDataType {
	// Get the variable selector to identify which node output this refers to
	variableSelector := p.getVariableSelectorByID(varID, nodeData)
	if len(variableSelector) >= 2 {
		sourceNodeID := variableSelector[0]
		outputName := variableSelector[1]

		// Try to get the type from the already parsed nodes in the current parsing context
		// This is more direct than using the variable reference system
		if outputType := p.getOutputTypeFromContext(sourceNodeID, outputName); outputType != models.DataTypeString {
			return outputType
		}
	}

	// Fallback: Get input variable configuration from node's original data
	if inputs, ok := nodeData["inputs"].([]interface{}); ok {
		for _, inputInterface := range inputs {
			if inputMap, ok := inputInterface.(map[string]interface{}); ok {
				// Check if input variable ID matches
				if id, exists := inputMap["id"].(string); exists && id == varID {
					// Get variable type from schema.type
					if schema, exists := inputMap["schema"].(map[string]interface{}); exists {
						if schemaType, exists := schema["type"].(string); exists {
							// Use unified mapping system, supports alias processing
							mapping := models.GetDefaultDataTypeMapping()
							return mapping.FromIFlytekType(schemaType)
						}
					}
				}
			}
		}
	}
	// Default to string type
	return models.DataTypeString
}

// getOutputTypeFromContext tries to get output type from parsing context
func (p *ConditionNodeParser) getOutputTypeFromContext(sourceNodeID, outputName string) models.UnifiedDataType {
	// Uses TypeProvider if available, otherwise returns default value
	if p.typeProvider != nil {
		return p.typeProvider.GetOutputType(sourceNodeID, outputName)
	}

	// Returns string as safe default when TypeProvider is unavailable
	return models.DataTypeString
}

// mapComparisonOperator maps comparison operators.
func (p *ConditionNodeParser) mapComparisonOperator(operator string) string {
	// Use operator mapping table for better maintainability
	operatorMap := p.getOperatorMappingTable()
	if mappedOp, exists := operatorMap[operator]; exists {
		return mappedOp
	}
	// Keep original operator if no mapping found
	return operator
}

// getOperatorMappingTable returns the operator mapping table
func (p *ConditionNodeParser) getOperatorMappingTable() map[string]string {
	return map[string]string{
		"contains":      "contains",
		"not_contains":  "not_contains",
		"equals":        "equals",
		"==":            "equals",
		"not_equals":    "not_equals",
		"!=":            "not_equals",
		"start_with":    "start_with",
		"starts_with":   "start_with",
		"end_with":      "end_with",
		"ends_with":     "end_with",
		"empty":         "is_empty",
		"is_empty":      "is_empty",
		"not_empty":     "is_not_empty",
		"is_not_empty":  "is_not_empty",
		"greater_than":  "greater_than",
		">":             "greater_than",
		"less_than":     "less_than",
		"<":             "less_than",
		"greater_equal": "greater_equal",
		">=":            "greater_equal",
		"less_equal":    "less_equal",
		"<=":            "less_equal",
	}
}
