package generator

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
	"regexp"
	"strings"
)

// IterationNodeGenerator iteration node generator
type IterationNodeGenerator struct {
	*BaseNodeGenerator
	variableSelectorConverter *VariableSelectorConverter
	nodeMapping               map[string]models.Node
}

func NewIterationNodeGenerator() *IterationNodeGenerator {
	return &IterationNodeGenerator{
		BaseNodeGenerator:         NewBaseNodeGenerator(models.NodeTypeIteration),
		variableSelectorConverter: NewVariableSelectorConverter(),
		nodeMapping:               make(map[string]models.Node),
	}
}

// GenerateNode generates iteration node
func (g *IterationNodeGenerator) GenerateNode(node models.Node) (DifyNode, error) {
	if node.Type != models.NodeTypeIteration {
		return DifyNode{}, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeIteration)
	}

	// Generate base node structure
	difyNode := g.generateBaseNode(node)

	// Set iteration node specific data - directly set to data field, not wrapped in config
	g.setIterationDataFields(&difyNode.Data, node)

	// According to Dify standards, iteration nodes use specific dimensions
	difyNode.Height = 250
	difyNode.Width = 508
	difyNode.ZIndex = 1 // zIndex for iteration main node

	// Restore Dify-specific fields from platform-specific configuration
	if difyConfig := node.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyNode)
	}

	return difyNode, nil
}

// SetNodeMapping sets node mapping for variable selector converter and internal nodeMapping
func (g *IterationNodeGenerator) SetNodeMapping(nodes []models.Node) {
	g.variableSelectorConverter.SetNodeMapping(nodes)

	// Build internal node mapping for field name resolution
	if g.nodeMapping == nil {
		g.nodeMapping = make(map[string]models.Node)
	}
	for _, node := range nodes {
		g.nodeMapping[node.ID] = node
	}
}

// GenerateIterationNodes generates complete iteration node structure (including child nodes)
func (g *IterationNodeGenerator) GenerateIterationNodes(node models.Node) ([]DifyNode, error) {
	if node.Type != models.NodeTypeIteration {
		return nil, fmt.Errorf("unsupported node type: %s, expected: %s", node.Type, models.NodeTypeIteration)
	}

	// Generate main iteration node
	mainNode, err := g.GenerateNode(node)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main iteration node: %w", err)
	}

	// Generate start node and update main node
	startNode := g.createIterationStartNode(node, mainNode)
	mainNode.Data.StartNodeID = startNode.ID

	// Generate internal nodes
	internalNodes, err := g.generateInternalNodes(node, mainNode)
	if err != nil {
		return nil, err
	}

	// Configure output selector
	g.configureIterationOutputSelector(node, &mainNode, internalNodes)

	// Assemble all nodes
	nodes := []DifyNode{mainNode, startNode}
	nodes = append(nodes, internalNodes...)

	return nodes, nil
}

// createIterationStartNode creates and positions start node
func (g *IterationNodeGenerator) createIterationStartNode(node models.Node, mainNode DifyNode) DifyNode {
	startNode := g.generateIterationStartNode(node, mainNode.ID)
	startNode.PositionAbsolute = DifyPosition{
		X: mainNode.PositionAbsolute.X + startNode.Position.X,
		Y: mainNode.PositionAbsolute.Y + startNode.Position.Y,
	}
	return startNode
}

// generateInternalNodes generates all internal processing nodes
func (g *IterationNodeGenerator) generateInternalNodes(node models.Node, mainNode DifyNode) ([]DifyNode, error) {
	iterConfig, ok := node.Config.(*models.IterationConfig)
	if !ok || len(iterConfig.SubWorkflow.Nodes) == 0 {
		return []DifyNode{}, nil
	}

	g.setupVariableContext(iterConfig, mainNode.ID)
	return g.processSubWorkflowNodes(iterConfig, mainNode)
}

// setupVariableContext sets up variable mapping context
func (g *IterationNodeGenerator) setupVariableContext(iterConfig *models.IterationConfig, mainNodeID string) {
	g.variableSelectorConverter.SetNodeMapping(iterConfig.SubWorkflow.Nodes)
	g.variableSelectorConverter.SetIterationContext(mainNodeID)
}

// processSubWorkflowNodes processes all sub-workflow nodes
func (g *IterationNodeGenerator) processSubWorkflowNodes(iterConfig *models.IterationConfig, mainNode DifyNode) ([]DifyNode, error) {
	var internalNodes []DifyNode

	for _, subNode := range iterConfig.SubWorkflow.Nodes {
		if g.shouldSkipNode(subNode) {
			continue
		}

		internalNode, err := g.createPositionedInternalNode(subNode, mainNode, iterConfig.SubWorkflow.Nodes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate internal node %s: %w", subNode.ID, err)
		}
		internalNodes = append(internalNodes, internalNode)
	}

	return internalNodes, nil
}

// shouldSkipNode checks if node should be skipped during generation
func (g *IterationNodeGenerator) shouldSkipNode(subNode models.Node) bool {
	return subNode.Type == models.NodeTypeStart || subNode.Type == models.NodeTypeEnd
}

// createPositionedInternalNode creates and positions internal node
func (g *IterationNodeGenerator) createPositionedInternalNode(subNode models.Node, mainNode DifyNode, allNodes []models.Node) (DifyNode, error) {
	internalNode, err := g.generateInternalNode(subNode, mainNode.ID, allNodes)
	if err != nil {
		return DifyNode{}, err
	}

	internalNode.PositionAbsolute = DifyPosition{
		X: mainNode.PositionAbsolute.X + internalNode.Position.X,
		Y: mainNode.PositionAbsolute.Y + internalNode.Position.Y,
	}

	return internalNode, nil
}

// configureIterationOutputSelector configures the iteration output selector
func (g *IterationNodeGenerator) configureIterationOutputSelector(node models.Node, mainNode *DifyNode, internalNodes []DifyNode) {
	if len(internalNodes) == 0 {
		return
	}

	if outputSelector := g.findIterationOutputSelectorWithMapping(node, internalNodes); len(outputSelector) == 2 {
		mainNode.Data.OutputSelector = outputSelector
	} else {
		g.setFallbackOutputSelector(mainNode, internalNodes, node)
	}
}

// setFallbackOutputSelector sets fallback output selector using last internal node
func (g *IterationNodeGenerator) setFallbackOutputSelector(mainNode *DifyNode, internalNodes []DifyNode, node models.Node) {
	lastInternalNode := internalNodes[len(internalNodes)-1]
	outputName := g.getNodeOutputName(&lastInternalNode, node)
	mainNode.Data.OutputSelector = []string{lastInternalNode.ID, outputName}

}

// generateIterationStartNode generates iteration start node
func (g *IterationNodeGenerator) generateIterationStartNode(parentNode models.Node, parentID string) DifyNode {
	startNodeID := fmt.Sprintf("%sstart", parentID)

	// Set special fields for iteration start node
	draggable := false
	selectable := false

	return DifyNode{
		ID:   startNodeID,
		Type: "custom-iteration-start",
		Data: DifyNodeData{
			Type:          "iteration-start",
			Title:         "",
			Desc:          "Iterator start node",
			Selected:      false,
			IsInIteration: true,
			// Note: don't set IsParallel field, as iteration start node doesn't need this field
		},
		Height:           48,
		Width:            44,
		Position:         DifyPosition{X: 60, Y: 101},
		PositionAbsolute: DifyPosition{X: 60, Y: 101}, // Set as relative position here, actual should calculate absolute position
		SourcePosition:   "right",
		TargetPosition:   "left",
		ParentID:         parentID,
		Draggable:        &draggable,  // Set according to standard examples
		Selectable:       &selectable, // Set according to standard examples
		ZIndex:           1002,
	}
}

// generateInternalNode generates iteration internal processing nodes
func (g *IterationNodeGenerator) generateInternalNode(subNode models.Node, parentID string, subWorkflowNodes []models.Node) (DifyNode, error) {
	baseNode, err := g.createBaseInternalNode(subNode, parentID, subWorkflowNodes)
	if err != nil {
		return DifyNode{}, err
	}

	g.configureInternalNodeProperties(&baseNode, parentID)
	g.setInternalNodeLayout(&baseNode)
	g.applyNodeTypeSpecificConfiguration(&baseNode, subNode, parentID)

	return baseNode, nil
}

// createBaseInternalNode creates the base internal node using factory
func (g *IterationNodeGenerator) createBaseInternalNode(subNode models.Node, parentID string, subWorkflowNodes []models.Node) (DifyNode, error) {
	factory := g.configureNodeFactory(parentID, subWorkflowNodes)

	baseNode, err := factory.GenerateNode(subNode)
	if err != nil {
		return DifyNode{}, fmt.Errorf("failed to generate base internal node: %w", err)
	}

	return baseNode, nil
}

// configureNodeFactory configures the node factory for iteration context
func (g *IterationNodeGenerator) configureNodeFactory(parentID string, subWorkflowNodes []models.Node) *NodeGeneratorFactory {
	factory := NewNodeGeneratorFactory()
	factory.SetNodeMapping(subWorkflowNodes)

	// Set iteration context for condition nodes
	if condGen, err := factory.GetGenerator(models.NodeTypeCondition); err == nil {
		if conditionGen, ok := condGen.(*ConditionNodeGenerator); ok {
			conditionGen.variableSelectorConverter.SetIterationContext(parentID)
		}
	}

	return factory
}

// configureInternalNodeProperties configures basic properties for iteration internal nodes
func (g *IterationNodeGenerator) configureInternalNodeProperties(baseNode *DifyNode, parentID string) {
	baseNode.ParentID = parentID
	baseNode.Data.IsInIteration = true
	baseNode.Data.IterationID = parentID
	isInLoop := false
	baseNode.Data.IsInLoop = &isInLoop
	baseNode.Data.Selected = false
	baseNode.ZIndex = 1002
}

// setInternalNodeLayout sets position and dimensions for internal nodes
func (g *IterationNodeGenerator) setInternalNodeLayout(baseNode *DifyNode) {
	position, dimensions := g.getNodeLayoutConfig(baseNode.Data.Type)
	baseNode.Position = position
	baseNode.Width = float64(dimensions.Width)
	baseNode.Height = float64(dimensions.Height)
	baseNode.PositionAbsolute = DifyPosition{X: position.X, Y: position.Y}
}

// getNodeLayoutConfig returns position and dimensions for different node types
func (g *IterationNodeGenerator) getNodeLayoutConfig(nodeType string) (DifyPosition, struct{ Width, Height int }) {
	baseX, baseY := 204.0, 60.0

	switch nodeType {
	case "question-classifier":
		return DifyPosition{X: baseX, Y: baseY}, struct{ Width, Height int }{148, 44}
	case "llm":
		return DifyPosition{X: baseX + 200, Y: baseY}, struct{ Width, Height int }{148, 44}
	case "code":
		return DifyPosition{X: baseX + 400, Y: baseY}, struct{ Width, Height int }{244, 82}
	case "if-else":
		return DifyPosition{X: baseX + 200, Y: baseY + 100}, struct{ Width, Height int }{132, 44}
	case "iteration-start":
		return DifyPosition{X: 60, Y: 101}, struct{ Width, Height int }{44, 48}
	default:
		return DifyPosition{X: baseX, Y: baseY + 200}, struct{ Width, Height int }{100, 44}
	}
}

// applyNodeTypeSpecificConfiguration applies node type specific configuration
func (g *IterationNodeGenerator) applyNodeTypeSpecificConfiguration(baseNode *DifyNode, subNode models.Node, parentID string) {
	switch baseNode.Data.Type {
	case "code":
		g.configureIterationCodeNode(baseNode, subNode, parentID)
	case "llm":
		g.configureIterationLLMNode(baseNode, subNode, parentID)
	case "question-classifier":
		g.configureIterationClassifierNode(baseNode, subNode, parentID)
	case "if-else":
		g.configureIterationConditionNode(baseNode, subNode, parentID)
	default:
		g.configureIterationBasicNode(baseNode, subNode, parentID)
	}
}

// configureIterationCodeNode configures iteration internal code nodes
func (g *IterationNodeGenerator) configureIterationCodeNode(difyNode *DifyNode, originalNode models.Node, parentID string) {
	// Set all input variables for code node
	variables := []map[string]interface{}{}

	// Process all input variables
	for _, input := range originalNode.Inputs {
		var valueSelector []string
		var valueType string

		// If input has reference, use referenced node and output
		if input.Reference != nil && input.Reference.NodeID != "" {
			// Infer correct output field name and node ID based on referenced node type
			actualNodeID, outputFieldName := g.inferNodeIDAndOutputFieldName(input.Reference.NodeID, input.Reference.OutputName, parentID)
			valueSelector = []string{actualNodeID, outputFieldName}
			valueType = g.mapToValueType(string(input.Type))
		} else {
			// If no reference, default to referencing iterator's item
			valueSelector = []string{parentID, "item"}
			valueType = "string"
		}

		variable := map[string]interface{}{
			"value_selector": valueSelector,
			"value_type":     valueType,
			"variable":       input.Name,
		}
		variables = append(variables, variable)
	}

	// If no inputs defined, set default item reference
	if len(variables) == 0 {
		variables = []map[string]interface{}{
			{
				"value_selector": []string{parentID, "item"},
				"value_type":     "string",
				"variable":       "content",
			},
		}
	}

	difyNode.Data.Variables = variables

	// Set outputs field (Dify format)
	if len(originalNode.Outputs) > 0 {
		outputs := make(map[string]interface{})
		for _, output := range originalNode.Outputs {
			outputs[output.Name] = map[string]interface{}{
				"children": nil,
				"type":     g.mapToDifyOutputType(string(output.Type)),
			}
		}
		difyNode.Data.Outputs = outputs
	} else {
		// If no outputs defined, use appropriate default output name based on node type
		defaultOutputName := "result" // Standard output name for code nodes
		if difyNode.Data.Type == "llm" {
			defaultOutputName = "text"
		}

		difyNode.Data.Outputs = map[string]interface{}{
			defaultOutputName: map[string]interface{}{
				"children": nil,
				"type":     g.mapToDifyOutputType("string"), // Infer based on node type
			},
		}
	}
}

// configureIterationLLMNode configures iteration internal LLM nodes
func (g *IterationNodeGenerator) configureIterationLLMNode(difyNode *DifyNode, originalNode models.Node, parentID string) {
	g.logLLMNodeConfiguration(originalNode, parentID)

	variables := g.buildLLMNodeVariables(originalNode, parentID)
	g.configureLLMNodeVariables(difyNode, variables, parentID)
	g.fixLLMPromptTemplateReferences(difyNode, parentID)
}

// logLLMNodeConfiguration logs LLM node configuration details
func (g *IterationNodeGenerator) logLLMNodeConfiguration(originalNode models.Node, parentID string) {
}

// buildLLMNodeVariables builds variables from input nodes
func (g *IterationNodeGenerator) buildLLMNodeVariables(originalNode models.Node, parentID string) []map[string]interface{} {
	variables := []map[string]interface{}{}

	for _, input := range originalNode.Inputs {
		variable := g.createLLMVariable(input, parentID)
		variables = append(variables, variable)
	}

	return variables
}

// createLLMVariable creates a single LLM variable from input
func (g *IterationNodeGenerator) createLLMVariable(input models.Input, parentID string) map[string]interface{} {
	valueSelector, valueType := g.determineLLMVariableSelector(input, parentID)

	return map[string]interface{}{
		"value_selector": valueSelector,
		"value_type":     valueType,
		"variable":       input.Name,
	}
}

// determineLLMVariableSelector determines value selector and type for LLM variable
func (g *IterationNodeGenerator) determineLLMVariableSelector(input models.Input, parentID string) ([]string, string) {
	if input.Reference != nil && input.Reference.NodeID != "" {
		actualNodeID, outputFieldName := g.inferNodeIDAndOutputFieldName(input.Reference.NodeID, input.Reference.OutputName, parentID)
		return []string{actualNodeID, outputFieldName}, g.mapToValueType(string(input.Type))
	}
	return []string{parentID, "item"}, "string"
}

// configureLLMNodeVariables configures variables and context for LLM node
func (g *IterationNodeGenerator) configureLLMNodeVariables(difyNode *DifyNode, variables []map[string]interface{}, parentID string) {
	// LLM nodes in Dify should NOT have variables configuration
	// Variables are referenced through prompt template only, not through variables field
	difyNode.Data.Variables = []interface{}{} // Always empty for LLM nodes

	// LLM nodes should have disabled context with empty variable_selector
	g.setCorrectLLMContext(difyNode)
}

// setCorrectLLMContext sets correct context configuration for LLM nodes
func (g *IterationNodeGenerator) setCorrectLLMContext(difyNode *DifyNode) {
	// Ensure context exists
	if difyNode.Data.Context == nil {
		difyNode.Data.Context = make(map[string]interface{})
	}

	difyNode.Data.Context["enabled"] = false
	difyNode.Data.Context["variable_selector"] = []interface{}{}
}

// fixLLMPromptTemplateReferences fixes variable references in prompt template
func (g *IterationNodeGenerator) fixLLMPromptTemplateReferences(difyNode *DifyNode, parentID string) {
	promptTemplate := difyNode.Data.PromptTemplate
	if len(promptTemplate) == 0 {
		return
	}

	for i, template := range promptTemplate {
		if text, ok := template["text"].(string); ok {
			text = g.fixIterationVariableReferences(text, parentID)
			promptTemplate[i]["text"] = text
		}
	}
	difyNode.Data.PromptTemplate = promptTemplate
}

// configureIterationClassifierNode configures iteration internal classifier nodes
func (g *IterationNodeGenerator) configureIterationClassifierNode(difyNode *DifyNode, originalNode models.Node, parentID string) {
	// Use unified variable mapping system to process all input variables
	variables := []map[string]interface{}{}

	// Process all input variables - use same logic as code nodes
	for _, input := range originalNode.Inputs {
		var valueSelector []string
		var valueType string

		if input.Reference != nil && input.Reference.NodeID != "" {
			// Use unified variable reference mapping function
			actualNodeID, outputFieldName := g.inferNodeIDAndOutputFieldName(input.Reference.NodeID, input.Reference.OutputName, parentID)
			valueSelector = []string{actualNodeID, outputFieldName}
			valueType = g.mapToValueType(string(input.Type))
		} else {
			// Default to referencing iterator's item
			valueSelector = []string{parentID, "item"}
			valueType = "string"
		}

		variable := map[string]interface{}{
			"value_selector": valueSelector,
			"value_type":     valueType,
			"variable":       input.Name,
		}
		variables = append(variables, variable)
	}

	// Set variables array
	if len(variables) > 0 {
		difyNode.Data.Variables = variables

		// Set query_variable_selector to first input's reference
		firstVar := variables[0]
		if selector, ok := firstVar["value_selector"].([]string); ok && len(selector) == 2 {
			difyNode.Data.QueryVariableSelector = selector
		} else {
			// Fallback to default value
			difyNode.Data.QueryVariableSelector = []string{parentID, "item"}
		}
	} else {
		// If no input variables, default to referencing iterator's item
		difyNode.Data.QueryVariableSelector = []string{parentID, "item"}
	}

	// Fix variable references in instruction
	if instruction := difyNode.Data.Instruction; instruction != "" {
		instruction = g.fixIterationVariableReferences(instruction, parentID)
		difyNode.Data.Instruction = instruction
	}
}

// configureIterationConditionNode configures iteration internal conditional branch nodes
func (g *IterationNodeGenerator) configureIterationConditionNode(difyNode *DifyNode, originalNode models.Node, parentID string) {
	g.processConditionInputVariables(difyNode, originalNode, parentID)
	g.fixConditionCaseReferences(difyNode, parentID)
	g.fixConditionLogicalOperator(difyNode, parentID)
}

// processConditionInputVariables processes input variables for condition nodes
func (g *IterationNodeGenerator) processConditionInputVariables(difyNode *DifyNode, originalNode models.Node, parentID string) {
	variables := g.buildConditionVariables(originalNode, parentID)
	if len(variables) > 0 {
		difyNode.Data.Variables = variables
	}
}

// buildConditionVariables builds variables array from input definitions
func (g *IterationNodeGenerator) buildConditionVariables(originalNode models.Node, parentID string) []map[string]interface{} {
	var variables []map[string]interface{}

	for _, input := range originalNode.Inputs {
		variable := g.createConditionVariable(input, parentID)
		variables = append(variables, variable)
	}

	return variables
}

// createConditionVariable creates a single condition variable
func (g *IterationNodeGenerator) createConditionVariable(input models.Input, parentID string) map[string]interface{} {
	valueSelector, valueType := g.resolveConditionVariableReference(input, parentID)

	return map[string]interface{}{
		"value_selector": valueSelector,
		"value_type":     valueType,
		"variable":       input.Name,
	}
}

// resolveConditionVariableReference resolves variable reference for condition
func (g *IterationNodeGenerator) resolveConditionVariableReference(input models.Input, parentID string) ([]string, string) {
	if input.Reference != nil && input.Reference.NodeID != "" {
		actualNodeID, outputFieldName := g.inferNodeIDAndOutputFieldName(input.Reference.NodeID, input.Reference.OutputName, parentID)
		return []string{actualNodeID, outputFieldName}, g.mapToValueType(string(input.Type))
	}

	// Default to referencing iterator's item
	return []string{parentID, "item"}, "string"
}

// fixConditionCaseReferences fixes variable selector references in conditional branches
func (g *IterationNodeGenerator) fixConditionCaseReferences(difyNode *DifyNode, parentID string) {
	if len(difyNode.Data.Cases) == 0 {
		return
	}

	for _, caseMap := range difyNode.Data.Cases {
		g.fixCaseConditions(caseMap, parentID)
	}
}

// fixCaseConditions fixes conditions within a single case
func (g *IterationNodeGenerator) fixCaseConditions(caseMap map[string]interface{}, parentID string) {
	conditions, ok := caseMap["conditions"].([]map[string]interface{})
	if !ok {
		return
	}

	for _, condition := range conditions {
		g.fixConditionVariableSelector(condition, parentID)
	}
}

// fixConditionVariableSelector fixes variable selector in a single condition
func (g *IterationNodeGenerator) fixConditionVariableSelector(condition map[string]interface{}, parentID string) {
	variableSelector, ok := condition["variable_selector"].([]interface{})
	if !ok || len(variableSelector) < 2 {
		return
	}

	nodeID, nodeOk := variableSelector[0].(string)
	field, fieldOk := variableSelector[1].(string)
	if !nodeOk || !fieldOk {
		return
	}

	actualNodeID, outputFieldName := g.inferNodeIDAndOutputFieldName(nodeID, field, parentID)
	variableSelector[0] = actualNodeID
	variableSelector[1] = outputFieldName
}

// fixConditionLogicalOperator fixes variable references in logical operators
func (g *IterationNodeGenerator) fixConditionLogicalOperator(difyNode *DifyNode, parentID string) {
	if logicalOperator, ok := difyNode.Data.Config["logical_operator"].(string); ok {
		difyNode.Data.Config["logical_operator"] = g.fixIterationVariableReferences(logicalOperator, parentID)
	}
}

// configureIterationBasicNode configures other types of iteration internal nodes
func (g *IterationNodeGenerator) configureIterationBasicNode(difyNode *DifyNode, originalNode models.Node, parentID string) {
	// Use unified variable mapping system to process all input variables
	variables := []map[string]interface{}{}

	// Process all input variables - use same logic as other nodes
	for _, input := range originalNode.Inputs {
		var valueSelector []string
		var valueType string

		if input.Reference != nil && input.Reference.NodeID != "" {
			// Use unified variable reference mapping function
			actualNodeID, outputFieldName := g.inferNodeIDAndOutputFieldName(input.Reference.NodeID, input.Reference.OutputName, parentID)
			valueSelector = []string{actualNodeID, outputFieldName}
			valueType = g.mapToValueType(string(input.Type))
		} else {
			// Default to referencing iterator's item
			valueSelector = []string{parentID, "item"}
			valueType = "string"
		}

		variable := map[string]interface{}{
			"value_selector": valueSelector,
			"value_type":     valueType,
			"variable":       input.Name,
		}
		variables = append(variables, variable)
	}

	// Set variables array (if there are input variables)
	if len(variables) > 0 {
		difyNode.Data.Variables = variables
	}
}

// mapToDifyOutputType maps output types to Dify format
func (g *IterationNodeGenerator) mapToDifyOutputType(outputType string) string {
	// Use unified mapping system, supports alias processing
	mapping := models.GetDefaultDataTypeMapping()
	return mapping.MapToDifyTypeWithAliases(outputType)
}

// generateIterationConfig generates iteration configuration

// setIterationDataFields sets iteration node data fields
func (g *IterationNodeGenerator) setIterationDataFields(data *DifyNodeData, node models.Node) {
	g.setIterationBasicFields(data)
	g.setIterationDimensions(data)
	g.processIterationConfig(data, node)
	g.inferIterationFromInputs(data, node)
	g.inferIterationFromOutputs(data, node)
	g.setIterationStartNodeID(data, node)
}

// setIterationBasicFields sets basic configuration for iteration nodes
func (g *IterationNodeGenerator) setIterationBasicFields(data *DifyNodeData) {
	data.ErrorHandleMode = "terminated"
	isParallel := false
	data.IsParallel = &isParallel
	data.ParallelNums = 10
	data.IteratorInputType = "array[string]"
	data.OutputType = "array[string]"
	data.Selected = false
}

// setIterationDimensions sets height and width according to Dify standards
func (g *IterationNodeGenerator) setIterationDimensions(data *DifyNodeData) {
	data.Height = 202
	data.Width = 508
}

// processIterationConfig processes iteration configuration from node config
func (g *IterationNodeGenerator) processIterationConfig(data *DifyNodeData, node models.Node) {
	iterConfig, ok := node.Config.(*models.IterationConfig)
	if !ok {
		return
	}

	g.setIterationInputType(data, iterConfig)
	g.setIterationExecutionConfig(data)
	g.setIterationSelector(data, iterConfig)
}

// setIterationInputType converts unified DSL types to Dify format
func (g *IterationNodeGenerator) setIterationInputType(data *DifyNodeData, iterConfig *models.IterationConfig) {
	if iterConfig.Iterator.InputType == "" {
		return
	}

	switch iterConfig.Iterator.InputType {
	case "array", "array[string]":
		data.IteratorInputType = "array[string]"
	case "array[object]":
		data.IteratorInputType = "array[object]"
	default:
		data.IteratorInputType = "array[string]"
	}
}

// setIterationExecutionConfig sets execution configuration using Dify standard values
func (g *IterationNodeGenerator) setIterationExecutionConfig(data *DifyNodeData) {
	isParallel := false
	data.IsParallel = &isParallel
	data.ParallelNums = 10
	data.ErrorHandleMode = "terminated"
}

// setIterationSelector sets iterator selector from configuration
func (g *IterationNodeGenerator) setIterationSelector(data *DifyNodeData, iterConfig *models.IterationConfig) {
	if iterConfig.Iterator.SourceNode != "" && iterConfig.Iterator.SourceOutput != "" {
		data.IteratorSelector = []string{iterConfig.Iterator.SourceNode, iterConfig.Iterator.SourceOutput}
	}
}

// inferIterationFromInputs infers iterator selector and type from inputs
func (g *IterationNodeGenerator) inferIterationFromInputs(data *DifyNodeData, node models.Node) {
	if len(node.Inputs) == 0 || len(data.IteratorSelector) > 0 {
		return
	}

	firstInput := node.Inputs[0]
	if firstInput.Reference == nil || firstInput.Reference.NodeID == "" {
		return
	}

	sourceOutput := firstInput.Reference.OutputName
	if sourceOutput == "" {
		sourceOutput = firstInput.Name
	}
	data.IteratorSelector = []string{firstInput.Reference.NodeID, sourceOutput}

	mapping := models.GetDefaultDataTypeMapping()
	data.IteratorInputType = mapping.ToDifyType(firstInput.Type)
}

// inferIterationFromOutputs infers output type from outputs
func (g *IterationNodeGenerator) inferIterationFromOutputs(data *DifyNodeData, node models.Node) {
	if len(node.Outputs) == 0 {
		return
	}

	firstOutput := node.Outputs[0]
	mapping := models.GetDefaultDataTypeMapping()
	data.OutputType = mapping.ToDifyType(firstOutput.Type)
}

// setIterationStartNodeID sets default start node ID
func (g *IterationNodeGenerator) setIterationStartNodeID(data *DifyNodeData, node models.Node) {
	iterConfig, ok := node.Config.(*models.IterationConfig)
	if ok && iterConfig.SubWorkflow.StartNodeID != "" {
		data.StartNodeID = iterConfig.SubWorkflow.StartNodeID
		return
	}

	data.StartNodeID = node.ID + "start"
}

// restoreDifyPlatformConfig restores Dify platform-specific configuration
func (g *IterationNodeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, node *DifyNode) {
	g.ensureIterationNodeConfig(node)

	g.restoreIterationDirectConfigs(config, node)
	g.restoreIterationArrayConfigs(config, node)
	g.restoreIterationNodeMetadata(config, node)
}

func (g *IterationNodeGenerator) ensureIterationNodeConfig(node *DifyNode) {
	if node.Data.Config == nil {
		node.Data.Config = make(map[string]interface{})
	}
}

func (g *IterationNodeGenerator) restoreIterationDirectConfigs(config map[string]interface{}, node *DifyNode) {
	directConfigs := map[string]interface{}{
		"error_handle_mode":   "",
		"height":              0,
		"width":               0,
		"is_parallel":         false,
		"iterator_input_type": "",
		"output_type":         "",
	}

	for key := range directConfigs {
		if value, exists := config[key]; exists {
			node.Data.Config[key] = value
		}
	}
}

func (g *IterationNodeGenerator) restoreIterationArrayConfigs(config map[string]interface{}, node *DifyNode) {
	g.restoreIteratorSelector(config, node)
}

func (g *IterationNodeGenerator) restoreIteratorSelector(config map[string]interface{}, node *DifyNode) {
	if iteratorSelector, exists := config["iterator_selector"].([]interface{}); exists {
		selector := g.convertInterfaceArrayToStringArray(iteratorSelector)
		node.Data.Config["iterator_selector"] = selector
	}
}

func (g *IterationNodeGenerator) convertInterfaceArrayToStringArray(items []interface{}) []string {
	selector := make([]string, 0, len(items))
	for _, item := range items {
		if str, ok := item.(string); ok {
			selector = append(selector, str)
		}
	}
	return selector
}

func (g *IterationNodeGenerator) restoreIterationNodeMetadata(config map[string]interface{}, node *DifyNode) {
	if desc, ok := config["desc"].(string); ok {
		node.Data.Desc = desc
	}
	if title, ok := config["title"].(string); ok {
		node.Data.Title = title
	}
}

// Based on iFlytek SparkAgent end node input references and generated Dify nodes, intelligently infer output selector
func (g *IterationNodeGenerator) findIterationOutputSelectorWithMapping(node models.Node, generatedNodes []DifyNode) []string {
	if iterConfigPtr, ok := node.Config.(*models.IterationConfig); ok && len(iterConfigPtr.SubWorkflow.Nodes) > 0 {
		// Build mapping from original node IDs to generated node IDs
		nodeMapping := g.buildNodeMapping(iterConfigPtr.SubWorkflow.Nodes, generatedNodes)

		// Strategy 1: Intelligent output node inference - analyze from iFlytek end node
		if originalNodeID, outputName := g.inferOutputFromEndNode(iterConfigPtr.SubWorkflow.Nodes); originalNodeID != "" {
			if difyNodeID, exists := nodeMapping[originalNodeID]; exists {
				// Need to ensure correct output name mapping is used
				mappedOutputName := g.mapNodeOutputNameToDify(originalNodeID, outputName, generatedNodes)
				// For code nodes, ensure correct output name is used (usually last result output, not intermediate calculation output)
				if mappedOutputName == "len" {
					mappedOutputName = "result"
				}
				return []string{difyNodeID, mappedOutputName}
			}
		}

		// Strategy 2: Select best output node based on node priority
		if originalNodeID, outputName := g.selectBestOutputNode(iterConfigPtr.SubWorkflow.Nodes); originalNodeID != "" {
			if difyNodeID, exists := nodeMapping[originalNodeID]; exists {
				mappedOutputName := g.mapNodeOutputNameToDify(originalNodeID, outputName, generatedNodes)
				// For code nodes, ensure correct output name is used (usually last result output, not intermediate calculation output)
				if mappedOutputName == "len" {
					mappedOutputName = "result"
				}
				return []string{difyNodeID, mappedOutputName}
			}
		}

		// Strategy 3: Fallback strategy - use last generated node
		if len(generatedNodes) > 0 {
			lastNode := generatedNodes[len(generatedNodes)-1]
			outputName := g.inferGeneratedNodeOutputName(lastNode)
			return []string{lastNode.ID, outputName}
		}
	}

	return []string{}
}

// buildNodeMapping builds mapping from original node IDs to generated node IDs
func (g *IterationNodeGenerator) buildNodeMapping(originalNodes []models.Node, generatedNodes []DifyNode) map[string]string {
	mapping := make(map[string]string)

	// Filter out start and end nodes, only map processing nodes
	var processNodes []models.Node
	for _, node := range originalNodes {
		if node.Type != models.NodeTypeStart && node.Type != models.NodeTypeEnd {
			processNodes = append(processNodes, node)
		}
	}

	// Map in order (assume generation order matches original order)
	for i, generatedNode := range generatedNodes {
		if i < len(processNodes) {
			mapping[processNodes[i].ID] = generatedNode.ID
		}
	}

	return mapping
}

// inferGeneratedNodeOutputName infers output name of generated node
func (g *IterationNodeGenerator) inferGeneratedNodeOutputName(node DifyNode) string {
	// Get output name from node's outputs configuration
	if outputs, ok := node.Data.Outputs.(map[string]interface{}); ok {
		for outputName := range outputs {
			return outputName // Return first output name
		}
	}

	// Infer default output name based on node type
	switch node.Data.Type {
	case "code":
		return "result"
	case "llm":
		return "text"
	case "question-classifier":
		return "class_name"
	default:
		return "output"
	}
}

// Based on iFlytek SparkAgent end node input references, intelligently infer which node Dify should use as output
func (g *IterationNodeGenerator) findIterationOutputSelector(node models.Node, defaultOutputName string) []string {
	if iterConfigPtr, ok := node.Config.(*models.IterationConfig); ok && len(iterConfigPtr.SubWorkflow.Nodes) > 0 {
		// Strategy 1: Intelligent output node inference - analyze from iFlytek end node
		if nodeID, outputName := g.inferOutputFromEndNode(iterConfigPtr.SubWorkflow.Nodes); nodeID != "" {
			return []string{nodeID, outputName}
		}

		// Strategy 2: Select best output node based on node priority
		if nodeID, outputName := g.selectBestOutputNode(iterConfigPtr.SubWorkflow.Nodes); nodeID != "" {
			return []string{nodeID, outputName}
		}

		// Strategy 3: Fallback to default strategy
		return g.fallbackOutputSelector(iterConfigPtr.SubWorkflow.Nodes, defaultOutputName)
	}

	// If no suitable child nodes found, return empty slice
	return []string{}
}

// inferOutputFromEndNode infers Dify output from iFlytek end node
func (g *IterationNodeGenerator) inferOutputFromEndNode(subNodes []models.Node) (string, string) {
	// Find end node
	var endNode *models.Node
	for i, subNode := range subNodes {
		if subNode.Type == models.NodeTypeEnd {
			endNode = &subNodes[i]
			break
		}
	}

	if endNode == nil || len(endNode.Inputs) == 0 {
		return "", ""
	}

	// Get source node information from end node's first input reference
	firstInput := endNode.Inputs[0]
	if firstInput.Reference == nil || firstInput.Reference.NodeID == "" {
		return "", ""
	}

	sourceNodeID := firstInput.Reference.NodeID
	originalOutputName := firstInput.Reference.OutputName

	// Find referenced source node and map to Dify output field name
	for _, subNode := range subNodes {
		if subNode.ID == sourceNodeID {
			difyOutputName := g.mapToDifyOutputName(subNode.Type, originalOutputName)
			return sourceNodeID, difyOutputName
		}
	}

	return "", ""
}

// selectBestOutputNode select best output node based on node priority
func (g *IterationNodeGenerator) selectBestOutputNode(subNodes []models.Node) (string, string) {
	// Define node priority (smaller number means higher priority)
	nodePriority := map[models.NodeType]int{
		models.NodeTypeCode:       1, // Code nodes are usually final processing nodes
		models.NodeTypeLLM:        2, // LLM nodes
		models.NodeTypeClassifier: 3, // Classifier nodes
		models.NodeTypeCondition:  4, // Condition nodes
	}

	var bestNode *models.Node
	bestPriority := 999

	// Iterate to find node with highest priority
	for i, subNode := range subNodes {
		if subNode.Type == models.NodeTypeStart || subNode.Type == models.NodeTypeEnd {
			continue // Skip start and end nodes
		}

		if priority, exists := nodePriority[subNode.Type]; exists && priority < bestPriority {
			bestPriority = priority
			bestNode = &subNodes[i]
		}
	}

	if bestNode != nil {
		outputName := g.inferNodeOutputName(*bestNode)
		return bestNode.ID, outputName
	}

	return "", ""
}

// fallbackOutputSelector fallback output selector strategy
func (g *IterationNodeGenerator) fallbackOutputSelector(subNodes []models.Node, defaultOutputName string) []string {
	// Find first non-start node as output
	for _, subNode := range subNodes {
		if subNode.Type != models.NodeTypeStart && subNode.Type != models.NodeTypeEnd {
			outputName := g.inferNodeOutputName(subNode)
			if outputName == "" && defaultOutputName != "" {
				outputName = defaultOutputName
			}
			if outputName == "" {
				outputName = "output"
			}

			return []string{subNode.ID, outputName}
		}
	}

	return []string{}
}

// mapToDifyOutputName maps iFlytek output names to Dify standard output names
func (g *IterationNodeGenerator) mapToDifyOutputName(nodeType models.NodeType, originalName string) string {
	switch nodeType {
	case models.NodeTypeCode:
		// Code nodes keep original output name, as user-defined output names are usually meaningful
		return originalName
	case models.NodeTypeLLM:
		// LLM nodes output uniformly as "text" in Dify
		return "text"
	case models.NodeTypeClassifier:
		// Classifier nodes output uniformly as "class_name" in Dify
		return "class_name"
	case models.NodeTypeCondition:
		// Condition nodes usually have no direct output, but if they do, keep original name
		return originalName
	default:
		// Other node types keep original output name
		return originalName
	}
}

// inferNodeOutputName infers node's output name
func (g *IterationNodeGenerator) inferNodeOutputName(node models.Node) string {
	// First try to use node-defined outputs
	if len(node.Outputs) > 0 {
		return g.mapToDifyOutputName(node.Type, node.Outputs[0].Name)
	}

	// If no outputs defined, infer standard output name based on node type
	switch node.Type {
	case models.NodeTypeCode:
		return "result" // Default output for code nodes
	case models.NodeTypeLLM:
		return "text" // Standard output for LLM nodes
	case models.NodeTypeClassifier:
		return "class_name" // Standard output for classifier nodes
	case models.NodeTypeCondition:
		return "result" // Default output for condition nodes
	default:
		return "output" // Generic default output
	}
}

// getNodeOutputName dynamically gets node's output name
func (g *IterationNodeGenerator) getNodeOutputName(difyNode *DifyNode, originalNode models.Node) string {
	// First check Dify node's outputs configuration (this is most accurate)
	if outputs, ok := difyNode.Data.Outputs.(map[string]interface{}); ok {
		for outputName := range outputs {
			return outputName // Return first output name
		}
	}

	// Second try to get from original node's outputs
	if len(originalNode.Outputs) > 0 {
		return originalNode.Outputs[0].Name
	}

	// If neither defined, infer default value based on node type
	switch difyNode.Data.Type {
	case "code":
		return "result" // Default output name for code nodes
	case "llm":
		return "text" // Standard output name for LLM nodes
	default:
		return "output" // Default output name
	}
}

// mapToValueType maps unified DSL types to Dify's value_type
func (g *IterationNodeGenerator) mapToValueType(unifiedType string) string {
	switch unifiedType {
	case "string":
		return "string"
	case "integer": // Support integer type
		return "number"
	case "float": // Support float type
		return "number"
	case "number": // Maintain backward compatibility
		return "number"
	case "boolean":
		return "boolean"
	case "array[string]":
		return "array[string]"
	case "array[number]":
		return "array[number]"
	case "object", "array[object]":
		return "object"
	default:
		return "string"
	}
}

// inferNodeIDAndOutputFieldName infers correct node ID and output field name for references
func (g *IterationNodeGenerator) inferNodeIDAndOutputFieldName(nodeID, originalOutputName, parentID string) (string, string) {
	// Core fix: In Dify iterators, internal node data source should be iterator's item, not start node's input

	// Try special iteration handling first
	if outputName := g.tryIterationSpecialHandling(nodeID, originalOutputName, parentID); outputName != "" {
		return parentID, outputName
	}

	// Try variable selector converter
	if convertedNodeID, convertedOutput := g.tryVariableSelectorConverter(nodeID, originalOutputName); convertedNodeID != "" {
		return convertedNodeID, convertedOutput
	}

	// Use fallback logic
	return g.useFallbackMapping(nodeID, originalOutputName)
}

// tryIterationSpecialHandling tries special iteration handling patterns
func (g *IterationNodeGenerator) tryIterationSpecialHandling(nodeID, originalOutputName, parentID string) string {
	// If referencing the iterator itself, use item
	if nodeID == parentID {
		return "item"
	}

	// Check for start node patterns that should map to iterator's item
	if g.shouldMapToIteratorItem(nodeID, originalOutputName) {
		return "item"
	}

	return ""
}

// shouldMapToIteratorItem checks if node reference should map to iterator's item
func (g *IterationNodeGenerator) shouldMapToIteratorItem(nodeID, originalOutputName string) bool {
	return g.isStartNodeWithInputOutput(nodeID, originalOutputName) ||
		g.isIterationStartNodeWithInput(nodeID, originalOutputName) ||
		g.isAnyStartNodeWithInput(nodeID, originalOutputName)
}

// isStartNodeWithInputOutput checks if it's a start node with input/steps output
func (g *IterationNodeGenerator) isStartNodeWithInputOutput(nodeID, originalOutputName string) bool {
	return strings.Contains(nodeID, "start") && (originalOutputName == "input" || originalOutputName == "steps")
}

// isIterationStartNodeWithInput checks if it's iteration-node-start with input output
func (g *IterationNodeGenerator) isIterationStartNodeWithInput(nodeID, originalOutputName string) bool {
	return strings.Contains(nodeID, "iteration-node-start") && originalOutputName == "input"
}

// isAnyStartNodeWithInput checks if any start node with input output
func (g *IterationNodeGenerator) isAnyStartNodeWithInput(nodeID, originalOutputName string) bool {
	return strings.Contains(nodeID, "start") && originalOutputName == "input"
}

// tryVariableSelectorConverter tries to use variable selector converter
func (g *IterationNodeGenerator) tryVariableSelectorConverter(nodeID, originalOutputName string) (string, string) {
	tempRef := &models.VariableReference{
		Type:       models.ReferenceTypeNodeOutput,
		NodeID:     nodeID,
		OutputName: originalOutputName,
	}

	valueSelector, err := g.variableSelectorConverter.ConvertVariableReference(tempRef)
	if err == nil && len(valueSelector) >= 2 {
		// Apply Dify field name mapping after conversion
		mappedFieldName := g.mapToDifyFieldName(valueSelector[1], nodeID)
		return valueSelector[0], mappedFieldName
	}

	return "", ""
}

// mapToDifyFieldName maps field names to Dify platform standards
func (g *IterationNodeGenerator) mapToDifyFieldName(originalFieldName, nodeID string) string {
	// Use node mapping to check actual node type from unified DSL
	if node, exists := g.nodeMapping[nodeID]; exists {
		switch node.Type {
		case models.NodeTypeLLM:
			// All LLM node outputs in Dify use 'text' field
			return "text"
		case models.NodeTypeClassifier:
			// Classifier nodes use 'class_name' field
			return "class_name"
		case models.NodeTypeCode, models.NodeTypeIteration:
			// Code and iteration nodes keep their original field names
			return originalFieldName
		}
	}

	// If node mapping fails, try to infer from node ID patterns
	if strings.Contains(nodeID, "llm") || strings.Contains(nodeID, "spark-llm") {
		return "text"
	}

	// Default: return original field name
	return originalFieldName
}

// useFallbackMapping uses fallback mapping logic based on node type
func (g *IterationNodeGenerator) useFallbackMapping(nodeID, originalOutputName string) (string, string) {
	// Check different node types and return appropriate mappings
	if mappedOutput := g.mapLLMNodeOutput(nodeID); mappedOutput != "" {
		return nodeID, mappedOutput
	}

	if mappedOutput := g.mapClassifierNodeOutput(nodeID); mappedOutput != "" {
		return nodeID, mappedOutput
	}

	if g.isCodeNode(nodeID) {
		return nodeID, originalOutputName
	}

	// Default return original node ID and output name
	return nodeID, originalOutputName
}

// mapLLMNodeOutput maps LLM node output
func (g *IterationNodeGenerator) mapLLMNodeOutput(nodeID string) string {
	if strings.Contains(nodeID, "spark-llm") || strings.Contains(nodeID, "llm") {
		return "text" // Standard output for LLM nodes is text
	}
	return ""
}

// mapClassifierNodeOutput maps classifier node output
func (g *IterationNodeGenerator) mapClassifierNodeOutput(nodeID string) string {
	if strings.Contains(nodeID, "decision-making") || strings.Contains(nodeID, "classifier") {
		return "class_name" // Standard output for classifier nodes is class_name
	}
	return ""
}

// isCodeNode checks if the node is a code node
func (g *IterationNodeGenerator) isCodeNode(nodeID string) bool {
	return strings.Contains(nodeID, "ifly-code") || strings.Contains(nodeID, "code")
}

// mapNodeOutputNameToDify maps output names to actual output names of generated Dify nodes based on original node ID
func (g *IterationNodeGenerator) mapNodeOutputNameToDify(originalNodeID, originalOutputName string, generatedNodes []DifyNode) string {
	// Find corresponding generated node
	for _, genNode := range generatedNodes {
		// Match through node ID (may need more intelligent mapping logic)
		if strings.Contains(genNode.ID, extractNodeTypeFromID(originalNodeID)) {
			// Get output name from actually generated node
			if outputs, ok := genNode.Data.Outputs.(map[string]interface{}); ok {
				for outputName := range outputs {
					return outputName // Return first actual output name
				}
			}

			// Map output name based on node type
			switch genNode.Data.Type {
			case "question-classifier":
				return "class_name" // Standard output for classifier nodes
			case "code":
				return originalOutputName // Code nodes keep original output name
			case "llm":
				return "text" // Standard output for LLM nodes
			default:
				return originalOutputName
			}
		}
	}

	// Default return original output name
	return originalOutputName
}

// extractNodeTypeFromID extracts node type identifier from node ID
func extractNodeTypeFromID(nodeID string) string {
	if strings.Contains(nodeID, "decision-making") || strings.Contains(nodeID, "classifier") {
		return "classifier"
	}
	if strings.Contains(nodeID, "ifly-code") || strings.Contains(nodeID, "code") {
		return "code"
	}
	if strings.Contains(nodeID, "spark-llm") || strings.Contains(nodeID, "llm") {
		return "llm"
	}
	return "unknown"
}

// fixIterationVariableReferences fixes variable references in iteration internal nodes
func (g *IterationNodeGenerator) fixIterationVariableReferences(text string, parentID string) string {

	text = strings.ReplaceAll(text, "{{#"+parentID+"start.input#}}", "{{#"+parentID+".item#}}")
	text = strings.ReplaceAll(text, "{{#"+parentID+"start.steps#}}", "{{#"+parentID+".item#}}")

	inputRe := regexp.MustCompile(`\{\{#\d+\.input#\}\}`)
	text = inputRe.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	stepsRe := regexp.MustCompile(`\{\{#\d+\.steps#\}\}`)
	text = stepsRe.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	startInputPattern := regexp.MustCompile(`\{\{#` + regexp.QuoteMeta(parentID) + `start\.input#\}\}`)
	text = startInputPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	startStepsPattern := regexp.MustCompile(`\{\{#` + regexp.QuoteMeta(parentID) + `start\.steps#\}\}`)
	text = startStepsPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	// For example: {{#1234567890123456start.input#}} -> {{#1234567890123456.item#}}
	anyStartInputPattern := regexp.MustCompile(`\{\{#(\d+)start\.input#\}\}`)
	text = anyStartInputPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract iteration ID (remove start suffix)
		matches := anyStartInputPattern.FindStringSubmatch(match)
		if len(matches) >= 2 {
			iterationID := matches[1]
			return "{{#" + iterationID + ".item#}}"
		}
		return match
	})

	anyStartStepsPattern := regexp.MustCompile(`\{\{#(\d+)start\.steps#\}\}`)
	text = anyStartStepsPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract iteration ID (remove start suffix)
		matches := anyStartStepsPattern.FindStringSubmatch(match)
		if len(matches) >= 2 {
			iterationID := matches[1]
			return "{{#" + iterationID + ".item#}}"
		}
		return match
	})

	// For example: {{#iteration-node-start::uuid.input#}} -> {{#parentID.item#}}
	uuidStartInputPattern := regexp.MustCompile(`\{\{#iteration-node-start::[^#]*\.input#\}\}`)
	text = uuidStartInputPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	uuidStartStepsPattern := regexp.MustCompile(`\{\{#iteration-node-start::[^#]*\.steps#\}\}`)
	text = uuidStartStepsPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	genericStartInputPattern := regexp.MustCompile(`\{\{#[^}]*start[^}]*\.input#\}\}`)
	text = genericStartInputPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	genericStartStepsPattern := regexp.MustCompile(`\{\{#[^}]*start[^}]*\.steps#\}\}`)
	text = genericStartStepsPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	inputPattern := regexp.MustCompile(`\{\{#[^}]*\.input#\}\}`)
	text = inputPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	stepsPattern := regexp.MustCompile(`\{\{#[^}]*\.steps#\}\}`)
	text = stepsPattern.ReplaceAllString(text, "{{#"+parentID+".item#}}")

	malformedPattern := regexp.MustCompile(`\{\{[^}#]*[^}#]"`)
	text = malformedPattern.ReplaceAllStringFunc(text, func(match string) string {
		// Extract the variable name from malformed pattern like {{class_name" or {{variableName"
		varName := strings.TrimPrefix(match, "{{")
		varName = strings.TrimSuffix(varName, "\"")

		// For specific known cases like class_name, we can try to be more intelligent
		if varName == "class_name" {
			// This should reference a classifier node's output, but without context
			return "{{#" + parentID + ".item#}}"
		}

		// Default fallback for other malformed patterns
		return "{{#" + parentID + ".item#}}"
	})

	return text
}
