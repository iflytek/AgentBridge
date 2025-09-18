package generator

import (
	"ai-agents-transformer/internal/models"
	"fmt"
)

// StartNodeGenerator generates start nodes
type StartNodeGenerator struct {
	*BaseNodeGenerator
}

func NewStartNodeGenerator() *StartNodeGenerator {
	return &StartNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeStart),
	}
}

// GenerateNode generates a start node
func (g *StartNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeStart {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeStart)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Set start node specific data
	// Priority: get variable definitions from config, fallback to outputs
	if startConfig, ok := node.Config.(models.StartConfig); ok && len(startConfig.Variables) > 0 {
		difyNode.Data.Variables = g.generateVariablesFromConfig(startConfig.Variables)
	} else {
		// For start nodes, use outputs to generate variables (conceptual mapping)
		difyNode.Data.Variables = g.generateVariablesFromOutputs(node.Outputs)
	}

	// Start node needs empty config field (variables directly at data level)
	if difyNode.Data.Config == nil {
		difyNode.Data.Config = make(map[string]interface{})
	}

	// Restore Dify-specific fields from platform configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// generateVariablesFromOutputs generates variable definitions from outputs (for start nodes)
func (g *StartNodeGenerator) generateVariablesFromOutputs(outputs []models.Output) []DifyVariable {
	variables := make([]DifyVariable, 0, len(outputs))

	for _, output := range outputs {
		// Check if type is supported by Dify start node
		if !g.isDifyStartNodeSupportedType(output.Type) {
			// Skip unsupported complex types (object, array[object])
			continue
		}

		variable := DifyVariable{
			Label:    output.Label,
			Variable: output.Name,
			Type:     g.mapOutputTypeToDify(output.Type),
			Required: true, // Start node variables are usually required
			Options:  []string{},
		}

		// Apply common variable settings
		g.applyCommonVariableSettings(&variable, output.Default, nil)

		variables = append(variables, variable)
	}

	return variables
}

// generateVariablesFromConfig generates variable definitions from unified DSL config (recommended approach)
func (g *StartNodeGenerator) generateVariablesFromConfig(variables []models.Variable) []DifyVariable {
	difyVariables := make([]DifyVariable, 0, len(variables))

	for _, variable := range variables {
		// Check if type is supported by Dify start node
		varType := models.UnifiedDataType(variable.Type)
		if !g.isDifyStartNodeSupportedType(varType) {
			// Skip unsupported complex types (object, array[object])
			continue
		}

		difyVar := DifyVariable{
			Label:    variable.Label,
			Variable: variable.Name,
			Type:     g.mapOutputTypeToDify(varType),
			Required: variable.Required,
			Options:  []string{},
		}

		// Apply common variable settings
		g.applyCommonVariableSettings(&difyVar, variable.Default, variable.Constraints)

		difyVariables = append(difyVariables, difyVar)
	}

	return difyVariables
}

// applyCommonVariableSettings applies common settings for variables
func (g *StartNodeGenerator) applyCommonVariableSettings(variable *DifyVariable, defaultValue interface{}, constraints *models.Constraints) {
	// Apply all variable settings in sequence
	g.fixVariableSpellingErrors(variable)
	g.setVariableDisplayName(variable, defaultValue)
	g.setVariableLengthLimits(variable, constraints)
	g.setVariableOptions(variable, constraints)
}

// fixVariableSpellingErrors fixes common variable name spelling errors
func (g *StartNodeGenerator) fixVariableSpellingErrors(variable *DifyVariable) {
	if variable.Variable == "intput_text_01" {
		variable.Variable = "input_text_01"
	}
}

// setVariableDisplayName sets appropriate display name for the variable
func (g *StartNodeGenerator) setVariableDisplayName(variable *DifyVariable, defaultValue interface{}) {
	// Handle special case for AGENT_USER_INPUT
	if g.isAgentUserInputVariable(variable) {
		g.setAgentUserInputLabel(variable)
		return
	}

	// Set label from default value or variable name
	g.setLabelFromDefaultOrName(variable, defaultValue)
}

// isAgentUserInputVariable checks if variable is AGENT_USER_INPUT type
func (g *StartNodeGenerator) isAgentUserInputVariable(variable *DifyVariable) bool {
	return variable.Variable == "AGENT_USER_INPUT"
}

// setAgentUserInputLabel sets friendly label for AGENT_USER_INPUT
func (g *StartNodeGenerator) setAgentUserInputLabel(variable *DifyVariable) {
	if variable.Label == "" || variable.Label == "AGENT_USER_INPUT" {
		variable.Label = "User Input Content"
	}
}

// setLabelFromDefaultOrName sets label from default value or variable name
func (g *StartNodeGenerator) setLabelFromDefaultOrName(variable *DifyVariable, defaultValue interface{}) {
	// If no label, use name as label
	if variable.Label == "" {
		variable.Label = variable.Variable
	}

	// Use default value as label if it's meaningful
	if g.shouldUseDefaultAsLabel(defaultValue, variable.Variable) {
		if defaultStr := g.extractStringFromDefault(defaultValue); defaultStr != "" {
			variable.Label = defaultStr
		}
	}
}

// shouldUseDefaultAsLabel checks if default value should be used as label
func (g *StartNodeGenerator) shouldUseDefaultAsLabel(defaultValue interface{}, variableName string) bool {
	if defaultValue == nil {
		return false
	}
	
	defaultStr := g.extractStringFromDefault(defaultValue)
	return defaultStr != "" && defaultStr != variableName
}

// extractStringFromDefault extracts string from default value
func (g *StartNodeGenerator) extractStringFromDefault(defaultValue interface{}) string {
	if defaultStr, ok := defaultValue.(string); ok {
		return defaultStr
	}
	return ""
}

// setVariableLengthLimits sets length limits for the variable
func (g *StartNodeGenerator) setVariableLengthLimits(variable *DifyVariable, constraints *models.Constraints) {
	if g.hasConstraintMaxLength(constraints) {
		variable.MaxLength = constraints.MaxLength
		return
	}

	// Set default length limits based on type
	variable.MaxLength = g.getDefaultMaxLengthForType(variable.Type)
}

// hasConstraintMaxLength checks if constraints have max length setting
func (g *StartNodeGenerator) hasConstraintMaxLength(constraints *models.Constraints) bool {
	return constraints != nil && constraints.MaxLength > 0
}

// getDefaultMaxLengthForType returns default max length for variable type
func (g *StartNodeGenerator) getDefaultMaxLengthForType(variableType string) int {
	switch variableType {
	case "text-input":
		return 200
	case "number":
		return 48
	default:
		return 100
	}
}

// setVariableOptions sets options for the variable from constraints
func (g *StartNodeGenerator) setVariableOptions(variable *DifyVariable, constraints *models.Constraints) {
	if !g.hasConstraintOptions(constraints) {
		return
	}

	options := g.extractStringOptions(constraints.Options)
	variable.Options = options
}

// hasConstraintOptions checks if constraints have options
func (g *StartNodeGenerator) hasConstraintOptions(constraints *models.Constraints) bool {
	return constraints != nil && len(constraints.Options) > 0
}

// extractStringOptions extracts string options from constraint options
func (g *StartNodeGenerator) extractStringOptions(constraintOptions []interface{}) []string {
	options := make([]string, 0, len(constraintOptions))
	
	for _, option := range constraintOptions {
		if optionStr, ok := option.(string); ok {
			options = append(options, optionStr)
		}
	}
	
	return options
}

// isDifyStartNodeSupportedType checks if type is supported by Dify start node
func (g *StartNodeGenerator) isDifyStartNodeSupportedType(dataType models.UnifiedDataType) bool {
	// Based on Dify platform actual limitations, start nodes only support:
	// - text-input (string)
	// - number (Dify does not distinguish integer/float, unified as number)
	// - paragraph (string variant)
	// - select (boolean UI representation)
	// - file-list 

	switch dataType {
	case models.DataTypeString:
		return true
	case models.DataTypeInteger:  // Map to Dify number type
		return true
	case models.DataTypeFloat:    // Map to Dify number type
		return true
	case models.DataTypeNumber:   // Maintain backward compatibility
		return true
	case models.DataTypeBoolean:
		return true
	case models.DataTypeArrayString:
		return true // Can use text-input, user inputs comma-separated strings
	case models.DataTypeObject:
		return false // Dify start nodes don't support object type input
	case models.DataTypeArrayObject:
		return false // Dify start nodes don't support object array type input
	default:
		return false // Unknown types default to unsupported
	}
}

// mapOutputTypeToDify maps output types to Dify UI component types
// Note: This function maps to UI component types, not data types, so special handling is needed
func (g *StartNodeGenerator) mapOutputTypeToDify(dataType models.UnifiedDataType) string {
	// UI component type mapping (different from data type mapping)
	// Only handle supported types, as unsupported types have been filtered out
	switch dataType {
	case models.DataTypeString:
		return "text-input"
	case models.DataTypeInteger:  // Dify uses unified number UI component, no distinction between integer/float
		return "number"
	case models.DataTypeFloat:    // Dify uses unified number UI component, no distinction between integer/float
		return "number"
	case models.DataTypeNumber:   // Maintain backward compatibility
		return "number"
	case models.DataTypeBoolean:
		return "select"
	case models.DataTypeArrayString:
		return "text-input" // User inputs comma-separated strings
	default:
		return "text-input" // Default to text input
	}
}

// restoreDifyPlatformConfig restores Dify platform-specific configuration
func (g *StartNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	g.restoreVariablesConfig(config, node)
	g.restoreNodeMetadata(config, node)
}

// restoreVariablesConfig restores variable configuration from platform config
func (g *StartNodeGenerator) restoreVariablesConfig(config map[string]interface{}, node *DifyNode) {
	variablesConfig, exists := config["variables"].([]interface{})
	if !exists {
		return
	}
	
	variables := make([]DifyVariable, 0, len(variablesConfig))
	for _, varConfig := range variablesConfig {
		if varMap, ok := varConfig.(map[string]interface{}); ok {
			variable := g.buildVariableFromConfig(varMap)
			variables = append(variables, variable)
		}
	}
	node.Data.Variables = variables
}

// buildVariableFromConfig builds DifyVariable from configuration map
func (g *StartNodeGenerator) buildVariableFromConfig(varMap map[string]interface{}) DifyVariable {
	variable := DifyVariable{
		Options: []string{},
	}
	
	g.setVariableBasicFields(&variable, varMap)
	g.setVariableOptionsFromConfig(&variable, varMap)
	
	return variable
}

// setVariableBasicFields sets basic variable fields from config map
func (g *StartNodeGenerator) setVariableBasicFields(variable *DifyVariable, varMap map[string]interface{}) {
	if label, ok := varMap["label"].(string); ok {
		variable.Label = label
	}
	if varName, ok := varMap["variable"].(string); ok {
		variable.Variable = varName
	}
	if varType, ok := varMap["type"].(string); ok {
		variable.Type = varType
	}
	if required, ok := varMap["required"].(bool); ok {
		variable.Required = required
	}
	if maxLength, ok := varMap["max_length"].(int); ok {
		variable.MaxLength = maxLength
	}
}

// setVariableOptions sets variable options from config map (overloaded method with different signature)
func (g *StartNodeGenerator) setVariableOptionsFromConfig(variable *DifyVariable, varMap map[string]interface{}) {
	options, ok := varMap["options"].([]interface{})
	if !ok {
		return
	}
	
	for _, option := range options {
		if optionStr, ok := option.(string); ok {
			variable.Options = append(variable.Options, optionStr)
		}
	}
}

// restoreNodeMetadata restores node metadata from platform config
func (g *StartNodeGenerator) restoreNodeMetadata(config map[string]interface{}, node *DifyNode) {
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}
