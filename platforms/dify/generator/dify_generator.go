package generator

import (
	"agentbridge/core/interfaces"
	"agentbridge/internal/models"
	"agentbridge/platforms/common"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Interface compliance check at compile time
var _ interfaces.DSLGenerator = (*DifyGenerator)(nil)

// DifyGenerator Dify DSL generator
type DifyGenerator struct {
	*common.BaseGenerator
	nodeGeneratorFactory      *NodeGeneratorFactory
	edgeGenerator             *EdgeGenerator
	variableSelectorConverter *VariableSelectorConverter
	conditionCaseIDMapping    map[string]map[string]string // nodeID -> (original case_id -> Dify case_id)
}

func NewDifyGenerator() *DifyGenerator {
	return &DifyGenerator{
		BaseGenerator:             common.NewBaseGenerator(models.PlatformDify),
		nodeGeneratorFactory:      NewNodeGeneratorFactory(),
		edgeGenerator:             NewEdgeGenerator(),
		variableSelectorConverter: NewVariableSelectorConverter(),
		conditionCaseIDMapping:    make(map[string]map[string]string),
	}
}

// Generate generates Dify DSL from unified DSL
func (g *DifyGenerator) Generate(unifiedDSL *models.UnifiedDSL) ([]byte, error) {
	// Validate input
	if err := g.Validate(unifiedDSL); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build Dify DSL structure
	difyDSL := &DifyRootStructure{}

	// Generate app metadata
	if err := g.generateAppMetadata(unifiedDSL, difyDSL); err != nil {
		return nil, fmt.Errorf("failed to generate app metadata: %w", err)
	}

	// Generate workflow structure framework and get node ID mapping
	nodeIDMapping, err := g.generateWorkflowFramework(unifiedDSL, difyDSL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate workflow framework: %w", err)
	}

	// Apply a final pass to update all node references using the complete ID mapping
	g.finalizeNodeReferences(difyDSL, nodeIDMapping)

	// Serialize to YAML
	yamlData, err := yaml.Marshal(difyDSL)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	// Apply node ID mappings using the ID mapper for safer replacements
	yamlString := string(yamlData)
	idMapper := common.NewUnifiedIDMapper(common.StrategyTimestampBased).(*common.UnifiedIDMapper)
	idMapper.SetMapping(nodeIDMapping)

	// Use structured YAML processing instead of string replacement
	yamlString = g.applyNodeIDMappingsToYAML(yamlString, idMapper)

	// Add required empty fields for classifier nodes (if missing)
	yamlString = g.ensureClassifierRequiredFields(yamlString)

	yamlString = g.fixIterationStartNodeIDFormat(yamlString)

	return []byte(yamlString), nil
}

// finalizeNodeReferences updates selectors and references using the node ID mapping.
// It updates iteration selectors, start node id, iteration child parent/iteration ids,
// variable selectors in code/condition/classifier/end nodes, and template references.
func (g *DifyGenerator) finalizeNodeReferences(difyDSL *DifyRootStructure, nodeIDMapping map[string]string) {
	if difyDSL == nil {
		return
	}

	// Update nodes
	for i := range difyDSL.Workflow.Graph.Nodes {
		// Reuse existing updater which covers:
		// - context.variable_selector
		// - prompt_template references
		// - if-else cases variable_selector
		// - code variables value_selector
		// - end node outputs value_selector
		// - classifier query_variable_selector and instruction references
		// - iteration iterator_selector/output_selector/start_node_id
		// - iteration child node parentId / iteration_id
		_ = g.updateVariableSelectorsWithNewIDs(&difyDSL.Workflow.Graph.Nodes[i], models.Node{}, nodeIDMapping)
	}
}

// Validate validates if the unified DSL meets Dify platform requirements
func (g *DifyGenerator) Validate(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL cannot be nil")
	}

	if unifiedDSL.Metadata.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(unifiedDSL.Workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must contain at least one node")
	}

	// Validate that start and end nodes must exist
	hasStart := false
	hasEnd := false
	for _, node := range unifiedDSL.Workflow.Nodes {
		switch node.Type {
		case models.NodeTypeStart:
			hasStart = true
		case models.NodeTypeEnd:
			hasEnd = true
		}
	}

	if !hasStart {
		return fmt.Errorf("workflow must contain at least one start node")
	}

	if !hasEnd {
		return fmt.Errorf("workflow must contain at least one end node")
	}

	return nil
}

// generateAppMetadata generates app metadata
func (g *DifyGenerator) generateAppMetadata(unifiedDSL *models.UnifiedDSL, difyDSL *DifyRootStructure) error {
	app := DifyApp{
		Name:                unifiedDSL.Metadata.Name,
		Description:         unifiedDSL.Metadata.Description,
		Mode:                "workflow", // Default workflow mode
		Icon:                "🤖",        // Default icon
		IconBackground:      "#E8F5E8",  // Default background color
		UseIconAsAnswerIcon: false,
	}

	// Restore Dify-specific fields from platform-specific metadata
	if difyMeta := unifiedDSL.PlatformMetadata.Dify; difyMeta != nil {
		if difyMeta.Icon != "" {
			app.Icon = difyMeta.Icon
		}
		if difyMeta.IconBackground != "" {
			app.IconBackground = difyMeta.IconBackground
		}
		if difyMeta.Mode != "" {
			app.Mode = difyMeta.Mode
		}
		app.UseIconAsAnswerIcon = difyMeta.UseIconAsAnswerIcon
	}

	difyDSL.App = app

	// Set basic information
	difyDSL.Kind = "app"
	difyDSL.Version = "0.3.1" // Dify DSL specification version

	// Set dependencies: use marketplace dependencies to support langgenius/openai_api_compatible
	// Use stable versions currently supported by Dify platform
	difyDSL.Dependencies = []DifyDependency{
		{
			CurrentIdentifier: nil,
			Type:              "marketplace",
			Value: DifyDepValue{
				MarketplacePluginUniqueIdentifier: "langgenius/openai_api_compatible:0.0.19@219552f62b54919d6fd317c737956d3b2cc97719b85f0179bb995e5a512b7ebb",
			},
		},
	}

	return nil
}

// generateWorkflowFramework generates basic workflow structure framework
func (g *DifyGenerator) generateWorkflowFramework(unifiedDSL *models.UnifiedDSL, difyDSL *DifyRootStructure) (map[string]string, error) {
	graph, nodeIDMapping, err := g.generateGraphFramework(unifiedDSL)
	if err != nil {
		return nil, err
	}

	workflow := DifyWorkflow{
		ConversationVariables: []interface{}{},
		EnvironmentVariables:  []interface{}{},
		Features:              g.generateFeatures(unifiedDSL),
		Graph:                 graph,
	}

	difyDSL.Workflow = workflow
	return nodeIDMapping, nil
}

// generateFeatures generates features configuration from unified DSL
func (g *DifyGenerator) generateFeatures(unifiedDSL *models.UnifiedDSL) DifyFeatures {
	features := DifyFeatures{
		FileUpload: DifyFileUpload{
			Enabled:                  false,
			AllowedFileExtensions:    []string{".JPG", ".JPEG", ".PNG", ".GIF", ".WEBP", ".SVG"},
			AllowedFileTypes:         []string{"image"},
			AllowedFileUploadMethods: []string{"local_file", "remote_url"},
			FileUploadConfig: DifyFileUploadConfig{
				AudioFileSizeLimit:      50,
				BatchCountLimit:         10,
				FileSizeLimit:           100,
				ImageFileSizeLimit:      20,
				VideoFileSizeLimit:      100,
				WorkflowFileUploadLimit: 10,
			},
			Image: DifyImageConfig{
				Enabled:         false,
				NumberLimits:    3,
				TransferMethods: []string{"local_file", "remote_url"},
			},
			NumberLimits: 3,
		},
		RetrieverResource: DifyRetrieverResource{
			Enabled: true,
		},
		SensitiveWordAvoidance: DifySensitiveWordAvoidance{
			Enabled: false,
		},
		SpeechToText: DifySpeechToText{
			Enabled: false,
		},
		SuggestedQuestionsAfterAnswer: DifySuggestedQuestionsAfterAnswer{
			Enabled: false,
		},
		TextToSpeech: DifyTextToSpeech{
			Enabled:  false,
			Language: "",
			Voice:    "",
		},
	}

	// Get configuration from unified DSL's UIConfig
	if unifiedDSL.Metadata.UIConfig != nil {
		uiConfig := unifiedDSL.Metadata.UIConfig

		// Set opening statement
		if uiConfig.OpeningStatement != "" {
			features.OpeningStatement = uiConfig.OpeningStatement
		}

		// Set suggested questions
		if len(uiConfig.SuggestedQuestions) > 0 {
			features.SuggestedQuestions = uiConfig.SuggestedQuestions
		}
	}

	// If not configured, set to empty (don't use hardcoded default values)
	if features.OpeningStatement == "" {
		features.OpeningStatement = ""
	}
	if len(features.SuggestedQuestions) == 0 {
		features.SuggestedQuestions = []string{}
	}

	return features
}

// generateGraphFramework generates graph structure framework
func (g *DifyGenerator) generateGraphFramework(unifiedDSL *models.UnifiedDSL) (DifyGraph, map[string]string, error) {
	// Initialize graph context and mappings
	graph, nodeIDMapping := g.initializeGraphContext(unifiedDSL)

	// Generate all nodes for the workflow
	if err := g.generateNodesForWorkflow(unifiedDSL, &graph, nodeIDMapping); err != nil {
		return DifyGraph{}, nil, err
	}

	// Generate all edges for the workflow
	if err := g.generateEdgesForWorkflow(unifiedDSL, &graph, nodeIDMapping); err != nil {
		return DifyGraph{}, nil, err
	}

	return graph, nodeIDMapping, nil
}

// initializeGraphContext initializes the graph generation context
func (g *DifyGenerator) initializeGraphContext(unifiedDSL *models.UnifiedDSL) (DifyGraph, map[string]string) {
	graph := DifyGraph{
		Edges: make([]DifyEdge, 0, len(unifiedDSL.Workflow.Edges)),
		Nodes: make([]DifyNode, 0, len(unifiedDSL.Workflow.Nodes)),
	}

	nodeIDMapping := make(map[string]string)

	// Set node mapping for variable selector converter
	g.variableSelectorConverter.SetNodeMapping(unifiedDSL.Workflow.Nodes)

	// Set node mapping for node generator factory (for variable reference resolution)
	g.nodeGeneratorFactory.SetNodeMapping(unifiedDSL.Workflow.Nodes)

	return graph, nodeIDMapping
}

// generateNodesForWorkflow generates all nodes for the workflow
func (g *DifyGenerator) generateNodesForWorkflow(unifiedDSL *models.UnifiedDSL, graph *DifyGraph, nodeIDMapping map[string]string) error {
	for i, node := range unifiedDSL.Workflow.Nodes {
		if err := g.generateSingleNode(node, i, unifiedDSL, graph, nodeIDMapping); err != nil {
			return fmt.Errorf("failed to generate node %s: %w", node.ID, err)
		}
	}

	// Post-process iteration nodes to fix output selectors after all mappings are established
	g.postProcessIterationNodes(graph, nodeIDMapping)

	return nil
}

// generateSingleNode generates a single node and adds it to the graph
func (g *DifyGenerator) generateSingleNode(node models.Node, index int, unifiedDSL *models.UnifiedDSL, graph *DifyGraph, nodeIDMapping map[string]string) error {
	difyNodes, err := g.generateNodesByType(node, unifiedDSL)
	if err != nil {
		return err
	}

	return g.processGeneratedNodes(difyNodes, node, index, nodeIDMapping, graph)
}

// generateNodesByType generates nodes based on node type
func (g *DifyGenerator) generateNodesByType(node models.Node, unifiedDSL *models.UnifiedDSL) ([]DifyNode, error) {
	switch node.Type {
	case models.NodeTypeIteration:
		return g.generateIterationNodes(node)
	case models.NodeTypeEnd:
		return g.generateEndNodes(node, unifiedDSL)
	default:
		return g.generateRegularNode(node)
	}
}

// generateIterationNodes generates iteration nodes
func (g *DifyGenerator) generateIterationNodes(node models.Node) ([]DifyNode, error) {
	if iterGenerator, ok := g.nodeGeneratorFactory.generators[models.NodeTypeIteration].(*IterationNodeGenerator); ok {
		return iterGenerator.GenerateIterationNodes(node)
	}
	// Fallback to single node generation
	singleNode, err := g.nodeGeneratorFactory.GenerateNode(node)
	if err != nil {
		return nil, err
	}
	return []DifyNode{singleNode}, nil
}

// generateEndNodes generates end nodes with workflow context
func (g *DifyGenerator) generateEndNodes(node models.Node, unifiedDSL *models.UnifiedDSL) ([]DifyNode, error) {
	if endGenerator, ok := g.nodeGeneratorFactory.generators[models.NodeTypeEnd].(*EndNodeGenerator); ok {
		singleNode, err := endGenerator.GenerateNodeWithWorkflowContext(node, &unifiedDSL.Workflow)
		if err != nil {
			return nil, err
		}
		return []DifyNode{singleNode}, nil
	}
	return g.generateRegularNode(node)
}

// generateRegularNode generates regular nodes
func (g *DifyGenerator) generateRegularNode(node models.Node) ([]DifyNode, error) {
	singleNode, err := g.nodeGeneratorFactory.GenerateNode(node)
	if err != nil {
		return nil, err
	}
	return []DifyNode{singleNode}, nil
}

// processGeneratedNodes processes generated nodes and adds them to the graph
func (g *DifyGenerator) processGeneratedNodes(difyNodes []DifyNode, originalNode models.Node, index int, nodeIDMapping map[string]string, graph *DifyGraph) error {
	for j, difyNode := range difyNodes {
		simpleID := g.generateNodeID(originalNode, index, j, nodeIDMapping)
		difyNode.ID = simpleID

		if err := g.updateNodeMappings(originalNode, difyNode, simpleID, j, nodeIDMapping); err != nil {
			return err
		}

		if err := g.updateVariableSelectorsWithNewIDs(&difyNode, originalNode, nodeIDMapping); err != nil {
			return fmt.Errorf("failed to update variable selectors for node %s: %w", originalNode.ID, err)
		}

		g.setDefaultPositionIfNeeded(&difyNode, index)
		g.updateIterationStartNode(&difyNode, originalNode, graph)

		graph.Nodes = append(graph.Nodes, difyNode)
	}
	return nil
}

// generateNodeID generates appropriate node ID based on node position
func (g *DifyGenerator) generateNodeID(originalNode models.Node, index, nodeIndex int, nodeIDMapping map[string]string) string {
	if nodeIndex == 0 {
		// Main node uses original ID generation logic
		simpleID := common.GenerateSimpleNodeID(originalNode, index)
		nodeIDMapping[originalNode.ID] = simpleID
		return simpleID
	}

	// Child nodes use special ID logic
	if originalNode.Type == models.NodeTypeIteration {
		mainNodeID := nodeIDMapping[originalNode.ID]
		if nodeIndex == 1 {
			// First child node is the iteration start node
			return mainNodeID + "start"
		} else {
			// Other child nodes get unique IDs
			return common.GenerateSimpleNodeID(originalNode, index*1000+nodeIndex)
		}
	}

	return common.GenerateSimpleNodeID(originalNode, index*1000+nodeIndex)
}

// updateNodeMappings updates node ID mappings for iteration nodes
func (g *DifyGenerator) updateNodeMappings(originalNode models.Node, difyNode DifyNode, simpleID string, nodeIndex int, nodeIDMapping map[string]string) error {
	if originalNode.Type == models.NodeTypeIteration && nodeIndex > 0 {
		if iterConfig, ok := originalNode.Config.(*models.IterationConfig); ok {
			g.updateIterationNodeMappings(iterConfig, difyNode, simpleID, nodeIndex, nodeIDMapping)
		}
	}

	// If child node has original ID, also add to mapping
	if difyNode.ID != "" {
		nodeIDMapping[difyNode.ID] = simpleID
	}

	return nil
}

// updateIterationNodeMappings updates mappings for iteration internal nodes
func (g *DifyGenerator) updateIterationNodeMappings(iterConfig *models.IterationConfig, difyNode DifyNode, simpleID string, nodeIndex int, nodeIDMapping map[string]string) {
	if difyNode.Data.Type == "iteration-start" {
		if iterConfig.SubWorkflow.StartNodeID != "" {
			nodeIDMapping[iterConfig.SubWorkflow.StartNodeID] = simpleID
		}
	} else if nodeIndex >= 2 {
		// Calculate index in internal processing nodes
		internalNodeIndex := 0
		targetIndex := nodeIndex - 2

		for _, subNode := range iterConfig.SubWorkflow.Nodes {
			if subNode.Type != models.NodeTypeStart && subNode.Type != models.NodeTypeEnd {
				if internalNodeIndex == targetIndex {
					nodeIDMapping[subNode.ID] = simpleID
					break
				}
				internalNodeIndex++
			}
		}
	}
}

// setDefaultPositionIfNeeded sets default position if node has no position
func (g *DifyGenerator) setDefaultPositionIfNeeded(difyNode *DifyNode, index int) {
	if difyNode.Position.X == 0 && difyNode.Position.Y == 0 {
		difyNode.Position = DifyPosition{X: float64(index * 300), Y: 100}
		difyNode.PositionAbsolute = DifyPosition{X: float64(index * 300), Y: 100}
	}
}

// updateIterationStartNode updates iteration start node reference
func (g *DifyGenerator) updateIterationStartNode(difyNode *DifyNode, originalNode models.Node, graph *DifyGraph) {
	if originalNode.Type == models.NodeTypeIteration && difyNode.Data.Type == "iteration-start" {
		// Find corresponding iteration main node and update its start_node_id
		for k := range graph.Nodes {
			if graph.Nodes[k].Data.Type == "iteration" && graph.Nodes[k].ParentID == "" {
				if graph.Nodes[k].ID == difyNode.ParentID {
					graph.Nodes[k].Data.StartNodeID = difyNode.ID
					break
				}
			}
		}
	}
}

// generateEdgesForWorkflow generates all edges for the workflow
func (g *DifyGenerator) generateEdgesForWorkflow(unifiedDSL *models.UnifiedDSL, graph *DifyGraph, nodeIDMapping map[string]string) error {
	// Collect iteration internal node IDs for filtering
	iterationInternalNodeIDs := common.CollectIterationInternalNodeIDs(unifiedDSL.Workflow.Nodes)

	// Collect all edges and nodes from main workflow and iterations
	allEdges, allNodes := common.CollectAllEdgesAndNodes(unifiedDSL, iterationInternalNodeIDs)

	// Set node mapping for iteration sub-workflow nodes
	g.setSubWorkflowNodeMappings(unifiedDSL.Workflow.Nodes)

	// Use connection generator to generate connections
	difyEdges, err := g.edgeGenerator.GenerateEdgesWithIDMapping(allEdges, allNodes, nodeIDMapping)
	if err != nil {
		return fmt.Errorf("failed to generate edges: %w", err)
	}

	graph.Edges = difyEdges
	return nil
}

// setSubWorkflowNodeMappings sets node mapping for iteration sub-workflow nodes
func (g *DifyGenerator) setSubWorkflowNodeMappings(nodes []models.Node) {
	for _, node := range nodes {
		if node.Type == models.NodeTypeIteration {
			if iterConfig, ok := node.Config.(*models.IterationConfig); ok {
				g.nodeGeneratorFactory.SetNodeMapping(iterConfig.SubWorkflow.Nodes)
			}
		}
	}
}

// updateVariableSelectorsWithNewIDs updates node IDs in variable selectors
func (g *DifyGenerator) updateVariableSelectorsWithNewIDs(difyNode *DifyNode, originalNode models.Node, nodeIDMapping map[string]string) error {
	g.updateContextVariableSelector(difyNode, nodeIDMapping)
	g.updatePromptTemplateReferences(difyNode, nodeIDMapping)
	g.updateCaseConditionSelectors(difyNode, nodeIDMapping)
	g.updateCodeVariableSelectors(difyNode, nodeIDMapping)
	g.updateOutputValueSelectors(difyNode, nodeIDMapping)
	g.updateClassifierQuerySelector(difyNode, nodeIDMapping)
	g.updateIterationNodeSelectors(difyNode, nodeIDMapping)
	g.updateIterationChildNodeReferences(difyNode, nodeIDMapping)

	return nil
}

// updateContextVariableSelector updates LLM node's context.variable_selector
func (g *DifyGenerator) updateContextVariableSelector(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if difyNode.Data.Context == nil {
		return
	}

	variableSelector, exists := difyNode.Data.Context["variable_selector"].([]string)
	if !exists || len(variableSelector) < 2 {
		return
	}

	oldNodeID := variableSelector[0]
	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		difyNode.Data.Context["variable_selector"] = []string{newNodeID, variableSelector[1]}
	}
}

// updatePromptTemplateReferences updates variable references in prompt_template
func (g *DifyGenerator) updatePromptTemplateReferences(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if difyNode.Data.PromptTemplate == nil {
		return
	}

	for _, template := range difyNode.Data.PromptTemplate {
		text, exists := template["text"].(string)
		if !exists {
			continue
		}

		updatedText := g.replaceTemplateNodeReferences(text, nodeIDMapping)
		template["text"] = updatedText
	}
}

// replaceTemplateNodeReferences replaces node ID references in template text
func (g *DifyGenerator) replaceTemplateNodeReferences(text string, nodeIDMapping map[string]string) string {
	return common.ReplaceTemplateNodeReferences(text, nodeIDMapping)
}

// updateCaseConditionSelectors updates variable_selector in if-else node cases
func (g *DifyGenerator) updateCaseConditionSelectors(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if difyNode.Data.Cases == nil {
		return
	}

	for _, caseItem := range difyNode.Data.Cases {
		conditions, exists := caseItem["conditions"].([]map[string]interface{})
		if !exists {
			continue
		}

		for _, condition := range conditions {
			g.updateConditionVariableSelector(condition, nodeIDMapping)
		}
	}
}

// updateConditionVariableSelector updates a single condition's variable selector
func (g *DifyGenerator) updateConditionVariableSelector(condition map[string]interface{}, nodeIDMapping map[string]string) {
	variableSelector, exists := condition["variable_selector"].([]string)
	if !exists {
		return
	}

	updatedSelector := common.UpdateVariableSelector(variableSelector, nodeIDMapping)
	condition["variable_selector"] = updatedSelector
}

// updateCodeVariableSelectors updates value_selector in code node variables
func (g *DifyGenerator) updateCodeVariableSelectors(difyNode *DifyNode, nodeIDMapping map[string]string) {
	variables, ok := difyNode.Data.Variables.([]map[string]interface{})
	if !ok {
		return
	}

	for _, variable := range variables {
		g.updateVariableValueSelector(variable, nodeIDMapping)
	}
}

// updateVariableValueSelector updates a single variable's value selector
func (g *DifyGenerator) updateVariableValueSelector(variable map[string]interface{}, nodeIDMapping map[string]string) {
	valueSelector, exists := variable["value_selector"].([]string)
	if !exists {
		return
	}

	updatedSelector := common.UpdateVariableSelector(valueSelector, nodeIDMapping)
	variable["value_selector"] = updatedSelector
}

// updateOutputValueSelectors updates value_selector in end node outputs
func (g *DifyGenerator) updateOutputValueSelectors(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if outputs, ok := difyNode.Data.Outputs.([]DifyOutput); ok {
		g.updateDifyOutputSelectors(outputs, nodeIDMapping)
		difyNode.Data.Outputs = outputs
	} else {
		g.updateValueSelectorsInOutputs(&difyNode.Data.Outputs, nodeIDMapping)
	}
}

// updateDifyOutputSelectors updates value selectors in DifyOutput array
func (g *DifyGenerator) updateDifyOutputSelectors(outputs []DifyOutput, nodeIDMapping map[string]string) {
	for i := range outputs {
		if len(outputs[i].ValueSelector) >= 2 {
			oldNodeID := outputs[i].ValueSelector[0]
			if newNodeID, found := nodeIDMapping[oldNodeID]; found {
				outputs[i].ValueSelector[0] = newNodeID
			}
		}
	}
}

// updateClassifierQuerySelector updates query_variable_selector and instruction fields in classifier nodes
func (g *DifyGenerator) updateClassifierQuerySelector(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if len(difyNode.Data.QueryVariableSelector) < 2 {
		return
	}

	oldNodeID := difyNode.Data.QueryVariableSelector[0]
	newNodeID, found := nodeIDMapping[oldNodeID]
	if !found {
		return
	}

	difyNode.Data.QueryVariableSelector[0] = newNodeID
	g.updateInstructionNodeReferences(difyNode, oldNodeID, newNodeID)
}

// updateInstructionNodeReferences updates variable references in instruction field
func (g *DifyGenerator) updateInstructionNodeReferences(difyNode *DifyNode, oldNodeID, newNodeID string) {
	if difyNode.Data.Instruction == "" {
		return
	}

	oldPattern := fmt.Sprintf("{{#%s.", oldNodeID)
	newPattern := fmt.Sprintf("{{#%s.", newNodeID)
	difyNode.Data.Instruction = strings.ReplaceAll(difyNode.Data.Instruction, oldPattern, newPattern)
}

// updateValueSelectorsInOutputs updates value_selector in outputs (generic method)
func (g *DifyGenerator) updateValueSelectorsInOutputs(outputs *interface{}, nodeIDMapping map[string]string) {
	if outputs == nil {
		return
	}

	// Handle different output slice types
	g.processInterfaceSliceOutputs(outputs, nodeIDMapping)
	g.processMapSliceOutputs(outputs, nodeIDMapping)
}

// processInterfaceSliceOutputs processes []interface{} type outputs
func (g *DifyGenerator) processInterfaceSliceOutputs(outputs *interface{}, nodeIDMapping map[string]string) {
	outputsSlice, ok := (*outputs).([]interface{})
	if !ok {
		return
	}

	for _, outputInterface := range outputsSlice {
		g.processInterfaceOutput(outputInterface, nodeIDMapping)
	}
}

// processInterfaceOutput processes a single interface{} output
func (g *DifyGenerator) processInterfaceOutput(outputInterface interface{}, nodeIDMapping map[string]string) {
	outputMap, ok := outputInterface.(map[string]interface{})
	if !ok {
		return
	}

	g.updateInterfaceValueSelector(outputMap, nodeIDMapping)
}

// updateInterfaceValueSelector updates value_selector in map[string]interface{}
func (g *DifyGenerator) updateInterfaceValueSelector(outputMap map[string]interface{}, nodeIDMapping map[string]string) {
	valueSelector, exists := outputMap["value_selector"].([]interface{})
	if !exists || len(valueSelector) < 2 {
		return
	}

	oldNodeID, ok := valueSelector[0].(string)
	if !ok {
		return
	}

	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		valueSelector[0] = newNodeID
	}
}

// processMapSliceOutputs processes []map[string]interface{} type outputs
func (g *DifyGenerator) processMapSliceOutputs(outputs *interface{}, nodeIDMapping map[string]string) {
	outputsSlice, ok := (*outputs).([]map[string]interface{})
	if !ok {
		return
	}

	for _, outputMap := range outputsSlice {
		g.processMapOutput(outputMap, nodeIDMapping)
	}
}

// processMapOutput processes a single map[string]interface{} output
func (g *DifyGenerator) processMapOutput(outputMap map[string]interface{}, nodeIDMapping map[string]string) {
	// Try interface{} value selector first
	if g.tryUpdateInterfaceValueSelector(outputMap, nodeIDMapping) {
		return
	}

	// Try string value selector
	g.tryUpdateStringValueSelector(outputMap, nodeIDMapping)
}

// tryUpdateInterfaceValueSelector tries to update []interface{} value selector
func (g *DifyGenerator) tryUpdateInterfaceValueSelector(outputMap map[string]interface{}, nodeIDMapping map[string]string) bool {
	valueSelector, exists := outputMap["value_selector"].([]interface{})
	if !exists || len(valueSelector) < 2 {
		return false
	}

	oldNodeID, ok := valueSelector[0].(string)
	if !ok {
		return false
	}

	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		valueSelector[0] = newNodeID
	}
	return true
}

// tryUpdateStringValueSelector tries to update []string value selector
func (g *DifyGenerator) tryUpdateStringValueSelector(outputMap map[string]interface{}, nodeIDMapping map[string]string) {
	valueSelector, exists := outputMap["value_selector"].([]string)
	if !exists || len(valueSelector) < 2 {
		return
	}

	oldNodeID := valueSelector[0]
	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		outputMap["value_selector"] = []string{newNodeID, valueSelector[1]}
	}
}

// fixIterationStartNodeIDFormat fixes YAML format issues for iteration start node IDs
func (g *DifyGenerator) fixIterationStartNodeIDFormat(yamlString string) string {
	// Use regex to find and fix issues with 'numeric'start format
	// Pattern: 'numeric'start -> 'numericstart'
	re := regexp.MustCompile(`'(\d+)'start`)
	yamlString = re.ReplaceAllString(yamlString, `'${1}start'`)

	return yamlString
}

// ensureClassifierRequiredFields ensures classifier nodes contain required empty fields
func (g *DifyGenerator) ensureClassifierRequiredFields(yamlString string) string {
	if !g.isClassifierNode(yamlString) {
		return yamlString
	}

	yamlString = g.ensureInstructionsField(yamlString)
	yamlString = g.ensureTopicsField(yamlString)

	return yamlString
}

// isClassifierNode checks if the YAML contains classifier nodes
func (g *DifyGenerator) isClassifierNode(yamlString string) bool {
	return strings.Contains(yamlString, "type: question-classifier")
}

// ensureInstructionsField adds missing instructions field to classifier nodes
func (g *DifyGenerator) ensureInstructionsField(yamlString string) string {
	if strings.Contains(yamlString, "instructions:") {
		return yamlString
	}

	return strings.Replace(yamlString,
		"                query_variable_selector:",
		"                instructions: \"\"\n                query_variable_selector:", 1)
}

// ensureTopicsField adds missing topics field to classifier nodes
func (g *DifyGenerator) ensureTopicsField(yamlString string) string {
	if strings.Contains(yamlString, "topics:") {
		return yamlString
	}

	return g.addTopicsFieldToClassifier(yamlString)
}

// addTopicsFieldToClassifier adds topics field after query_variable_selector
func (g *DifyGenerator) addTopicsFieldToClassifier(yamlString string) string {
	lines := strings.Split(yamlString, "\n")

	for i, line := range lines {
		if g.isQueryVariableSelectorLine(line, lines, i) {
			return g.insertTopicsField(lines, i)
		}
	}

	return yamlString
}

// isQueryVariableSelectorLine checks if current line is query_variable_selector in classifier
func (g *DifyGenerator) isQueryVariableSelectorLine(line string, lines []string, index int) bool {
	if !strings.Contains(line, "query_variable_selector:") || index == 0 {
		return false
	}

	return g.isInClassifierContext(lines, index)
}

// isInClassifierContext checks if we're in classifier node context
func (g *DifyGenerator) isInClassifierContext(lines []string, currentIndex int) bool {
	start := currentIndex - 10
	if start < 0 {
		start = 0
	}

	for j := start; j < currentIndex; j++ {
		if strings.Contains(lines[j], "type: question-classifier") {
			return true
		}
	}
	return false
}

// insertTopicsField inserts topics field at appropriate position
func (g *DifyGenerator) insertTopicsField(lines []string, selectorIndex int) string {
	insertPos := selectorIndex + 3
	if insertPos >= len(lines) {
		return strings.Join(lines, "\n")
	}

	newLines := append(lines[:insertPos], append([]string{"                topics: []"}, lines[insertPos:]...)...)
	return strings.Join(newLines, "\n")
}

// applyNodeIDMappingsToYAML applies node ID mappings using structured processing
func (g *DifyGenerator) applyNodeIDMappingsToYAML(yamlString string, idMapper *common.UnifiedIDMapper) string {
	// Process YAML using regex patterns to safely replace node IDs
	// This approach is safer than direct string replacement as it targets specific patterns

	for oldID, newID := range idMapper.GetMapping() {
		// Replace node IDs in edges (source/target fields)
		yamlString = g.replaceNodeIDInField(yamlString, "source:", oldID, newID)
		yamlString = g.replaceNodeIDInField(yamlString, "target:", oldID, newID)

		// Replace node IDs in variable selectors and references
		yamlString = g.replaceNodeIDInVariableSelectors(yamlString, oldID, newID)

		// Replace node IDs in template references
		yamlString = g.replaceNodeIDInTemplates(yamlString, oldID, newID)
	}

	return yamlString
}

// replaceNodeIDInField safely replaces node IDs in specific YAML fields
func (g *DifyGenerator) replaceNodeIDInField(yamlString, fieldName, oldID, newID string) string {
	// Pattern to match field: "oldID" with proper YAML formatting
	pattern := fmt.Sprintf(`(%s\s+)(%s)(\s|$)`, regexp.QuoteMeta(fieldName), regexp.QuoteMeta(oldID))
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllString(yamlString, fmt.Sprintf("${1}%s${3}", newID))
}

// replaceNodeIDInVariableSelectors replaces node IDs in variable selector arrays
func (g *DifyGenerator) replaceNodeIDInVariableSelectors(yamlString, oldID, newID string) string {
	// Pattern to match variable selectors: [oldID, fieldName]
	pattern := fmt.Sprintf(`(\[\s*)(%s)(\s*,)`, regexp.QuoteMeta(oldID))
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllString(yamlString, fmt.Sprintf("${1}%s${3}", newID))
}

// replaceNodeIDInTemplates replaces node IDs in template expressions
func (g *DifyGenerator) replaceNodeIDInTemplates(yamlString, oldID, newID string) string {
	// Pattern to match template references: {{#oldID.field#}}
	pattern := fmt.Sprintf(`(\{\{#)(%s)(\.[\w]+#\}\})`, regexp.QuoteMeta(oldID))
	re := regexp.MustCompile(pattern)

	return re.ReplaceAllString(yamlString, fmt.Sprintf("${1}%s${3}", newID))
}

// updateIterationNodeSelectors updates iteration node specific selectors and references
func (g *DifyGenerator) updateIterationNodeSelectors(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if difyNode.Data.Type != "iteration" {
		return
	}

	// Update iterator_selector
	g.updateIteratorSelector(difyNode, nodeIDMapping)

	// Update output_selector
	g.updateIterationOutputSelector(difyNode, nodeIDMapping)

	// Update start_node_id
	g.updateStartNodeID(difyNode, nodeIDMapping)
}

// updateIteratorSelector updates iterator_selector field in iteration nodes
func (g *DifyGenerator) updateIteratorSelector(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if len(difyNode.Data.IteratorSelector) < 2 {
		return
	}

	oldNodeID := difyNode.Data.IteratorSelector[0]
	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		difyNode.Data.IteratorSelector[0] = newNodeID
	}
}

// updateIterationOutputSelector updates output_selector field in iteration nodes
func (g *DifyGenerator) updateIterationOutputSelector(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if len(difyNode.Data.OutputSelector) < 2 {
		return
	}

	oldNodeID := difyNode.Data.OutputSelector[0]
	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		difyNode.Data.OutputSelector[0] = newNodeID
	}
}

// updateStartNodeID updates start_node_id field in iteration nodes
func (g *DifyGenerator) updateStartNodeID(difyNode *DifyNode, nodeIDMapping map[string]string) {
	if difyNode.Data.StartNodeID == "" {
		return
	}

	// Handle cases like "originalNodeIDstart" -> "newNodeIDstart"
	for oldNodeID, newNodeID := range nodeIDMapping {
		if strings.HasPrefix(difyNode.Data.StartNodeID, oldNodeID) {
			suffix := strings.TrimPrefix(difyNode.Data.StartNodeID, oldNodeID)
			difyNode.Data.StartNodeID = newNodeID + suffix
			break
		}
	}
}

// updateIterationChildNodeReferences updates parentId and iteration_id in iteration child nodes
func (g *DifyGenerator) updateIterationChildNodeReferences(difyNode *DifyNode, nodeIDMapping map[string]string) {
	// Update ParentID if it exists in the mapping
	if difyNode.ParentID != "" {
		if newParentID, found := nodeIDMapping[difyNode.ParentID]; found {
			difyNode.ParentID = newParentID
		}
	}

	// Update iteration_id in Data if it exists in the mapping
	if difyNode.Data.IterationID != "" {
		if newIterationID, found := nodeIDMapping[difyNode.Data.IterationID]; found {
			difyNode.Data.IterationID = newIterationID
		}
	}
}

// postProcessIterationNodes post-processes iteration nodes to fix selectors after all mappings are established
func (g *DifyGenerator) postProcessIterationNodes(graph *DifyGraph, nodeIDMapping map[string]string) {
	for i := range graph.Nodes {
		node := &graph.Nodes[i]
		if node.Data.Type == "iteration" {
			g.fixIterationOutputSelector(node, nodeIDMapping)
		}
	}
}

// fixIterationOutputSelector fixes output_selector for iteration nodes using complete node mapping
func (g *DifyGenerator) fixIterationOutputSelector(iterationNode *DifyNode, nodeIDMapping map[string]string) {
	if len(iterationNode.Data.OutputSelector) < 2 {
		return
	}

	oldNodeID := iterationNode.Data.OutputSelector[0]
	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		iterationNode.Data.OutputSelector[0] = newNodeID
	}
}
