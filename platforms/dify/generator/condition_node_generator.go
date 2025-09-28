package generator

import (
	"github.com/iflytek/agentbridge/internal/models"
	"github.com/iflytek/agentbridge/platforms/common"
	"fmt"
	"strings"
)

// ConditionNodeGenerator generates conditional branch nodes.
type ConditionNodeGenerator struct {
	*BaseNodeGenerator
	variableSelectorConverter *VariableSelectorConverter
	caseIDCache               map[string]string // Cache mapping from original case_id to Dify case_id
	usedIDs                   map[string]bool   // Track used IDs to ensure uniqueness within the same node
}

func NewConditionNodeGenerator() *ConditionNodeGenerator {
	return &ConditionNodeGenerator{
		BaseNodeGenerator:         NewBaseNodeGenerator(models.NodeTypeCondition),
		variableSelectorConverter: NewVariableSelectorConverter(),
		caseIDCache:               make(map[string]string),
		usedIDs:                   make(map[string]bool),
	}
}

// GenerateNode generates a conditional branch node.
func (g *ConditionNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeCondition {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeCondition)
	}

	// Generate basic node structure
	difyNode := g.generateBaseNode(node)

	// Set conditional branch node specific data
	difyNode.Data.Cases = g.generateCases(node)

	// Save case_id mapping to node's PlatformConfig for edge generator use
	g.saveCaseIDMapping(node, &difyNode)

	// Restore Dify specific fields from platform specific configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// SetNodeMapping sets node mapping for variable selector converter
func (g *ConditionNodeGenerator) SetNodeMapping(nodes []models.Node) {
	g.variableSelectorConverter.SetNodeMapping(nodes)
}

// generateCases generates conditional branch cases (simplified version, ensures ID consistency).
func (g *ConditionNodeGenerator) generateCases(node models.Node) []map[string]interface{} {
	// Restore original conditional configuration from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		if cases, exists := difyConfig["cases"].([]interface{}); exists {
			// Convert to correct format
			result := make([]map[string]interface{}, len(cases))
			for i, caseInterface := range cases {
				if caseMap, ok := caseInterface.(map[string]interface{}); ok {
					result[i] = caseMap
				}
			}
			return result
		}
	}

	// Get conditional configuration from unified DSL config
	if conditionConfig, ok := common.AsConditionConfig(node.Config); ok && conditionConfig != nil && len(conditionConfig.Cases) > 0 {
		cases := make([]map[string]interface{}, 0)

		// Sort by level to ensure correct branch order
		sortedCases := g.sortCasesByLevel(conditionConfig.Cases)

		actualCaseIndex := 0 // Actual generated case index
		for _, caseItem := range sortedCases {
			// Skip default branches (branches with empty conditions) - Dify handles ELSE branches automatically
			if len(caseItem.Conditions) == 0 {
				// Cache mapping for default branch (edge generator needs it), but don't generate explicit case
				g.caseIDCache[caseItem.CaseID] = "false" // Default branch maps to false
				continue
			}

			// Generate stable case_id - Key: this ID will be used for edge connection sourceHandle
			caseID := g.generateDifyCaseID(caseItem, actualCaseIndex)

			// Cache mapping relationship for edge generator use
			g.caseIDCache[caseItem.CaseID] = caseID

			difyCase := map[string]interface{}{
				"case_id":          caseID,
				"id":               caseID,
				"logical_operator": g.mapLogicalOperator(caseItem.LogicalOperator),
				"conditions":       g.convertConditions(caseItem.Conditions, node),
			}

			cases = append(cases, difyCase)
			actualCaseIndex++
		}

		return cases
	}

	return []map[string]interface{}{}
}

// generateConditionID generates condition ID.
func (g *ConditionNodeGenerator) generateConditionID() string {
	return "condition-" + generateRandomUUID()
}

// mapLogicalOperator maps logical operators.
func (g *ConditionNodeGenerator) mapLogicalOperator(operator string) string {
	switch operator {
	case "and", "AND":
		return "and"
	case "or", "OR":
		return "or"
	default:
		return "and" // Default to and
	}
}

// convertConditions converts condition list.
func (g *ConditionNodeGenerator) convertConditions(conditions []models.Condition, node models.Node) []map[string]interface{} {
	difyConditions := make([]map[string]interface{}, 0, len(conditions))

	for _, condition := range conditions {
		// Handle condition values - for empty value check operators, keep original values
		conditionValue := condition.Value
		mappedOperator := g.mapComparisonOperator(condition.ComparisonOperator)
		// Note: For empty/not empty operators, keep original values unchanged

		difyCondition := map[string]interface{}{
			"id":                  g.generateConditionID(),
			"comparison_operator": mappedOperator,
			"value":               conditionValue,
			"varType":             g.mapVarType(condition.VarType),
			"variable_selector":   g.mapVariableSelector(condition.VariableSelector, node),
		}

		difyConditions = append(difyConditions, difyCondition)
	}

	return difyConditions
}

// sortCasesByLevel sorts branches by level.
func (g *ConditionNodeGenerator) sortCasesByLevel(cases []models.ConditionCase) []models.ConditionCase {
	sortedCases := make([]models.ConditionCase, len(cases))
	copy(sortedCases, cases)

	// Use simple selection sort for easier understanding
	for i := 0; i < len(sortedCases)-1; i++ {
		for j := i + 1; j < len(sortedCases); j++ {
			if sortedCases[i].Level > sortedCases[j].Level {
				sortedCases[i], sortedCases[j] = sortedCases[j], sortedCases[i]
			}
		}
	}

	return sortedCases
}

// generateDifyCaseID generates semantic case IDs with fallback to compatibility.
func (g *ConditionNodeGenerator) generateDifyCaseID(caseItem models.ConditionCase, caseIndex int) string {
	// If original ID is already in Dify compatible format, use it directly
	if caseItem.CaseID == "true" || caseItem.CaseID == "false" {
		return caseItem.CaseID
	}

	// Check if it's existing UUID format (keep unchanged for backward compatibility)
	if len(caseItem.CaseID) > 10 && !strings.Contains(caseItem.CaseID, "branch_one_of::") {
		return caseItem.CaseID
	}

	// Generate semantic case ID based on conditions
	semanticID := g.generateSemanticCaseID(caseItem, caseIndex)
	return g.ensureUniqueID(semanticID)
}

// generateSemanticCaseID creates meaningful case IDs based on condition content.
func (g *ConditionNodeGenerator) generateSemanticCaseID(caseItem models.ConditionCase, caseIndex int) string {
	// Handle empty conditions (default branch)
	if len(caseItem.Conditions) == 0 {
		return "default"
	}

	// Single condition case - generate descriptive ID
	if len(caseItem.Conditions) == 1 {
		condition := caseItem.Conditions[0]
		return g.buildSingleConditionID(condition)
	}

	// Multiple conditions case - use logical operator and primary condition
	if len(caseItem.Conditions) > 1 {
		logicalOp := strings.ToLower(caseItem.LogicalOperator)
		primaryCondition := caseItem.Conditions[0] // Use first condition as primary identifier
		primaryID := g.buildSingleConditionID(primaryCondition)

		return fmt.Sprintf("%s_%s_%d", logicalOp, primaryID, len(caseItem.Conditions))
	}

	// Fallback: use index-based ID
	if caseIndex == 0 {
		return "true"
	}
	return fmt.Sprintf("case_%d", caseIndex+1)
}

// buildSingleConditionID creates an ID for a single condition.
func (g *ConditionNodeGenerator) buildSingleConditionID(condition models.Condition) string {
	value := g.extractConditionValue(condition)

	operatorTemplate := g.getOperatorTemplate(condition.ComparisonOperator)
	return g.buildConditionIDFromTemplate(operatorTemplate, value)
}

// extractConditionValue safely extracts and sanitizes condition value
func (g *ConditionNodeGenerator) extractConditionValue(condition models.Condition) string {
	var valueStr string
	if condition.Value != nil {
		valueStr = fmt.Sprintf("%v", condition.Value)
	} else {
		valueStr = ""
	}
	return g.sanitizeValueForID(valueStr)
}

// getOperatorTemplate returns the template for a given operator
func (g *ConditionNodeGenerator) getOperatorTemplate(operator string) string {
	templates := map[string]string{
		"contains":      "contains_%s",
		"not_contains":  "not_contains_%s",
		"not contains":  "not_contains_%s",
		"is":            "equals_%s",
		"equals":        "equals_%s",
		"is_not":        "not_equals_%s",
		"is not":        "not_equals_%s",
		"start_with":    "starts_%s",
		"start with":    "starts_%s",
		"end_with":      "ends_%s",
		"end with":      "ends_%s",
		"empty":         "is_empty",
		"not_empty":     "not_empty",
		"not empty":     "not_empty",
		"greater_than":  "gt_%s",
		">":             "gt_%s",
		"less_than":     "lt_%s",
		"<":             "lt_%s",
		"greater_equal": "gte_%s",
		">=":            "gte_%s",
		"less_equal":    "lte_%s",
		"<=":            "lte_%s",
	}

	if template, exists := templates[operator]; exists {
		return template
	}

	return operator + "_%s"
}

// buildConditionIDFromTemplate builds the final condition ID from template
func (g *ConditionNodeGenerator) buildConditionIDFromTemplate(template, value string) string {
	if strings.Contains(template, "%s") {
		return fmt.Sprintf(template, value)
	}
	return template
}

// sanitizeValueForID cleans and shortens values for use in case IDs.
func (g *ConditionNodeGenerator) sanitizeValueForID(value string) string {
	if value == "" {
		return "empty"
	}

	cleaned := g.filterValidCharacters(value)
	cleaned = g.truncateToMaxLength(cleaned)
	cleaned = strings.Trim(cleaned, "_")

	return g.getFallbackIfEmpty(cleaned)
}

// filterValidCharacters filters and keeps only valid characters for ID
func (g *ConditionNodeGenerator) filterValidCharacters(value string) string {
	var result strings.Builder

	for _, r := range value {
		if g.isValidIDCharacter(r) {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' {
			result.WriteRune('_')
		}
		// Skip other special characters
	}

	return result.String()
}

// isValidIDCharacter checks if a rune is valid for ID usage
func (g *ConditionNodeGenerator) isValidIDCharacter(r rune) bool {
	return (r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') ||
		r == '_' ||
		(r >= 0x4e00 && r <= 0x9fff) // Chinese characters range
}

// truncateToMaxLength truncates the string to maximum allowed length
func (g *ConditionNodeGenerator) truncateToMaxLength(value string) string {
	if len(value) <= 20 {
		return value
	}

	runes := []rune(value)
	if len(runes) > 20 {
		return string(runes[:20])
	}

	return value
}

// getFallbackIfEmpty returns a fallback value if the cleaned string is empty
func (g *ConditionNodeGenerator) getFallbackIfEmpty(cleaned string) string {
	if cleaned == "" {
		return "condition"
	}
	return cleaned
}

// ensureUniqueID ensures the generated case ID doesn't conflict with existing ones.
func (g *ConditionNodeGenerator) ensureUniqueID(baseID string) string {
	// Track used IDs to avoid conflicts within the same node
	if g.usedIDs == nil {
		g.usedIDs = make(map[string]bool)
	}

	if !g.usedIDs[baseID] {
		g.usedIDs[baseID] = true
		return baseID
	}

	// If base ID exists, try numbered variants
	for i := 1; i <= 99; i++ {
		candidateID := fmt.Sprintf("%s_%d", baseID, i)
		if !g.usedIDs[candidateID] {
			g.usedIDs[candidateID] = true
			return candidateID
		}
	}

	// Ultimate fallback: random suffix
	suffix := generateShortID(4)
	finalID := fmt.Sprintf("%s_%s", baseID, suffix)
	g.usedIDs[finalID] = true
	return finalID
}

// saveCaseIDMapping saves case_id mapping to node configuration.
func (g *ConditionNodeGenerator) saveCaseIDMapping(originalNode models.Node, difyNode *DifyNode) {
	// Ensure PlatformConfig.Dify exists
	if originalNode.PlatformConfig.Dify == nil {
		return
	}

	// Save cached mapping to Dify configuration
	if len(g.caseIDCache) > 0 {
		originalNode.PlatformConfig.Dify["case_id_mapping"] = g.caseIDCache
	}
}

// mapComparisonOperator maps comparison operators.
func (g *ConditionNodeGenerator) mapComparisonOperator(operator string) string {
	operatorMap := g.getComparisonOperatorMap()

	if mappedOp, exists := operatorMap[operator]; exists {
		return mappedOp
	}

	return "contains" // Default to contains
}

// getComparisonOperatorMap returns the mapping table for comparison operators
func (g *ConditionNodeGenerator) getComparisonOperatorMap() map[string]string {
	return map[string]string{
		"contains":      "contains",
		"not_contains":  "not contains",
		"is":            "is",
		"equals":        "is",
		"==":            "is",
		"is_not":        "is not",
		"not_equals":    "is not",
		"!=":            "is not",
		"start_with":    "start with",
		"starts_with":   "start with",
		"end_with":      "end with",
		"ends_with":     "end with",
		"empty":         "empty",
		"is_empty":      "empty",
		"not_empty":     "not empty",
		"is_not_empty":  "not empty",
		"greater_than":  ">",
		">":             ">",
		"less_than":     "<",
		"<":             "<",
		"greater_equal": ">=",
		">=":            ">=",
		"less_equal":    "<=",
		"<=":            "<=",
	}
}

// mapVarType maps variable types.
func (g *ConditionNodeGenerator) mapVarType(varType models.UnifiedDataType) string {
	switch varType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger: // Map integer to number
		return "number"
	case models.DataTypeFloat: // Map float to number
		return "number"
	case models.DataTypeNumber: // Maintain backward compatibility
		return "number"
	case models.DataTypeBoolean:
		return "boolean"
	default:
		return "string"
	}
}

// mapVariableSelector maps variable selectors.
func (g *ConditionNodeGenerator) mapVariableSelector(selector []string, node models.Node) []string {
	if len(selector) == 0 {
		return g.handleEmptySelector(node)
	}

	if len(selector) == 1 {
		return g.handleSingleSelector(selector[0], node)
	}

	if len(selector) >= 2 {
		return g.handleMultiSelector(selector)
	}

	return selector
}

// handleEmptySelector handles empty selector case
func (g *ConditionNodeGenerator) handleEmptySelector(node models.Node) []string {
	if len(node.Inputs) == 0 || node.Inputs[0].Reference == nil {
		return []string{}
	}

	return g.convertVariableReference(node.Inputs[0].Reference)
}

// handleSingleSelector handles single element selector case
func (g *ConditionNodeGenerator) handleSingleSelector(selectorValue string, node models.Node) []string {
	firstValidInput := g.findMatchingInput(selectorValue, node)

	if firstValidInput != nil {
		return g.convertVariableReference(firstValidInput)
	}

	return []string{selectorValue}
}

// findMatchingInput finds exact match or first valid input as fallback
func (g *ConditionNodeGenerator) findMatchingInput(selectorValue string, node models.Node) *models.VariableReference {
	var firstValidInput *models.VariableReference

	for _, input := range node.Inputs {
		if input.Reference != nil && input.Reference.NodeID != "" {
			if firstValidInput == nil {
				firstValidInput = input.Reference
			}

			if input.Reference.OutputName == selectorValue {
				return input.Reference
			}
		}
	}

	return firstValidInput
}

// handleMultiSelector handles multi-element selector case
func (g *ConditionNodeGenerator) handleMultiSelector(selector []string) []string {
	tempRef := &models.VariableReference{
		Type:       models.ReferenceTypeNodeOutput,
		NodeID:     selector[0],
		OutputName: selector[1],
	}

	if convertedSelector, err := g.variableSelectorConverter.ConvertVariableReference(tempRef); err == nil {
		return convertedSelector
	}

	return selector
}

// convertVariableReference converts variable reference using converter with fallback
func (g *ConditionNodeGenerator) convertVariableReference(ref *models.VariableReference) []string {
	valueSelector, err := g.variableSelectorConverter.ConvertVariableReference(ref)
	if err != nil {
		return []string{ref.NodeID, ref.OutputName}
	}
	return valueSelector
}

// restoreDifyPlatformConfig restores Dify platform specific configuration.
func (g *ConditionNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	// Restore conditional case configuration - directly at data level, not under config
	if cases, exists := config["cases"].([]interface{}); exists {
		casesSlice := make([]map[string]interface{}, len(cases))
		for i, caseInterface := range cases {
			if caseMap, ok := caseInterface.(map[string]interface{}); ok {
				casesSlice[i] = caseMap
			}
		}
		node.Data.Cases = casesSlice
	}

	// Restore other node specific configuration
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}
