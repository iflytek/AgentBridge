package parser

import (
	"ai-agents-transformer/internal/models"
	"fmt"
	"strconv"
)

// SelectorNodeParser parses Coze selector nodes.
type SelectorNodeParser struct {
	*BaseNodeParser
}

func NewSelectorNodeParser(variableRefSystem *models.VariableReferenceSystem) *SelectorNodeParser {
	return &SelectorNodeParser{
		BaseNodeParser: NewBaseNodeParser("8", variableRefSystem),
	}
}

// GetSupportedType returns the supported node type.
func (p *SelectorNodeParser) GetSupportedType() string {
	return "8"
}

// ParseNode parses a Coze selector node into unified DSL.
func (p *SelectorNodeParser) ParseNode(cozeNode CozeNode) (*models.Node, error) {
	if err := p.ValidateNode(cozeNode); err != nil {
		return nil, fmt.Errorf("selector node validation failed: %w", err)
	}

	node := p.parseBasicNodeInfo(cozeNode)
	node.Type = models.NodeTypeCondition

	// Parse selector configuration
	config, err := p.parseSelectorConfig(cozeNode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse selector config: %w", err)
	}
	node.Config = config

	// Parse inputs (empty for selector nodes)
	node.Inputs = []models.Input{}

	// Parse outputs (empty for selector nodes)
	node.Outputs = []models.Output{}

	return node, nil
}

// parseSelectorConfig parses the selector configuration from branches data.
func (p *SelectorNodeParser) parseSelectorConfig(cozeNode CozeNode) (models.ConditionConfig, error) {
	config := models.ConditionConfig{
		Cases: []models.ConditionCase{},
	}

	// Access branches from the appropriate location
	branches, err := p.extractBranches(cozeNode)
	if err != nil {
		return config, err
	}

	// Parse each branch as a condition case
	for i, branch := range branches {
		conditionCase, err := p.parseConditionCase(branch, i)
		if err != nil {
			return config, fmt.Errorf("failed to parse branch %d: %w", i, err)
		}
		config.Cases = append(config.Cases, conditionCase)
	}

	return config, nil
}

// extractBranches extracts branches from either schema or data location.
func (p *SelectorNodeParser) extractBranches(cozeNode CozeNode) ([]map[string]interface{}, error) {
	// Try direct branches in inputs first
	if cozeNode.Data.Inputs != nil && len(cozeNode.Data.Inputs.Branches) > 0 {
		return p.convertBranchesInterface(cozeNode.Data.Inputs.Branches)
	}

	// Try selector.branches structure
	if cozeNode.Data.Inputs != nil && cozeNode.Data.Inputs.Selector != nil {
		if selectorMap, ok := cozeNode.Data.Inputs.Selector.(map[string]interface{}); ok {
			if branches, ok := selectorMap["branches"].([]interface{}); ok {
				return p.convertBranchesInterface(branches)
			}
		}
	}

	return []map[string]interface{}{}, nil
}

// convertBranchesInterface converts []interface{} to []map[string]interface{}.
func (p *SelectorNodeParser) convertBranchesInterface(branches []interface{}) ([]map[string]interface{}, error) {
	var result []map[string]interface{}
	for _, branch := range branches {
		if branchMap, ok := branch.(map[string]interface{}); ok {
			result = append(result, branchMap)
		}
	}
	return result, nil
}

// parseConditionCase parses a single branch into a condition case.
func (p *SelectorNodeParser) parseConditionCase(branch map[string]interface{}, index int) (models.ConditionCase, error) {
	conditionCase := models.ConditionCase{
		CaseID:     fmt.Sprintf("case_%d", index),
		Conditions: []models.Condition{},
	}

	// Extract condition data
	conditionData, ok := branch["condition"].(map[string]interface{})
	if !ok {
		return conditionCase, fmt.Errorf("missing condition in branch")
	}

	// Parse logical operator
	conditionCase.LogicalOperator = p.parseLogicalOperator(conditionData)

	// Parse conditions array
	conditions, ok := conditionData["conditions"].([]interface{})
	if !ok {
		return conditionCase, fmt.Errorf("missing conditions array in branch")
	}

	// Parse each condition
	for _, conditionItem := range conditions {
		if conditionMap, ok := conditionItem.(map[string]interface{}); ok {
			condition, err := p.parseCondition(conditionMap)
			if err != nil {
				return conditionCase, fmt.Errorf("failed to parse condition: %w", err)
			}
			conditionCase.Conditions = append(conditionCase.Conditions, condition)
		}
	}

	return conditionCase, nil
}

// parseLogicalOperator parses logical operator from condition data.
func (p *SelectorNodeParser) parseLogicalOperator(conditionData map[string]interface{}) string {
	if logic, ok := conditionData["logic"].(float64); ok {
		return p.mapCozeLogicToOperator(int(logic))
	}
	if logic, ok := conditionData["logic"].(int); ok {
		return p.mapCozeLogicToOperator(logic)
	}
	return "and" // Default
}

// mapCozeLogicToOperator maps Coze logic values to operator strings.
func (p *SelectorNodeParser) mapCozeLogicToOperator(logic int) string {
	switch logic {
	case 1:
		return "or"
	case 2:
		return "and"
	default:
		return "and"
	}
}

// parseCondition parses a single condition from Coze format.
func (p *SelectorNodeParser) parseCondition(conditionMap map[string]interface{}) (models.Condition, error) {
	condition := models.Condition{}

	// Parse operator
	if operator, ok := conditionMap["operator"].(float64); ok {
		condition.ComparisonOperator = p.mapCozeOperatorToUnified(int(operator))
	} else if operator, ok := conditionMap["operator"].(int); ok {
		condition.ComparisonOperator = p.mapCozeOperatorToUnified(operator)
	} else {
		return condition, fmt.Errorf("missing or invalid operator in condition")
	}

	// Parse left operand (variable reference)
	if left, ok := conditionMap["left"].(map[string]interface{}); ok {
		variableSelector, err := p.parseVariableReference(left)
		if err != nil {
			return condition, fmt.Errorf("failed to parse left operand: %w", err)
		}
		condition.VariableSelector = variableSelector
	}

	// Parse right operand (literal value)
	if right, ok := conditionMap["right"].(map[string]interface{}); ok {
		value, err := p.parseLiteralValue(right)
		if err != nil {
			return condition, fmt.Errorf("failed to parse right operand: %w", err)
		}
		condition.Value = value
	}

	return condition, nil
}

// parseVariableReference parses a variable reference from Coze format.
func (p *SelectorNodeParser) parseVariableReference(left map[string]interface{}) ([]string, error) {
	if input, ok := left["input"].(map[string]interface{}); ok {
		// Try both "Value" (uppercase) and "value" (lowercase) for compatibility
		var value map[string]interface{}
		var found bool

		if v, ok := input["Value"].(map[string]interface{}); ok {
			value = v
			found = true
		} else if v, ok := input["value"].(map[string]interface{}); ok {
			value = v
			found = true
		}

		if found {
			// Check if type is directly in value
			if valueType, ok := value["type"].(string); ok && valueType == "ref" {
				if content, ok := value["content"].(map[string]interface{}); ok {
					blockID, _ := content["blockID"].(string)
					name, _ := content["name"].(string)
					if blockID != "" && name != "" {
						return []string{blockID, name}, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid variable reference format")
}

// parseLiteralValue parses a literal value from Coze format.
func (p *SelectorNodeParser) parseLiteralValue(right map[string]interface{}) (interface{}, error) {
	if input, ok := right["input"].(map[string]interface{}); ok {
		// Try both "Value" (uppercase) and "value" (lowercase) for compatibility
		var value map[string]interface{}
		var found bool

		if v, ok := input["Value"].(map[string]interface{}); ok {
			value = v
			found = true
		} else if v, ok := input["value"].(map[string]interface{}); ok {
			value = v
			found = true
		}

		if found {
			// Check if type is directly in value
			if valueType, ok := value["type"].(string); ok && valueType == "literal" {
				if content, exists := value["content"]; exists {
					return p.convertLiteralContent(content)
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid literal value format")
}

// convertLiteralContent converts literal content to appropriate type.
func (p *SelectorNodeParser) convertLiteralContent(content interface{}) (interface{}, error) {
	switch v := content.(type) {
	case string:
		// Try to parse as number if it looks like one
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal, nil
		}
		if floatVal, err := strconv.ParseFloat(v, 64); err == nil {
			return floatVal, nil
		}
		return v, nil
	case float64, int, bool:
		return v, nil
	default:
		return content, nil
	}
}

// mapCozeOperatorToUnified maps Coze operators to unified DSL operators.
func (p *SelectorNodeParser) mapCozeOperatorToUnified(operator int) string {
	switch operator {
	case 1: // Equal
		return "equals"
	case 2: // Not equal
		return "not equals"
	case 3: // Contains
		return "contains"
	case 4: // Not contains
		return "not contains"
	case 5: // Empty
		return "is_empty"
	case 6: // Greater than or equal
		return "gte"
	case 7: // Greater than
		return "gt"
	case 8: // Less than
		return "lt"
	case 9: // Less than or equal
		return "lte"
	case 10: // Not empty
		return "is_not_empty"
	case 11: // Is true
		return "is_true"
	case 12: // Is false
		return "is_false"
	case 13: // Length greater than
		return "length_gt"
	case 14: // Greater than or equal (used for numeric/date comparisons)
		return "gte"
	case 15: // Length less than
		return "length_lt"
	case 16: // Length less than or equal
		return "length_lte"
	case 17: // Contain key
		return "contain_key"
	case 18: // Not contain key
		return "not_contain_key"
	default:
		return "equals" // Default fallback
	}
}
