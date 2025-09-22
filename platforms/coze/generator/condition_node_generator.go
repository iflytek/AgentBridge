package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"fmt"
)

// ConditionNodeGenerator generates Coze condition nodes (selectors)
type ConditionNodeGenerator struct {
	idGenerator *CozeIDGenerator
}

// NewConditionNodeGenerator creates a condition node generator
func NewConditionNodeGenerator() *ConditionNodeGenerator {
	return &ConditionNodeGenerator{
		idGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator sets the shared ID generator
func (g *ConditionNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// GetNodeType returns the node type this generator handles
func (g *ConditionNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeCondition
}

// ValidateNode validates the unified node before generation
func (g *ConditionNodeGenerator) ValidateNode(unifiedNode *models.Node) error {
	if unifiedNode.Type != models.NodeTypeCondition {
		return fmt.Errorf("expected condition node, got %s", unifiedNode.Type)
	}

	conditionConfig, ok := common.AsConditionConfig(unifiedNode.Config)
	if !ok || conditionConfig == nil {
		return fmt.Errorf("invalid condition node config type: %T", unifiedNode.Config)
	}

	if len(conditionConfig.Cases) == 0 {
		return fmt.Errorf("condition node must have at least one case")
	}

	return nil
}

// GenerateNode generates a Coze workflow condition node
func (g *ConditionNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Create condition node data structure
	conditionInputs := g.generateConditionInputs(unifiedNode)

	nodeData := &CozeNodeData{
		Meta: &CozeNodeMetaInfo{
			Title:       unifiedNode.Title,
			Description: "连接多个下游分支，若设定的条件成立则仅运行对应的分支，若均不成立则只运行\"否则\"分支",
			Icon:        "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Condition-v2.jpg",
			Subtitle:    "选择器",
			MainColor:   "#00B2B2",
		},
		Outputs: []CozeNodeOutput{},
		Inputs:  conditionInputs,
		Size:    nil,
	}

	// Generate node position
	position := &CozePosition{
		X: unifiedNode.Position.X,
		Y: unifiedNode.Position.Y,
	}

	return &CozeNode{
		ID:      cozeNodeID,
		Type:    "8", // Coze condition node type
		Meta:    &CozeNodeMeta{Position: position},
		Data:    nodeData,
		Blocks:  []interface{}{},
		Edges:   []interface{}{},
		Version: "",
		Size:    nil,
	}, nil
}

// GenerateSchemaNode generates a Coze schema condition node
func (g *ConditionNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Generate branches from condition cases
	branches, err := g.generateBranches(unifiedNode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate branches: %w", err)
	}

	schemaInputs := map[string]interface{}{
		"branches": branches,
	}

	return &CozeSchemaNode{
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Title:       unifiedNode.Title,
				Description: "连接多个下游分支，若设定的条件成立则仅运行对应的分支，若均不成立则只运行\"否则\"分支",
				Icon:        "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Condition-v2.jpg",
				SubTitle:    "选择器",
				MainColor:   "#00B2B2",
			},
			Inputs: schemaInputs,
		},
		ID:   cozeNodeID,
		Type: "8",
		Meta: &CozeNodeMeta{
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
	}, nil
}

// generateConditionInputs generates condition inputs for nodes section
func (g *ConditionNodeGenerator) generateConditionInputs(unifiedNode *models.Node) map[string]interface{} {
	conditionConfig, _ := common.AsConditionConfig(unifiedNode.Config)

	// Generate selector branches, excluding empty condition branches (default cases with level=999)
	branches := make([]map[string]interface{}, 0)
	for _, caseItem := range conditionConfig.Cases {
		// Skip empty condition branches (typically default cases with level=999)
		if len(caseItem.Conditions) == 0 {
			continue
		}

		branch := g.generateSelectorBranch(caseItem, unifiedNode)
		branches = append(branches, branch)
	}

	return map[string]interface{}{
		"inputparameters": []interface{}{},
		"settingonerror":  nil,
		"nodebatchinfo":   nil,
		"llmparam":        nil,
		"outputemitter":   nil,
		"exit":            nil,
		"llm":             nil,
		"loop":            nil,
		"selector": map[string]interface{}{
			"branches": branches,
		},
		"textprocessor":      nil,
		"subworkflow":        nil,
		"intentdetector":     nil,
		"databasenode":       nil,
		"httprequestnode":    nil,
		"knowledge":          nil,
		"coderunner":         nil,
		"pluginapiparam":     nil,
		"variableaggregator": nil,
		"variableassigner":   nil,
		"qa":                 nil,
		"batch":              nil,
		"comment":            nil,
		"inputreceiver":      nil,
	}
}

// generateBranches generates branches for schema section
func (g *ConditionNodeGenerator) generateBranches(unifiedNode *models.Node) ([]map[string]interface{}, error) {
	conditionConfig, _ := common.AsConditionConfig(unifiedNode.Config)
	branches := make([]map[string]interface{}, 0)

	// Add condition branches, excluding empty condition branches (default cases with level=999)
	for _, caseItem := range conditionConfig.Cases {
		// Skip empty condition branches (typically default cases with level=999)
		if len(caseItem.Conditions) == 0 {
			continue
		}

		branch, err := g.GenerateSchemaBranch(caseItem, unifiedNode)
		if err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}

	return branches, nil
}

// generateSelectorBranch generates a selector branch for nodes section
func (g *ConditionNodeGenerator) generateSelectorBranch(caseItem models.ConditionCase, unifiedNode *models.Node) map[string]interface{} {
	// Convert logical operator
	logic := g.mapLogicalOperator(caseItem.LogicalOperator)

	// Generate conditions
	conditions := make([]map[string]interface{}, 0)
	for _, condition := range caseItem.Conditions {
		cond := g.generateSelectorCondition(condition, unifiedNode)
		conditions = append(conditions, cond)
	}

	return map[string]interface{}{
		"condition": map[string]interface{}{
			"logic":      logic,
			"conditions": conditions,
		},
	}
}

// GenerateSchemaBranch generates a schema branch for schema section (public for iteration use)
func (g *ConditionNodeGenerator) GenerateSchemaBranch(caseItem models.ConditionCase, unifiedNode *models.Node) (map[string]interface{}, error) {
	// Convert logical operator
	logic := g.mapLogicalOperator(caseItem.LogicalOperator)

	// Generate conditions
	conditions := make([]map[string]interface{}, 0)
	for _, condition := range caseItem.Conditions {
		cond, err := g.generateSchemaCondition(condition, unifiedNode)
		if err != nil {
			return nil, err
		}
		conditions = append(conditions, cond)
	}

	return map[string]interface{}{
		"condition": map[string]interface{}{
			"logic":      logic,
			"conditions": conditions,
		},
	}, nil
}

// generateSelectorCondition generates a selector condition for nodes section
func (g *ConditionNodeGenerator) generateSelectorCondition(condition models.Condition, unifiedNode *models.Node) map[string]interface{} {
	// Map comparison operator
	operator := g.mapComparisonOperator(condition.ComparisonOperator)

	// Generate left operand (variable reference)
	leftOperand := g.generateVariableReference(condition.VariableSelector, unifiedNode, condition.VarType)

	// Generate right operand (literal value)
	rightOperand := g.generateLiteralValue(condition.Value, condition.VarType)

	return map[string]interface{}{
		"operator": operator,
		"left": map[string]interface{}{
			"name":      "",
			"input":     leftOperand,
			"left":      nil,
			"right":     nil,
			"variables": []interface{}{},
		},
		"right": map[string]interface{}{
			"name":      "",
			"input":     rightOperand,
			"left":      nil,
			"right":     nil,
			"variables": []interface{}{},
		},
	}
}

// generateSchemaCondition generates a schema condition for schema section
func (g *ConditionNodeGenerator) generateSchemaCondition(condition models.Condition, unifiedNode *models.Node) (map[string]interface{}, error) {
	// Map comparison operator
	operator := g.mapComparisonOperator(condition.ComparisonOperator)

	// Generate left operand (variable reference)
	leftOperand := g.generateSchemaVariableReference(condition.VariableSelector, unifiedNode, condition.VarType)

	// Generate right operand (literal value)
	rightOperand := g.generateSchemaLiteralValue(condition.Value, condition.VarType)

	return map[string]interface{}{
		"operator": operator,
		"left": map[string]interface{}{
			"input": leftOperand,
		},
		"right": map[string]interface{}{
			"input": rightOperand,
		},
	}, nil
}

// mapLogicalOperator maps unified logical operator to Coze logic type
func (g *ConditionNodeGenerator) mapLogicalOperator(logicalOperator string) int {
	switch logicalOperator {
	case "or":
		return 1
	case "and":
		return 2
	default:
		return 2 // Default to AND
	}
}

// mapComparisonOperator maps unified comparison operator to Coze operator type
// Based on Coze source code ConditionType enum:
func (g *ConditionNodeGenerator) mapComparisonOperator(compareOperator string) int {
	switch compareOperator {
	case "is", "equals":
		return 1 // Equal
	case "is_not", "not_equals":
		return 2 // NotEqual
	case "length_gt":
		return 3 // LengthGt
	case "length_gte":
		return 4 // LengthGtEqual
	case "length_lt":
		return 5 // LengthLt
	case "length_lte":
		return 6 // LengthLtEqual
	case "contains", "start_with":
		return 7 // Contains (iFlytek start_with maps to Coze Contains)
	case "not_contains":
		return 8 // NotContains
	case "empty", "is_empty":
		return 9 // Null
	case "not_empty", "is_not_empty":
		return 10 // NotNull (fixed: unified DSL uses is_not_empty)
	case "true":
		return 11 // True
	case "false":
		return 12 // False
	case "gt", ">", "greater_than":
		return 13 // Gt
	case "ge", ">=", "greater_equal":
		return 14 // GtEqual
	case "lt", "<", "less_than":
		return 15 // Lt
	case "le", "<=", "less_equal":
		return 16 // LtEqual
	default:
		return 1 // Default to Equal
	}
}

// generateVariableReference generates variable reference for nodes section
func (g *ConditionNodeGenerator) generateVariableReference(variableSelector []string, unifiedNode *models.Node, varType models.UnifiedDataType) map[string]interface{} {
	// Extract node ID and field name from variable selector
	if len(variableSelector) < 2 {
		return g.generateDefaultVariableReference(varType)
	}

	sourceNodeID := variableSelector[0]
	fieldName := variableSelector[1]

	// Map source node ID to Coze format
	cozeSourceNodeID := g.idGenerator.MapToCozeNodeID(sourceNodeID)

	return map[string]interface{}{
		"type": g.mapUnifiedTypeToCozeType(varType),
		"value": map[string]interface{}{
			"type": "ref",
			"content": map[string]interface{}{
				"blockID": cozeSourceNodeID,
				"name":    fieldName,
				"source":  "block-output",
			},
			"rawMeta": nil,
		},
	}
}

// generateSchemaVariableReference generates variable reference for schema section
func (g *ConditionNodeGenerator) generateSchemaVariableReference(variableSelector []string, unifiedNode *models.Node, varType models.UnifiedDataType) map[string]interface{} {
	// Extract node ID and field name from variable selector
	if len(variableSelector) < 2 {
		return g.generateDefaultSchemaVariableReference(varType)
	}

	sourceNodeID := variableSelector[0]
	fieldName := variableSelector[1]

	// Map source node ID to Coze format
	cozeSourceNodeID := g.idGenerator.MapToCozeNodeID(sourceNodeID)

	return map[string]interface{}{
		"type": g.mapUnifiedTypeToCozeTypeString(varType),
		"value": map[string]interface{}{
			"content": map[string]interface{}{
				"blockID": cozeSourceNodeID,
				"name":    fieldName,
				"source":  "block-output",
			},
			"type": "ref",
		},
	}
}

// generateLiteralValue generates literal value for nodes section
func (g *ConditionNodeGenerator) generateLiteralValue(value interface{}, varType models.UnifiedDataType) map[string]interface{} {
	return map[string]interface{}{
		"type": g.mapUnifiedTypeToCozeType(varType),
		"value": map[string]interface{}{
			"type":    "literal",
			"content": value,
			"rawMeta": map[string]interface{}{
				"type": g.mapUnifiedTypeToCozeRawMetaType(varType),
			},
		},
	}
}

// generateSchemaLiteralValue generates literal value for schema section
func (g *ConditionNodeGenerator) generateSchemaLiteralValue(value interface{}, varType models.UnifiedDataType) map[string]interface{} {
	return map[string]interface{}{
		"type": g.mapUnifiedTypeToCozeTypeString(varType),
		"value": map[string]interface{}{
			"content": value,
			"rawMeta": map[string]interface{}{
				"type": g.mapUnifiedTypeToCozeRawMetaType(varType),
			},
			"type": "literal",
		},
	}
}

// generateDefaultVariableReference generates default variable reference for nodes section
func (g *ConditionNodeGenerator) generateDefaultVariableReference(varType models.UnifiedDataType) map[string]interface{} {
	return map[string]interface{}{
		"type": g.mapUnifiedTypeToCozeType(varType),
		"value": map[string]interface{}{
			"type":    "literal",
			"content": "",
			"rawMeta": nil,
		},
	}
}

// generateDefaultSchemaVariableReference generates default variable reference for schema section
func (g *ConditionNodeGenerator) generateDefaultSchemaVariableReference(varType models.UnifiedDataType) map[string]interface{} {
	return map[string]interface{}{
		"type": g.mapUnifiedTypeToCozeTypeString(varType),
		"value": map[string]interface{}{
			"content": "",
			"type":    "literal",
		},
	}
}

// mapUnifiedTypeToCozeType maps unified data type to Coze type (capitalized for nodes section)
func (g *ConditionNodeGenerator) mapUnifiedTypeToCozeType(dataType models.UnifiedDataType) string {
	switch dataType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "float"
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeNumber:
		return "integer" // Default to integer for generic number
	default:
		return "string"
	}
}

// mapUnifiedTypeToCozeTypeString maps unified data type to Coze type string (lowercase for schema section)
func (g *ConditionNodeGenerator) mapUnifiedTypeToCozeTypeString(dataType models.UnifiedDataType) string {
	return g.mapUnifiedTypeToCozeType(dataType)
}

// mapUnifiedTypeToCozeRawMetaType maps unified data type to Coze rawMeta type
func (g *ConditionNodeGenerator) mapUnifiedTypeToCozeRawMetaType(dataType models.UnifiedDataType) int {
	switch dataType {
	case models.DataTypeString:
		return 1
	case models.DataTypeInteger:
		return 2
	case models.DataTypeBoolean:
		return 3
	case models.DataTypeFloat:
		return 4
	case models.DataTypeNumber:
		return 2 // Default to integer for generic number
	default:
		return 1
	}
}
