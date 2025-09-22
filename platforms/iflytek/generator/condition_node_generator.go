package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"fmt"
	"strings"
)

// ConditionNodeGenerator handles condition branch node generation
type ConditionNodeGenerator struct {
	*BaseNodeGenerator
	idMapping        map[string]string  // Dify ID to iFlytek SparkAgent ID mapping
	nodeTitleMapping map[string]string  // iFlytek SparkAgent ID to node title mapping
	branchIDMapping  map[string]string  // Dify case ID to iFlytek branch_one_of ID mapping
	unifiedDSL       *models.UnifiedDSL // Full DSL context for type inference
}

func NewConditionNodeGenerator() *ConditionNodeGenerator {
	return &ConditionNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeCondition),
		idMapping:         make(map[string]string),
		nodeTitleMapping:  make(map[string]string),
		branchIDMapping:   make(map[string]string),
	}
}

// SetIDMapping sets ID mapping
func (g *ConditionNodeGenerator) SetIDMapping(idMapping map[string]string) {
	g.idMapping = idMapping
}

// SetNodeTitleMapping sets node title mapping
func (g *ConditionNodeGenerator) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	g.nodeTitleMapping = nodeTitleMapping
}

// SetUnifiedDSL sets the complete DSL context for type inference
func (g *ConditionNodeGenerator) SetUnifiedDSL(dsl *models.UnifiedDSL) {
	g.unifiedDSL = dsl
}

// GenerateNode generates condition branch node
func (g *ConditionNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	// generate basic node information
	iflytekNode := g.generateBasicNodeInfo(node)
	iflytekNode.Type = "分支器"

	// set node metadata
	iflytekNode.Data.NodeMeta = IFlytekNodeMeta{
		AliasName: "分支器",
		NodeType:  "分支器",
	}

	// set icon and description
	iflytekNode.Data.Icon = g.getNodeIcon(models.NodeTypeCondition)
	iflytekNode.Data.Description = "根据设立的条件，判断选择分支走向"

	// set input/output permissions
	iflytekNode.Data.AllowInputReference = true
	iflytekNode.Data.AllowOutputReference = false

	// generate special inputs for condition branch node (including variable and literal inputs)
	inputs, inputIDMap := g.generateConditionInputs(node)
	iflytekNode.Data.Inputs = inputs

	// condition branch node has no outputs
	iflytekNode.Data.Outputs = []IFlytekOutput{}

	// Generate nodeParam (use inputIDMap to correctly set conditions)
	iflytekNode.Data.NodeParam = g.generateNodeParamWithInputIDs(node, inputIDMap)

	// Generate references
	iflytekNode.Data.References = g.generateConditionReferences(node)

	// Set other properties
	iflytekNode.Data.Status = ""
	iflytekNode.Data.Updatable = false

	return iflytekNode, nil
}

// generateConditionInputs generates inputs for condition branch node (variable inputs + literal inputs)
func (g *ConditionNodeGenerator) generateConditionInputs(node models.Node) ([]IFlytekInput, map[string]string) {
	var inputs []IFlytekInput
	inputIDMap := make(map[string]string)
	inputCounter := 0

	// Extract condition configuration and process inputs
	condConfig, ok := common.AsConditionConfig(node.Config)
	if !ok || condConfig == nil {
		return inputs, inputIDMap
	}

	for _, caseItem := range condConfig.Cases {
		inputs, inputIDMap, inputCounter = g.processCaseConditions(caseItem.Conditions, inputs, inputIDMap, inputCounter)
	}

	return inputs, inputIDMap
}

// processCaseConditions processes all conditions in a case
func (g *ConditionNodeGenerator) processCaseConditions(conditions []models.Condition, inputs []IFlytekInput, inputIDMap map[string]string, inputCounter int) ([]IFlytekInput, map[string]string, int) {
	for _, condition := range conditions {
		inputs, inputIDMap, inputCounter = g.processConditionInputs(condition, inputs, inputIDMap, inputCounter)
	}
	return inputs, inputIDMap, inputCounter
}

// processConditionInputs processes variable and literal inputs for a single condition
func (g *ConditionNodeGenerator) processConditionInputs(condition models.Condition, inputs []IFlytekInput, inputIDMap map[string]string, inputCounter int) ([]IFlytekInput, map[string]string, int) {
	if len(condition.VariableSelector) < 2 {
		return inputs, inputIDMap, inputCounter
	}

	sourceNodeID := condition.VariableSelector[0]
	sourceOutput := condition.VariableSelector[1]

	// Process variable reference input
	inputs, inputIDMap, inputCounter = g.processVariableInput(sourceNodeID, sourceOutput, inputs, inputIDMap, inputCounter)

	// Process literal input
	inputs, inputIDMap, inputCounter = g.processLiteralInput(condition.Value, inputs, inputIDMap, inputCounter)

	return inputs, inputIDMap, inputCounter
}

// processVariableInput processes variable reference input generation
func (g *ConditionNodeGenerator) processVariableInput(sourceNodeID, sourceOutput string, inputs []IFlytekInput, inputIDMap map[string]string, inputCounter int) ([]IFlytekInput, map[string]string, int) {
	varKey := fmt.Sprintf("var_%s_%s", sourceNodeID, sourceOutput)
	if _, exists := inputIDMap[varKey]; exists {
		return inputs, inputIDMap, inputCounter
	}

	varInputID := g.generateInputID()
	inputIDMap[varKey] = varInputID

	mappedNodeID := g.getMappedNodeID(sourceNodeID)
	inputName := g.generateInputName(inputCounter)
	inputCounter++

	varInput := g.createVariableReferenceInput(varInputID, inputName, sourceOutput, mappedNodeID)
	inputs = append(inputs, varInput)

	return inputs, inputIDMap, inputCounter
}

// processLiteralInput processes literal input generation
func (g *ConditionNodeGenerator) processLiteralInput(value interface{}, inputs []IFlytekInput, inputIDMap map[string]string, inputCounter int) ([]IFlytekInput, map[string]string, int) {
	literalKey := fmt.Sprintf("literal_%v", value)
	if _, exists := inputIDMap[literalKey]; exists {
		return inputs, inputIDMap, inputCounter
	}

	literalInputID := g.generateInputID()
	inputIDMap[literalKey] = literalInputID

	inputName := g.generateInputName(inputCounter)
	inputCounter++

	literalInput := g.createLiteralInput(literalInputID, inputName, value)
	inputs = append(inputs, literalInput)

	return inputs, inputIDMap, inputCounter
}

// getMappedNodeID gets the mapped node ID from ID mapping
func (g *ConditionNodeGenerator) getMappedNodeID(sourceNodeID string) string {
	if g.idMapping == nil {
		return sourceNodeID
	}

	if mapped, exists := g.idMapping[sourceNodeID]; exists {
		return mapped
	}

	return sourceNodeID
}

// generateInputName generates input name based on counter
func (g *ConditionNodeGenerator) generateInputName(inputCounter int) string {
	if inputCounter == 0 {
		return "input"
	}
	return fmt.Sprintf("input%d", inputCounter)
}

// createVariableReferenceInput creates variable reference input structure
func (g *ConditionNodeGenerator) createVariableReferenceInput(inputID, inputName, sourceOutput, mappedNodeID string) IFlytekInput {
	// Use unified data type mapping system instead of hardcoded type determination
	dataType := g.inferDataTypeFromOutput(sourceOutput, mappedNodeID)

	return IFlytekInput{
		ID:         inputID,
		Name:       inputName,
		NameErrMsg: "",
		Schema: IFlytekSchema{
			Type: dataType,
			Value: &IFlytekSchemaValue{
				Type: "ref",
				Content: &IFlytekRefContent{
					Name:   sourceOutput,
					ID:     g.generateRefID(),
					NodeID: mappedNodeID,
				},
				ContentErrMsg: "",
			},
		},
	}
}

// createLiteralInput creates literal input structure
func (g *ConditionNodeGenerator) createLiteralInput(inputID, inputName string, value interface{}) IFlytekInput {
	// Convert numeric values to string format as required by iFlytek
	var content interface{}
	switch v := value.(type) {
	case int:
		content = fmt.Sprintf("%d", v)
	case float64:
		// Check if it's actually an integer
		if v == float64(int(v)) {
			content = fmt.Sprintf("%.0f", v)
		} else {
			content = fmt.Sprintf("%g", v)
		}
	case string:
		// String values should be quoted for literals
		content = v
	default:
		content = fmt.Sprintf("%v", v)
	}

	return IFlytekInput{
		ID:         inputID,
		Name:       inputName,
		NameErrMsg: "",
		Schema: IFlytekSchema{
			Type: "string",
			Value: &IFlytekSchemaValue{
				Type:          "literal",
				Content:       content,
				ContentErrMsg: "",
			},
		},
	}
}

// generateNodeParamWithInputIDs generates node parameters using input ID mapping
func (g *ConditionNodeGenerator) generateNodeParamWithInputIDs(node models.Node, inputIDMap map[string]string) map[string]interface{} {
	nodeParam := map[string]interface{}{
		"uid":   "20718349453",
		"appId": "12a0a7e2",
	}

	// Extract condition branch information from configuration
	if condConfig, ok := common.AsConditionConfig(node.Config); ok && condConfig != nil {
		cases := g.generateCasesWithInputIDs(condConfig.Cases, inputIDMap)
		nodeParam["cases"] = cases
	}

	return nodeParam
}

// generateCasesWithInputIDs generates condition branch cases using input ID mapping
func (g *ConditionNodeGenerator) generateCasesWithInputIDs(cases []models.ConditionCase, inputIDMap map[string]string) []map[string]interface{} {
	var iflytekCases []map[string]interface{}

	// Generate actual condition branches
	iflytekCases = g.generateActualConditionBranches(cases, inputIDMap, iflytekCases)

	// Add default branch
	iflytekCases = g.addDefaultBranch(iflytekCases)

	return iflytekCases
}

// generateActualConditionBranches generates condition branches from cases
func (g *ConditionNodeGenerator) generateActualConditionBranches(cases []models.ConditionCase, inputIDMap map[string]string, iflytekCases []map[string]interface{}) []map[string]interface{} {
	for i, caseItem := range cases {
		iflytekCase := g.createConditionCase(caseItem, i+1, inputIDMap)
		g.saveBranchIDMappings(caseItem.CaseID, iflytekCase["id"].(string), i+1)
		iflytekCases = append(iflytekCases, iflytekCase)
	}
	return iflytekCases
}

// createConditionCase creates a single condition case structure
func (g *ConditionNodeGenerator) createConditionCase(caseItem models.ConditionCase, level int, inputIDMap map[string]string) map[string]interface{} {
	branchID := g.generateBranchID(caseItem.CaseID)

	iflytekCase := map[string]interface{}{
		"level":           level,
		"logicalOperator": g.mapLogicalOperator(caseItem.LogicalOperator),
		"id":              branchID,
	}

	// Generate condition list
	conditions := g.generateConditionsWithInputIDs(caseItem.Conditions, inputIDMap)
	iflytekCase["conditions"] = conditions

	return iflytekCase
}

// saveBranchIDMappings saves branch ID mappings for edge generation
func (g *ConditionNodeGenerator) saveBranchIDMappings(caseID, branchID string, level int) {
	// Save branch ID mapping for edge generation
	g.branchIDMapping[caseID] = branchID

	// Save mapping by level for backward compatibility
	levelKey := fmt.Sprintf("%d", level)
	g.branchIDMapping[levelKey] = branchID

	// Save special mappings for Coze sourcePortID format (dynamic mapping)
	if level == 999 {
		// Default branch maps to "false"
		g.branchIDMapping["false"] = branchID
		g.branchIDMapping["__default__"] = branchID
	} else if level == 1 {
		// First level maps to "true"
		g.branchIDMapping["true"] = branchID
	} else if level >= 2 {
		// Level 2 and above map to "true_X" format (Coze format)
		cozePortID := fmt.Sprintf("true_%d", level-1)
		g.branchIDMapping[cozePortID] = branchID
	}
}

// addDefaultBranch adds default branch (level 999) to condition cases
func (g *ConditionNodeGenerator) addDefaultBranch(iflytekCases []map[string]interface{}) []map[string]interface{} {
	defaultBranchID := "branch_one_of::" + generateUUID()
	defaultCase := g.createDefaultCase(defaultBranchID)

	// Save default branch ID mapping using the same logic as regular branches
	g.saveBranchIDMappings("__default__", defaultBranchID, 999)

	return append(iflytekCases, defaultCase)
}

// createDefaultCase creates default case structure
func (g *ConditionNodeGenerator) createDefaultCase(defaultBranchID string) map[string]interface{} {
	return map[string]interface{}{
		"level":           999,
		"logicalOperator": "and",
		"id":              defaultBranchID,
		"conditions":      []interface{}{}, // Empty conditions indicate default branch
	}
}

// generateConditionsWithInputIDs generates condition list using input ID mapping
func (g *ConditionNodeGenerator) generateConditionsWithInputIDs(conditions []models.Condition, inputIDMap map[string]string) []map[string]interface{} {
	var iflytekConditions []map[string]interface{}

	for _, condition := range conditions {
		iflytekCondition := g.processConditionWithInputIDs(condition, inputIDMap)
		if iflytekCondition != nil {
			iflytekConditions = append(iflytekConditions, iflytekCondition)
		}
	}

	return iflytekConditions
}

// processConditionWithInputIDs processes a single condition with input ID mapping
func (g *ConditionNodeGenerator) processConditionWithInputIDs(condition models.Condition, inputIDMap map[string]string) map[string]interface{} {
	if !g.isValidConditionSelector(condition.VariableSelector) {
		return nil
	}

	sourceNodeID, sourceOutput := g.extractVariableSelector(condition.VariableSelector)
	leftVarIndex, rightVarIndex := g.getInputIndices(sourceNodeID, sourceOutput, condition.Value, inputIDMap)

	if leftVarIndex == "" || rightVarIndex == "" {
		return nil
	}

	return g.createConditionStructure(leftVarIndex, rightVarIndex, condition.ComparisonOperator)
}

// isValidConditionSelector checks if variable selector is valid
func (g *ConditionNodeGenerator) isValidConditionSelector(variableSelector []string) bool {
	return len(variableSelector) >= 2
}

// extractVariableSelector extracts source node ID and output from variable selector
func (g *ConditionNodeGenerator) extractVariableSelector(variableSelector []string) (string, string) {
	return variableSelector[0], variableSelector[1]
}

// getInputIndices gets left and right variable indices from input ID mapping
func (g *ConditionNodeGenerator) getInputIndices(sourceNodeID, sourceOutput string, value interface{}, inputIDMap map[string]string) (string, string) {
	// Get variable input ID
	varKey := fmt.Sprintf("var_%s_%s", sourceNodeID, sourceOutput)
	leftVarIndex := inputIDMap[varKey]

	// Get literal input ID
	literalKey := fmt.Sprintf("literal_%v", value)
	rightVarIndex := inputIDMap[literalKey]

	return leftVarIndex, rightVarIndex
}

// createConditionStructure creates condition structure with input indices and operator
func (g *ConditionNodeGenerator) createConditionStructure(leftVarIndex, rightVarIndex, comparisonOperator string) map[string]interface{} {
	return map[string]interface{}{
		"leftVarIndex":          leftVarIndex,
		"rightVarIndex":         rightVarIndex,
		"compareOperator":       g.mapComparisonOperator(comparisonOperator),
		"compareOperatorErrMsg": "",
		"id":                    "",
	}
}

// mapLogicalOperator maps logical operators
func (g *ConditionNodeGenerator) mapLogicalOperator(op string) string {
	switch op {
	case "and":
		return "and"
	case "or":
		return "or"
	default:
		return "and"
	}
}

// mapComparisonOperator maps comparison operators
func (g *ConditionNodeGenerator) mapComparisonOperator(op string) string {
	// Try direct mapping first
	if mapped := g.getDirectOperatorMapping(op); mapped != "" {
		return mapped
	}

	// Check if it's already a valid operator
	if g.isValidOperator(op) {
		return op
	}

	// Return default safe fallback
	return "is"
}

// getDirectOperatorMapping returns direct operator mappings
func (g *ConditionNodeGenerator) getDirectOperatorMapping(op string) string {
	mappings := g.getOperatorMappings()
	return mappings[op]
}

// getOperatorMappings returns the operator mapping table
func (g *ConditionNodeGenerator) getOperatorMappings() map[string]string {
	return map[string]string{
		"contains":     "contains",
		"not contains": "not_contains",
		"equals":       "is",
		"not equals":   "is_not",
		"gt":           "gt",
		"gte":          "ge",
		"ge":           "ge",
		"lt":           "lt",
		"lte":          "le",
		"le":           "le",
		"starts_with":  "start_with",
		"start with":   "start_with",
		"ends_with":    "end_with",
		"end with":     "end_with",
		"is_empty":     "empty",
		"is_not_empty": "not_empty",
		"not empty":    "not_empty",
		"is_null":      "null",
		"null":         "null",
		"is_not_null":  "not_null",
		"not null":     "not_null",
		"eq":           "eq",
		"ne":           "ne",
	}
}

// isValidOperator checks if operator is already valid
func (g *ConditionNodeGenerator) isValidOperator(op string) bool {
	validOperators := g.getValidOperators()
	for _, validOp := range validOperators {
		if op == validOp {
			return true
		}
	}
	return false
}

// getValidOperators returns list of valid operators
func (g *ConditionNodeGenerator) getValidOperators() []string {
	return []string{
		"contains", "not_contains", "empty", "not_empty",
		"is", "is_not", "start_with", "end_with",
		"eq", "ne", "gt", "ge", "lt", "le", "null", "not_null",
	}
}

// generateBranchID generates branch ID
func (g *ConditionNodeGenerator) generateBranchID(caseID string) string {
	// Always generate branch_one_of format for consistency with iFlytek requirements
	// Generate ID that matches regex: ^branch_one_of::[0-9a-zA-Z-]+
	uuid := generateUUID()
	// Replace any characters that are not alphanumeric or hyphen
	// The UUID should already be compliant, but ensure it matches the regex
	return "branch_one_of::" + uuid
}

// generateReferences generates variable reference information
func (g *ConditionNodeGenerator) generateReferences(inputs []models.Input) []IFlytekReference {
	// Group inputs by source node
	nodeGroups := g.groupInputsBySourceNode(inputs)

	// Generate references for each node group
	return g.generateReferencesFromNodeGroups(nodeGroups)
}

// groupInputsBySourceNode groups inputs by their source node ID
func (g *ConditionNodeGenerator) groupInputsBySourceNode(inputs []models.Input) map[string][]models.Input {
	nodeGroups := make(map[string][]models.Input)

	for _, input := range inputs {
		if !g.isValidInputReference(input) {
			continue
		}

		mappedNodeID := g.getMappedNodeID(input.Reference.NodeID)
		nodeGroups[mappedNodeID] = append(nodeGroups[mappedNodeID], input)
	}

	return nodeGroups
}

// isValidInputReference checks if input has valid reference information
func (g *ConditionNodeGenerator) isValidInputReference(input models.Input) bool {
	return input.Reference != nil && input.Reference.NodeID != ""
}

// generateReferencesFromNodeGroups generates references from grouped inputs
func (g *ConditionNodeGenerator) generateReferencesFromNodeGroups(nodeGroups map[string][]models.Input) []IFlytekReference {
	var references []IFlytekReference

	for nodeID, nodeInputs := range nodeGroups {
		reference := g.createReferenceForNodeGroup(nodeID, nodeInputs)
		references = append(references, reference)
	}

	return references
}

// createReferenceForNodeGroup creates reference structure for a node group
func (g *ConditionNodeGenerator) createReferenceForNodeGroup(nodeID string, nodeInputs []models.Input) IFlytekReference {
	refDetails := g.createRefDetailsFromInputs(nodeID, nodeInputs)

	return IFlytekReference{
		Children: []IFlytekReference{
			{
				References: refDetails,
				Label:      "",
				Value:      "",
			},
		},
		Label:      g.getNodeDisplayLabel(nodeID),
		ParentNode: true,
		Value:      nodeID,
	}
}

// getNodeDisplayLabel gets the correct display label for a node
func (g *ConditionNodeGenerator) getNodeDisplayLabel(nodeID string) string {
	// First try to get from node title mapping
	if g.nodeTitleMapping != nil {
		if title, exists := g.nodeTitleMapping[nodeID]; exists {
			return title
		}
	}

	// If DSL is available, get the actual node title
	if g.unifiedDSL != nil {
		if sourceNode := g.findSourceNodeByMappedID(nodeID); sourceNode != nil {
			return sourceNode.Title
		}
	}

	// Fallback: use base generator's logic
	return g.determineLabelByID(nodeID, g.nodeTitleMapping)
}

// createRefDetailsFromInputs creates reference details from node inputs
func (g *ConditionNodeGenerator) createRefDetailsFromInputs(nodeID string, nodeInputs []models.Input) []IFlytekRefDetail {
	var refDetails []IFlytekRefDetail

	for _, input := range nodeInputs {
		refDetail := g.createRefDetailFromInput(nodeID, input)
		refDetails = append(refDetails, refDetail)
	}

	return refDetails
}

// createRefDetailFromInput creates a single reference detail from input
func (g *ConditionNodeGenerator) createRefDetailFromInput(nodeID string, input models.Input) IFlytekRefDetail {
	return IFlytekRefDetail{
		OriginID: nodeID,
		ID:       g.generateRefID(),
		Label:    input.Reference.OutputName,
		Type:     g.convertDataType(input.Reference.DataType),
		Value:    input.Reference.OutputName,
		FileType: "",
	}
}

// GetBranchIDMapping returns the branch ID mapping
func (g *ConditionNodeGenerator) GetBranchIDMapping() map[string]string {
	return g.branchIDMapping
}

// SetBranchIDMapping sets the branch ID mapping
func (g *ConditionNodeGenerator) SetBranchIDMapping(mapping map[string]string) {
	g.branchIDMapping = mapping
}

// inferDataTypeFromOutput infers data type from actual source node output in DSL
func (g *ConditionNodeGenerator) inferDataTypeFromOutput(sourceOutput, mappedNodeID string) string {
	// Use unified data type mapping system
	mapping := models.GetDefaultDataTypeMapping()

	// Try to get actual data type from source node if DSL is available
	if g.unifiedDSL != nil {
		if actualType := g.getActualOutputType(sourceOutput, mappedNodeID); actualType != "" {
			return actualType
		}
	}

	// Fallback to pattern-based inference if DSL not available or output not found
	unifiedType := g.getUnifiedTypeFromOutput(sourceOutput)
	return mapping.ToIFlytekType(unifiedType)
}

// getActualOutputType gets actual data type from source node in DSL
func (g *ConditionNodeGenerator) getActualOutputType(sourceOutput, mappedNodeID string) string {
	// Find the actual source node (may need to map ID back)
	sourceNode := g.findSourceNodeByMappedID(mappedNodeID)
	if sourceNode == nil {
		return ""
	}

	// Look for the output in the source node
	for _, output := range sourceNode.Outputs {
		if output.Name == sourceOutput {
			// Convert unified type to iFlytek type
			mapping := models.GetDefaultDataTypeMapping()
			return mapping.ToIFlytekType(output.Type)
		}
	}

	return ""
}

// findSourceNodeByMappedID finds the original source node by mapped ID
func (g *ConditionNodeGenerator) findSourceNodeByMappedID(mappedNodeID string) *models.Node {
	// First try direct lookup
	sourceNode := g.unifiedDSL.GetNodeByID(mappedNodeID)
	if sourceNode != nil {
		return sourceNode
	}

	// If not found, try reverse ID mapping lookup
	if g.idMapping != nil {
		for originalID, mappedID := range g.idMapping {
			if mappedID == mappedNodeID {
				return g.unifiedDSL.GetNodeByID(originalID)
			}
		}
	}

	return nil
}

// generateConditionReferences generates references for condition node from variable selectors
func (g *ConditionNodeGenerator) generateConditionReferences(node models.Node) []IFlytekReference {
	// Extract condition configuration
	condConfig, ok := common.AsConditionConfig(node.Config)
	if !ok || condConfig == nil {
		return []IFlytekReference{}
	}

	// Collect all variable references from conditions
	nodeGroups := g.collectVariableReferencesFromConditions(condConfig.Cases)

	// Generate references for each node group
	return g.generateReferencesFromConditionGroups(nodeGroups)
}

// collectVariableReferencesFromConditions collects variable references from condition cases
func (g *ConditionNodeGenerator) collectVariableReferencesFromConditions(cases []models.ConditionCase) map[string][]string {
	nodeGroups := make(map[string][]string)

	for _, caseItem := range cases {
		for _, condition := range caseItem.Conditions {
			if len(condition.VariableSelector) >= 2 {
				sourceNodeID := condition.VariableSelector[0]
				sourceOutput := condition.VariableSelector[1]

				// Map the node ID
				mappedNodeID := g.getMappedNodeID(sourceNodeID)

				// Add to node groups, avoid duplicates
				if !g.contains(nodeGroups[mappedNodeID], sourceOutput) {
					nodeGroups[mappedNodeID] = append(nodeGroups[mappedNodeID], sourceOutput)
				}
			}
		}
	}

	return nodeGroups
}

// contains checks if a string slice contains a specific string
func (g *ConditionNodeGenerator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateReferencesFromConditionGroups generates references from condition variable groups
func (g *ConditionNodeGenerator) generateReferencesFromConditionGroups(nodeGroups map[string][]string) []IFlytekReference {
	var references []IFlytekReference

	for nodeID, outputNames := range nodeGroups {
		reference := g.createReferenceForConditionGroup(nodeID, outputNames)
		references = append(references, reference)
	}

	return references
}

// createReferenceForConditionGroup creates reference structure for a condition node group
func (g *ConditionNodeGenerator) createReferenceForConditionGroup(nodeID string, outputNames []string) IFlytekReference {
	refDetails := g.createRefDetailsFromOutputNames(nodeID, outputNames)

	return IFlytekReference{
		Children: []IFlytekReference{
			{
				References: refDetails,
				Label:      "",
				Value:      "",
			},
		},
		Label:      g.getNodeDisplayLabel(nodeID),
		ParentNode: true,
		Value:      nodeID,
	}
}

// createRefDetailsFromOutputNames creates reference details from output names
func (g *ConditionNodeGenerator) createRefDetailsFromOutputNames(nodeID string, outputNames []string) []IFlytekRefDetail {
	var refDetails []IFlytekRefDetail

	for _, outputName := range outputNames {
		refDetail := g.createRefDetailFromOutputName(nodeID, outputName)
		refDetails = append(refDetails, refDetail)
	}

	return refDetails
}

// createRefDetailFromOutputName creates a single reference detail from output name
func (g *ConditionNodeGenerator) createRefDetailFromOutputName(nodeID, outputName string) IFlytekRefDetail {
	// Get actual data type from DSL if available
	dataType := g.getActualOutputDataType(nodeID, outputName)

	return IFlytekRefDetail{
		OriginID: nodeID,
		ID:       g.generateRefID(),
		Label:    outputName,
		Type:     dataType,
		Value:    outputName,
		FileType: "",
	}
}

// getActualOutputDataType gets actual data type for reference detail
func (g *ConditionNodeGenerator) getActualOutputDataType(nodeID, outputName string) string {
	if g.unifiedDSL != nil {
		if sourceNode := g.findSourceNodeByMappedID(nodeID); sourceNode != nil {
			for _, output := range sourceNode.Outputs {
				if output.Name == outputName {
					mapping := models.GetDefaultDataTypeMapping()
					return mapping.ToIFlytekType(output.Type)
				}
			}
		}
	}

	// Fallback to string type
	return "string"
}

// getUnifiedTypeFromOutput determines unified data type based on output name patterns
func (g *ConditionNodeGenerator) getUnifiedTypeFromOutput(outputName string) models.UnifiedDataType {
	// Common numeric field patterns
	numericPatterns := []string{
		"_year", "_month", "_day", "_hour", "_minute", "_second",
		"_count", "_number", "_num", "_index", "_id", "_size", "_length",
		"birth_year", "birth_month", "birth_day", "count", "index",
	}

	// Check for numeric patterns
	for _, pattern := range numericPatterns {
		if strings.Contains(strings.ToLower(outputName), pattern) {
			return models.DataTypeInteger
		}
	}

	// Boolean field patterns
	booleanPatterns := []string{
		"is_", "has_", "can_", "should_", "_enabled", "_active", "_valid",
		"enabled", "disabled", "active", "inactive", "valid", "invalid",
	}

	for _, pattern := range booleanPatterns {
		if strings.Contains(strings.ToLower(outputName), pattern) {
			return models.DataTypeBoolean
		}
	}

	// Float/decimal patterns
	floatPatterns := []string{
		"_rate", "_ratio", "_percentage", "_score", "_weight", "_temperature",
		"price", "amount", "value", "cost", "fee", "rate", "ratio",
	}

	for _, pattern := range floatPatterns {
		if strings.Contains(strings.ToLower(outputName), pattern) {
			return models.DataTypeFloat
		}
	}

	// Default to string type for unknown patterns
	return models.DataTypeString
}
