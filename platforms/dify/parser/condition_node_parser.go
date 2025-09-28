package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
)

// ConditionNodeParser parses Dify conditional branch nodes.
type ConditionNodeParser struct {
	*BaseNodeParser
}

func NewConditionNodeParser(variableRefSystem *models.VariableReferenceSystem) NodeParser {
	return &ConditionNodeParser{
		BaseNodeParser: NewBaseNodeParser("if-else", variableRefSystem),
	}
}

// ParseNode parses conditional branch node.
func (p *ConditionNodeParser) ParseNode(difyNode DifyNode) (*models.Node, error) {
	// Create basic node
	node := p.parseBasicNodeInfo(difyNode)
	node.Type = models.NodeTypeCondition

	// Parse conditional branch configuration
	conditionConfig := models.ConditionConfig{}

	// Parse cases
	if difyNode.Data.Cases != nil {
		conditionConfig.Cases = p.parseCases(difyNode.Data.Cases)
	}

	// Set default branch (Dify if-else structure has a default false branch)
	conditionConfig.DefaultCase = "false"

	node.Config = conditionConfig

	// Parse inputs
	node.Inputs = p.parseInputsFromConditions(difyNode.Data.Cases, difyNode.ID)

	// Conditional branch nodes have no outputs (branching is implemented through edge conditions)
	node.Outputs = []models.Output{}

	return node, nil
}

// parseCases parses conditional branch cases.
func (p *ConditionNodeParser) parseCases(cases []DifyCase) []models.ConditionCase {
	var conditionCases []models.ConditionCase

	for _, difyCase := range cases {
		conditionCase := models.ConditionCase{
			CaseID:          difyCase.CaseID,
			LogicalOperator: p.mapLogicalOperator(difyCase.LogicalOperator),
			Conditions:      p.parseConditions(difyCase.Conditions),
		}

		conditionCases = append(conditionCases, conditionCase)
	}

	return conditionCases
}

// parseConditions parses condition list.
func (p *ConditionNodeParser) parseConditions(difyConditions []DifyCondition) []models.Condition {
	var conditions []models.Condition

	for _, difyCondition := range difyConditions {
		condition := models.Condition{
			VariableSelector:   difyCondition.VariableSelector,
			ComparisonOperator: p.mapComparisonOperator(difyCondition.ComparisonOperator),
			Value:              difyCondition.Value,
			VarType:            p.mapVarType(difyCondition.VarType),
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

// parseInputsFromConditions parses input parameters from conditions.
func (p *ConditionNodeParser) parseInputsFromConditions(cases []DifyCase, nodeID string) []models.Input {
	var inputs []models.Input
	inputMap := make(map[string]bool) // For deduplication

	for _, difyCase := range cases {
		for _, condition := range difyCase.Conditions {
			if len(condition.VariableSelector) >= 2 {
				sourceNodeID := condition.VariableSelector[0]
				sourceOutput := condition.VariableSelector[1]

				// Build unique key for deduplication
				inputKey := fmt.Sprintf("%s.%s", sourceNodeID, sourceOutput)
				if inputMap[inputKey] {
					continue // Skip duplicate inputs
				}
				inputMap[inputKey] = true

				input := models.Input{
					Name:        p.generateInputName(sourceNodeID, sourceOutput),
					Description: fmt.Sprintf("Condition input from node %s output %s", sourceNodeID, sourceOutput),
					Type:        p.mapVarType(condition.VarType),
					Required:    false,
					Reference: &models.VariableReference{
						Type:       models.ReferenceTypeNodeOutput,
						NodeID:     sourceNodeID,
						OutputName: sourceOutput,
						DataType:   p.mapVarType(condition.VarType),
					},
				}

				inputs = append(inputs, input)
			}
		}
	}

	return inputs
}

// generateInputName generates input parameter name.
func (p *ConditionNodeParser) generateInputName(nodeID, outputName string) string {
	return fmt.Sprintf("condition_input_%s_%s", nodeID, outputName)
}

// mapLogicalOperator maps logical operators.
func (p *ConditionNodeParser) mapLogicalOperator(operator string) string {
	switch operator {
	case "and", "AND":
		return "and"
	case "or", "OR":
		return "or"
	default:
		return "and" // Default to and
	}
}

// mapComparisonOperator maps comparison operators.
func (p *ConditionNodeParser) mapComparisonOperator(operator string) string {
	return p.getOperatorMapping(operator)
}

// getOperatorMapping returns the mapped operator using lookup table
func (p *ConditionNodeParser) getOperatorMapping(operator string) string {
	mapping := p.getOperatorMappingTable()
	if mapped, exists := mapping[operator]; exists {
		return mapped
	}
	return operator // Keep as is if not found
}

// getOperatorMappingTable returns the operator mapping table
func (p *ConditionNodeParser) getOperatorMappingTable() map[string]string {
	return map[string]string{
		"contains":              "contains",
		"not contains":          "not_contains",
		"start with":            "starts_with",
		"end with":              "ends_with",
		"is":                    "equals",
		"is not":                "not_equals",
		"empty":                 "is_empty",
		"not empty":             "is_not_empty",
		"greater than":          "gt",
		"gt":                    "gt",
		"greater than or equal": "gte",
		"gte":                   "gte",
		"less than":             "lt",
		"lt":                    "lt",
		"less than or equal":    "lte",
		"lte":                   "lte",
	}
}

// mapVarType maps variable types.
func (p *ConditionNodeParser) mapVarType(varType string) models.UnifiedDataType {
	switch varType {
	case "string":
		return models.DataTypeString
	case "number":
		return models.DataTypeNumber
	case "boolean":
		return models.DataTypeBoolean
	case "array", "array[string]":
		return models.DataTypeArrayString
	case "array[object]":
		return models.DataTypeArrayObject
	case "object":
		return models.DataTypeObject
	default:
		return models.DataTypeString // Default to string
	}
}
