package generator

import (
    "ai-agents-transformer/internal/models"
    "ai-agents-transformer/platforms/common"
    "fmt"
    "regexp"
    "strings"
)

// IterationNodeGenerator generates Coze iteration nodes
type IterationNodeGenerator struct {
	idGenerator   *CozeIDGenerator
	nodeFactory   *NodeGeneratorFactory
	edgeGenerator *EdgeGenerator
}

// NewIterationNodeGenerator creates an iteration node generator
func NewIterationNodeGenerator() *IterationNodeGenerator {
	return &IterationNodeGenerator{
		idGenerator:   nil, // Set by the main generator
		nodeFactory:   nil, // Set by the main generator
		edgeGenerator: nil, // Set by the main generator
	}
}

// SetIDGenerator sets the shared ID generator
func (g *IterationNodeGenerator) SetIDGenerator(idGenerator *CozeIDGenerator) {
	g.idGenerator = idGenerator
}

// SetNodeFactory sets the node factory for generating sub-nodes
func (g *IterationNodeGenerator) SetNodeFactory(factory *NodeGeneratorFactory) {
	g.nodeFactory = factory
}

// SetEdgeGenerator sets the edge generator for generating internal edges
func (g *IterationNodeGenerator) SetEdgeGenerator(edgeGenerator *EdgeGenerator) {
	g.edgeGenerator = edgeGenerator
}

// GetNodeType returns the node type this generator handles
func (g *IterationNodeGenerator) GetNodeType() models.NodeType {
	return models.NodeTypeIteration
}

// ValidateNode validates the unified node before generation
func (g *IterationNodeGenerator) ValidateNode(unifiedNode *models.Node) error {
	if unifiedNode == nil {
		return fmt.Errorf("unified node is nil")
	}

	if unifiedNode.Type != models.NodeTypeIteration {
		return fmt.Errorf("invalid node type: expected %s, got %s", models.NodeTypeIteration, unifiedNode.Type)
	}

	if unifiedNode.Config == nil {
		return fmt.Errorf("node config is nil")
	}

    iterationConfig, ok := common.AsIterationConfig(unifiedNode.Config)
    if !ok || iterationConfig == nil {
        return fmt.Errorf("invalid config type: expected IterationConfig")
    }

	if iterationConfig.Iterator.SourceNode == "" {
		return fmt.Errorf("iteration source node is empty")
	}

	return nil
}

// GenerateNode generates a Coze workflow iteration node
func (g *IterationNodeGenerator) GenerateNode(unifiedNode *models.Node) (*CozeNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// Set current iteration node ID for edge generator use
	g.idGenerator.SetCurrentIterationNodeID(cozeNodeID)

    // Extract iteration configuration
    iterationConfig, ok := common.AsIterationConfig(unifiedNode.Config)
    if !ok || iterationConfig == nil {
        return nil, fmt.Errorf("invalid iteration config type for node %s", unifiedNode.ID)
    }

	// Generate sub-blocks
	blocks, err := g.generateSubBlocks(iterationConfig.SubWorkflow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sub blocks: %v", err)
	}

	// Generate internal edges
	edges, err := g.generateInternalEdges(iterationConfig.SubWorkflow.Edges, iterationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate internal edges: %v", err)
	}

	// Generate iteration inputs configuration - match Coze format exactly
	inputParams := g.generateInputParameters(unifiedNode, iterationConfig)

	// Generate node metadata
	nodeMeta := g.generateNodeMeta(unifiedNode)

	// Note: outputs are now generated using generateIterationOutputs method

	// Correct input structure based on Coze official example
	inputs := map[string]interface{}{
		"loopType":           "array",                                  // Top level loop type
		"loopCount":          g.generateLoopCountConfig(iterationConfig), // Top level loop count
		"variableParameters": g.generateVariableParameters(iterationConfig), // Variable parameters
		"inputParameters":    inputParams,                             // Input parameters
		// All other fields are nil to maintain Coze compatibility
		"settingOnError":     nil,
		"nodeBatchInfo":      nil,
		"llmParam":           nil,
		"outputEmitter":      nil,
		"exit":               nil,
		"llm":                nil,
		"selector":           nil,
		"textProcessor":      nil,
		"subWorkflow":        nil,
		"intentDetector":     nil,
		"databaseNode":       nil,
		"httpRequestNode":    nil,
		"knowledge":          nil,
		"codeRunner":         nil,
		"pluginApiParam":     nil,
		"variableAggregator": nil,
		"variableAssigner":   nil,
		"qa":                 nil,
		"batch":              nil,
		"comment":            nil,
		"inputReceiver":      nil,
	}

	return &CozeNode{
		ID:   cozeNodeID,
		Type: "21", // Coze iteration node type (correct type)
		Meta: &CozeNodeMeta{
			// FIXED: Add missing canvasPosition field matching Coze format
			CanvasPosition: &CozePosition{
				X: unifiedNode.Position.X * 1.5, // Apply scaling for canvas position
				Y: unifiedNode.Position.Y * 1.5,
			},
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
		Data: &CozeNodeData{
			Meta: &CozeNodeMetaInfo{
				Title:       nodeMeta["title"].(string),
				Description: nodeMeta["description"].(string),
				Icon:        nodeMeta["icon"].(string),
				SubTitle:    nodeMeta["subtitle"].(string), // FIXED: Use lowercase subtitle
				MainColor:   nodeMeta["maincolor"].(string),
			},
			Outputs: g.generateIterationOutputs(unifiedNode),
			Inputs:  inputs, // Use direct inputs structure matching Coze nodes format
			Size:    nil,
		},
		Blocks:  blocks,
		Edges:   edges,
		Version: "",
	}, nil
}

// GenerateSchemaNode generates a Coze schema iteration node
func (g *IterationNodeGenerator) GenerateSchemaNode(unifiedNode *models.Node) (*CozeSchemaNode, error) {
	if err := g.ValidateNode(unifiedNode); err != nil {
		return nil, err
	}

	cozeNodeID := g.idGenerator.MapToCozeNodeID(unifiedNode.ID)

	// FIXED: Set current iteration node ID for edge generation
	g.idGenerator.SetCurrentIterationNodeID(cozeNodeID)

    // Extract iteration configuration
    iterationConfig, ok := common.AsIterationConfig(unifiedNode.Config)
    if !ok || iterationConfig == nil {
        return nil, fmt.Errorf("invalid iteration config type for node %s", unifiedNode.ID)
    }

	// Generate sub-blocks (same as in GenerateNode)
	blocks, err := g.generateSubBlocks(iterationConfig.SubWorkflow.Nodes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sub blocks: %v", err)
	}

	// Generate internal edges (same as in GenerateNode)
	edges, err := g.generateInternalEdges(iterationConfig.SubWorkflow.Edges, iterationConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to generate internal edges: %v", err)
	}

	// Generate schema iteration inputs
	schemaInputs := g.generateSchemaIterationInputs(unifiedNode, iterationConfig)

	// Note: schema outputs are now generated using generateIterationOutputs method

	return &CozeSchemaNode{
		Data: &CozeSchemaNodeData{
			NodeMeta: &CozeNodeMetaInfo{
				Title:       g.getNodeTitle(unifiedNode),
				Description: g.getNodeDescription(unifiedNode),
				Icon:        g.getNodeIcon(),
				SubTitle:    "循环",      // Changed to match example
				MainColor:   "#00B2B2", // Changed to match example
			},
			Inputs:  schemaInputs,
			Outputs: g.generateIterationOutputs(unifiedNode),
		},
		ID:   cozeNodeID,
		Type: "21", // Coze iteration node type
		Meta: &CozeNodeMeta{
			// Add canvasPosition for schema nodes
			CanvasPosition: &CozePosition{
				X: unifiedNode.Position.X * 1.5,
				Y: unifiedNode.Position.Y * 1.5,
			},
			Position: &CozePosition{
				X: unifiedNode.Position.X,
				Y: unifiedNode.Position.Y,
			},
		},
		Blocks: blocks, // CRITICAL: Include blocks in schema node
		Edges:  edges,  // CRITICAL: Include edges in schema node
	}, nil
}

// generateSubBlocks generates sub-blocks from sub-workflow nodes based on Coze loop architecture
func (g *IterationNodeGenerator) generateSubBlocks(subNodes []models.Node) ([]interface{}, error) {
	var blocks []interface{}

	if g.nodeFactory == nil {
		return blocks, fmt.Errorf("node factory not set")
	}

	// CRITICAL: Based on Coze source code understanding - loop internal uses processing nodes directly, no need for start/end nodes
	// Internal nodes connect to loop node through special ports: loop-function-inline-output/input

	// 1. Filter out actual processing nodes (skip internal start/end nodes)
	var processingNodes []models.Node
	for _, subNode := range subNodes {
		// Based on Coze architecture, loop internal does not need independent start and end nodes
		if subNode.Type == models.NodeTypeStart || subNode.Type == models.NodeTypeEnd {
			continue
		}
		processingNodes = append(processingNodes, subNode)
	}

	// 2. Generate internal processing nodes, using correct CozeBlockNode structure
	for _, subNode := range processingNodes {
		blockNode, err := g.generateBlockNode(&subNode)
		if err != nil {
			return nil, fmt.Errorf("failed to generate block node %s: %v", subNode.ID, err)
		}

		blocks = append(blocks, blockNode)
	}

	return blocks, nil
}

// generateBlockNode generates a CozeBlockNode with correct field ordering and structure
func (g *IterationNodeGenerator) generateBlockNode(subNode *models.Node) (*CozeBlockNode, error) {
	
	// Generate original Coze node to get data
	nodeGenerator, err := g.nodeFactory.GetNodeGenerator(subNode.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get generator for node type %s: %v", subNode.Type, err)
	}

	nodeGenerator.SetIDGenerator(g.idGenerator)

	// Set iteration context for code nodes and LLM nodes
	if codeGen, ok := nodeGenerator.(*CodeNodeGenerator); ok {
		codeGen.SetIterationContext(true)
	}
	if llmGen, ok := nodeGenerator.(*LLMNodeGenerator); ok {
		llmGen.SetIterationContext(true)
	}

	// Generate original node to get data
	cozeNode, err := nodeGenerator.GenerateNode(subNode)
	if err != nil {
		return nil, fmt.Errorf("failed to generate node %s: %v", subNode.ID, err)
	}
	
	// Special handling for condition nodes to ensure correct YAML serialization
	if subNode.Type == models.NodeTypeCondition {
		return g.generateConditionBlockNode(subNode, cozeNode)
	}

	// Create CozeBlockNode with correct format
	// Field order must be: data -> id -> meta -> type
	blockNode := &CozeBlockNode{
		// ✅ data field at the front, contains: inputs -> nodeMeta -> outputs -> version
		Data: &CozeBlockNodeData{
			Inputs:   cozeNode.Data.Inputs,   // ✅ inputs first
			NodeMeta: cozeNode.Data.Meta,     // ✅ nodeMeta second (converted from Data.Meta)
			Outputs:  cozeNode.Data.Outputs,  // ✅ outputs third
			Version:  cozeNode.Version,       // ✅ version last
		},
		// ✅ id field after data
		ID: cozeNode.ID,
		// ✅ meta field after id
		Meta: cozeNode.Meta,
		// ✅ type field at the end
		Type: cozeNode.Type,
	}
	

	return blockNode, nil
}

// generateConditionBlockNode specifically generates CozeBlockNode for condition nodes, using simplified structure in schema format
func (g *IterationNodeGenerator) generateConditionBlockNode(subNode *models.Node, cozeNode *CozeNode) (*CozeBlockNode, error) {
	// CRITICAL: Condition nodes inside iteration need to use schema format, not nodes format
	// This way they can be correctly identified by coze platform and establish branch connections
	
	// Get condition node configuration
	conditionConfig, ok := subNode.Config.(*models.ConditionConfig)
	if !ok {
		return nil, fmt.Errorf("invalid condition config type")
	}
	
	// Use condition node generator to create schema format branches
	conditionGenerator := NewConditionNodeGenerator()
	conditionGenerator.SetIDGenerator(g.idGenerator)
	
	// Generate schema format branches (simplified format, no extra fields)
	branches := make([]map[string]interface{}, 0)
	for _, caseItem := range conditionConfig.Cases {
		// Skip empty condition branches (default case, level=999)
		if len(caseItem.Conditions) == 0 {
			continue
		}
		
		// Use schema format to generate branch
		branch, err := conditionGenerator.GenerateSchemaBranch(caseItem, subNode)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema branch: %w", err)
		}
		branches = append(branches, branch)
	}
	
	// Create simplified selector structure (schema format)
	conditionInputs := map[string]interface{}{
		"branches": branches, // Directly use simplified schema format branches
	}
	
	// Create CozeBlockNode with correct format, using simplified schema format
	blockNode := &CozeBlockNode{
		Data: &CozeBlockNodeData{
			Inputs:   conditionInputs, // Use simplified schema format inputs
			NodeMeta: cozeNode.Data.Meta,
			Outputs:  cozeNode.Data.Outputs,
			Version:  cozeNode.Version,
		},
		ID:   cozeNode.ID,
		Meta: cozeNode.Meta,
		Type: cozeNode.Type,
	}
	
	return blockNode, nil
}


// generateInternalEdges generates internal edges based on Coze loop architecture
func (g *IterationNodeGenerator) generateInternalEdges(subEdges []models.Edge, iterationConfig *models.IterationConfig) ([]interface{}, error) {
	var edges []interface{}

	if len(subEdges) == 0 {
		return edges, nil
	}

	// Correct loop architecture based on Coze source code understanding:
	// 1. Internal nodes connect directly to each other (no need for special entry/exit edges)
	// 2. Loop node automatically manages loop flow through special ports
	// 3. loop-function-inline-output/input ports are handled by the loop node itself

	iterationNodeID := g.idGenerator.GetCurrentIterationNodeID()
	handleMappings := g.buildHandleMappings(subEdges, iterationConfig)

	// Generate direct connections between internal processing nodes
	for _, edge := range subEdges {
		// Skip edges involving internal start/end nodes
		if g.isIterationInternalNode(edge.Source) || g.isIterationInternalNode(edge.Target) {
			continue
		}

		// Generate standard connections between internal nodes
		cozeEdge := g.generateCozeInternalEdgeWithMappings(edge, handleMappings)
		if cozeEdge != nil {
			edges = append(edges, cozeEdge)
		}
	}

	// Add loop special port connections (based on Coze official example)
	g.addCozeLoopPortConnections(&edges, subEdges, iterationNodeID)

	return edges, nil
}

// buildHandleMappings builds handle mapping table based on edge appearance order, supporting dynamic multi-branch intent recognition
func (g *IterationNodeGenerator) buildHandleMappings(subEdges []models.Edge, iterationConfig *models.IterationConfig) map[string]string {
	mappings := make(map[string]string)

	// Get classifier node configuration for dynamic multi-branch mapping
	classifierBranchMappings := g.buildDynamicClassifierBranchMappings(iterationConfig)
	
	// Track indices for different types of handles separately
	conditionBranchIndex := 0

	// CRITICAL: Process all edges in order, establish mapping from handles to Coze ports
	for _, edge := range subEdges {
		handle := edge.SourceHandle
		if handle == "" {
			continue // Skip empty handles
		}

		// Handle intent recognition handles - use dynamic mapping, ensure each branch has unique mapping
		if strings.HasPrefix(handle, "intent-one-of::") {
			if branchName, exists := classifierBranchMappings[handle]; exists {
				mappings[handle] = branchName
			} else {
				// CRITICAL: Avoid multiple branches mapping to same port
				// Record unmapped handles for subsequent dynamic analysis
				fmt.Printf("⚠️  Classifier branch mapping not found: %s\n", handle)
				mappings[handle] = "default"
			}
		}

		// Handle condition branch handles using branch mapping logic
		if strings.HasPrefix(handle, "branch_one_of::") {
			if _, exists := mappings[handle]; !exists {
				// Use completely dynamic condition branch mapping logic, supporting any number of branches
				mappedPort := g.mapConditionBranchToCozePort(handle, subEdges, iterationConfig, conditionBranchIndex)
				mappings[handle] = mappedPort
				conditionBranchIndex++
			}
		}

		// Handle known fixed handle types
		if handle == "true" || handle == "false" || handle == "default" {
			mappings[handle] = handle // Keep original value
		}
	}

	return mappings
}

// mapConditionBranchToCozePort dynamically maps condition branches to ports based on Coze source calcPortId rules
// Rules: 1st branch="true", 2nd="true_1", Nth="true_{N-1}", default="false" 
func (g *IterationNodeGenerator) mapConditionBranchToCozePort(handle string, subEdges []models.Edge, iterationConfig *models.IterationConfig, branchIndex int) string {
	// 1. Find condition node configuration
	var conditionConfig *models.ConditionConfig
	for _, node := range iterationConfig.SubWorkflow.Nodes {
		if node.Type == models.NodeTypeCondition {
			if config, ok := node.Config.(*models.ConditionConfig); ok {
				conditionConfig = config
				break
			}
		}
	}
	
	if conditionConfig == nil {
		return "default"
	}
	
	// 2. Collect all non-default branches (level != 999) and sort by level
	type BranchInfo struct {
		CaseID string
		Level  int
	}
	
	var nonDefaultBranches []BranchInfo
	var isDefaultBranch bool
	
	for _, caseItem := range conditionConfig.Cases {
		if caseItem.CaseID == handle {
			if caseItem.Level == 999 {
				isDefaultBranch = true
				break
			}
		}
		
		// Collect all non-default branches
		if caseItem.Level != 999 {
			nonDefaultBranches = append(nonDefaultBranches, BranchInfo{
				CaseID: caseItem.CaseID,
				Level:  caseItem.Level,
			})
		}
	}
	
	// Return false port for default branches
	if isDefaultBranch {
		return "false"
	}
	
	// Sort non-default branches by level in ascending order
	for i := 0; i < len(nonDefaultBranches)-1; i++ {
		for j := i + 1; j < len(nonDefaultBranches); j++ {
			if nonDefaultBranches[i].Level > nonDefaultBranches[j].Level {
				nonDefaultBranches[i], nonDefaultBranches[j] = nonDefaultBranches[j], nonDefaultBranches[i]
			}
		}
	}
	
	// Generate port IDs based on Coze calcPortId rules
	// index=0 -> "true", index=1 -> "true_1", index=N -> "true_N"
	for index, branch := range nonDefaultBranches {
		if branch.CaseID == handle {
			if index == 0 {
				return "true"
			} else {
				return fmt.Sprintf("true_%d", index)
			}
		}
	}
	
	// Fallback logic
	return "default"
}

// generateCozeInternalEdgeWithMappings generates Coze format edge connections using mapping table
func (g *IterationNodeGenerator) generateCozeInternalEdgeWithMappings(edge models.Edge, mappings map[string]string) map[string]interface{} {
	sourceNodeID := g.idGenerator.MapToCozeNodeID(edge.Source)
	targetNodeID := g.idGenerator.MapToCozeNodeID(edge.Target)

	// CRITICAL: Use mapping table to convert source port, unmapped UUID handles use default
	sourcePortID := ""
	if edge.SourceHandle != "" {
		if mappedPort, exists := mappings[edge.SourceHandle]; exists {
			sourcePortID = mappedPort
		} else {
			// FIXED: UUID handles cannot be recognized by Coze, use default as fallback
			sourcePortID = "default"
		}
	}

	// Target port is usually empty (characteristic of iteration internal edges)
	targetPortID := ""

	// CRITICAL: Use Coze official lowercase naming format (based on source code analysis)
	return map[string]interface{}{
		"sourceNodeID": sourceNodeID, // Keep camelCase naming for schema section
		"targetNodeID": targetNodeID,
		"sourcePortID": sourcePortID,
		"targetPortID": targetPortID,
	}
}

// addCozeLoopPortConnections adds loop port connections based on Coze official examples
func (g *IterationNodeGenerator) addCozeLoopPortConnections(edges *[]interface{}, subEdges []models.Edge, iterationNodeID string) {
	// Based on Coze official example loop_with_object_input.json connection pattern:
	// 1. Loop node -> Internal first node (loop-function-inline-output)
	// 2. Internal last node -> Loop node (loop-function-inline-input)

	var firstNodeID, lastNodeID string

	// Identify first and last processing nodes
	for _, edge := range subEdges {
		// Find the first processing node connected from internal start node
		if g.isIterationInternalNode(edge.Source) && !g.isIterationInternalNode(edge.Target) {
			if firstNodeID == "" {
				firstNodeID = g.idGenerator.MapToCozeNodeID(edge.Target)
			}
		}
		// Find the last processing node connected to internal end node
		if !g.isIterationInternalNode(edge.Source) && g.isIterationInternalNode(edge.Target) {
			lastNodeID = g.idGenerator.MapToCozeNodeID(edge.Source)
		}
	}

	// Add loop entry connection (Coze official format)
	if firstNodeID != "" {
		entryEdge := map[string]interface{}{
			"sourceNodeID": iterationNodeID,
			"targetNodeID": firstNodeID,
			"sourcePortID": "loop-function-inline-output",
		}
		*edges = append(*edges, entryEdge)
	}

	// Add loop exit connection (Coze official format)
	if lastNodeID != "" {
		exitEdge := map[string]interface{}{
			"sourceNodeID": lastNodeID,
			"targetNodeID": iterationNodeID,
			"targetPortID": "loop-function-inline-input",
		}
		*edges = append(*edges, exitEdge)
	}
}



// isIterationInternalNode checks if a node ID represents an iteration internal node (start/end)
func (g *IterationNodeGenerator) isIterationInternalNode(nodeID string) bool {
	return strings.Contains(nodeID, "iteration-node-start") || strings.Contains(nodeID, "iteration-node-end")
}








// generateLoopCountConfig generates loop count configuration matching Coze format
func (g *IterationNodeGenerator) generateLoopCountConfig(iterConfig *models.IterationConfig) map[string]interface{} {
	// Match exact format from Coze official example
	return map[string]interface{}{
		"type": "integer", // lowercase 'type'
		"value": map[string]interface{}{ // lowercase 'value'
			"content": "10",      // Default loop count
			"type":    "literal", // type field first
		},
	}
}

// generateVariableParameters generates variable parameters for iteration
func (g *IterationNodeGenerator) generateVariableParameters(iterConfig *models.IterationConfig) []interface{} {
	// CRITICAL: Coze loop nodes must have variableParameters to work properly!
	// According to official examples, need to define loop variables to provide internal context

	// Generate standard loop variable parameters, matching official example format
	variableParam := map[string]interface{}{
		"input": map[string]interface{}{
			"type": "string", // Loop variable type
			"value": map[string]interface{}{
				"content": "init", // Initial value, matching official example
				"rawMeta": map[string]interface{}{
					"type": 1, // String type rawMeta code
				},
				"type": "literal", // Literal value type
			},
		},
		"name": "variable", // Variable name, matching official example
	}

	return []interface{}{variableParam}
}

// generateInputParameters generates input parameters matching exact Coze format
func (g *IterationNodeGenerator) generateInputParameters(unifiedNode *models.Node, iterConfig *models.IterationConfig) []map[string]interface{} {
	var inputParams []map[string]interface{}

	if unifiedNode.Inputs == nil || len(unifiedNode.Inputs) == 0 {
		return inputParams
	}

	// Get the first input as the iteration source
	input := unifiedNode.Inputs[0]
	if input.Reference == nil {
		return inputParams
	}

	// Map source node ID to Coze ID
	sourceBlockID := g.idGenerator.MapToCozeNodeID(input.Reference.NodeID)

	// Generate input parameter matching EXACT Coze official format
	// Based on loop_selector_variable_assign_text_processor.json line 432-449
	param := map[string]interface{}{
		"input": map[string]interface{}{
			"schema": map[string]interface{}{ // CRITICAL: Must have schema field!
				"type": g.mapUnifiedTypeToCozeSchemaType(input.Type),
			},
			"type": "list", // Coze requires "list" type for array inputs
			"value": map[string]interface{}{ // Note: lowercase 'value'
				"content": map[string]interface{}{
					"blockID": sourceBlockID,
					"name":    g.mapOutputFieldNameForCoze(input.Reference.NodeID, input.Reference.OutputName),
					"source":  "block-output",
				},
				"type": "ref", // Note: 'type' at this level
				// Note: No rawMeta field in official example for inputParameters value
			},
		},
		"name": input.Name, // 'name' comes after 'input'
	}
	inputParams = append(inputParams, param)

	return inputParams
}






// generateNodeMeta generates node metadata matching exact Coze format
func (g *IterationNodeGenerator) generateNodeMeta(unifiedNode *models.Node) map[string]interface{} {
	title := unifiedNode.Title
	if title == "" {
		title = "循环" // Use exact text from Coze example
	}

	description := unifiedNode.Description
	if description == "" {
		description = "用于通过设定循环次数和逻辑，重复执行一系列任务" // Exact text from Coze example
	}

	return map[string]interface{}{
		"title":       title,
		"description": description,
		"icon":        "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Loop-v2.jpg",
		"subtitle":    "循环", // lowercase 's' and correct text
		"maincolor":   "#00B2B2",
	}
}

// getLastProcessingNodeID gets the ID of the last actual processing node in the iteration
func (g *IterationNodeGenerator) getLastProcessingNodeID(unifiedNode *models.Node) string {
    iterationConfig, ok := common.AsIterationConfig(unifiedNode.Config)
    if !ok || iterationConfig == nil {
        return g.idGenerator.MapToCozeNodeID(unifiedNode.ID)
    }

	// CRITICAL: Need to analyze edge connections to find the real last processing node
	// Cannot simply search from back to front, because node order does not represent execution order
	
	// 1. Build internal node indegree mapping (find nodes with no subsequent connections)
	internalNodeIDs := make(map[string]bool)
	outgoingEdges := make(map[string][]models.Edge)
	incomingEdges := make(map[string][]models.Edge)
	
	// Collect all internal processing nodes (skip start and end nodes)
	for _, subNode := range iterationConfig.SubWorkflow.Nodes {
		if subNode.Type != models.NodeTypeStart && subNode.Type != models.NodeTypeEnd {
			internalNodeIDs[subNode.ID] = true
			outgoingEdges[subNode.ID] = []models.Edge{}
			incomingEdges[subNode.ID] = []models.Edge{}
		}
	}
	
	// Analyze edge connection relationships
	for _, edge := range iterationConfig.SubWorkflow.Edges {
		// Only consider connections between internal processing nodes
		if internalNodeIDs[edge.Source] && internalNodeIDs[edge.Target] {
			outgoingEdges[edge.Source] = append(outgoingEdges[edge.Source], edge)
			incomingEdges[edge.Target] = append(incomingEdges[edge.Target], edge)
		}
	}
	
	// 2. Find nodes that do not output to other internal nodes (terminal nodes)
	var terminalNodes []string
	for nodeID := range internalNodeIDs {
		hasInternalOutgoing := false
		for _, edge := range outgoingEdges[nodeID] {
			if internalNodeIDs[edge.Target] {
				hasInternalOutgoing = true
				break
			}
		}
		if !hasInternalOutgoing {
			terminalNodes = append(terminalNodes, nodeID)
		}
	}
	
	// 3. If there are multiple terminal nodes, choose code node as final output node
	for _, nodeID := range terminalNodes {
		for _, subNode := range iterationConfig.SubWorkflow.Nodes {
			if subNode.ID == nodeID && subNode.Type == models.NodeTypeCode {
				return g.idGenerator.MapToCozeNodeID(nodeID)
			}
		}
	}
	
	// 4. If there are no code nodes, choose the first terminal node
	if len(terminalNodes) > 0 {
		return g.idGenerator.MapToCozeNodeID(terminalNodes[0])
	}
	
	// 5. Fallback logic: search from back to front for first non-start/end node
	for i := len(iterationConfig.SubWorkflow.Nodes) - 1; i >= 0; i-- {
		subNode := iterationConfig.SubWorkflow.Nodes[i]
		if subNode.Type != models.NodeTypeEnd && subNode.Type != models.NodeTypeStart {
			return g.idGenerator.MapToCozeNodeID(subNode.ID)
		}
	}

	// Final fallback to iteration node itself
	return g.idGenerator.MapToCozeNodeID(unifiedNode.ID)
}

// getLastProcessingNodeOutputName gets output field name of final processing node - completely dynamic version without hardcoding
func (g *IterationNodeGenerator) getLastProcessingNodeOutputName(unifiedNode *models.Node, lastProcessingNodeID string) string {
    iterationConfig, ok := common.AsIterationConfig(unifiedNode.Config)
    if !ok || iterationConfig == nil {
        return "result"
    }
	
	// CRITICAL: Find corresponding unified DSL node, get its actual output definition
	for _, subNode := range iterationConfig.SubWorkflow.Nodes {
		cozeNodeID := g.idGenerator.MapToCozeNodeID(subNode.ID)
		if cozeNodeID == lastProcessingNodeID {
			
			// 1. Prioritize using explicitly defined output field names in unified DSL
			if len(subNode.Outputs) > 0 && subNode.Outputs[0].Name != "" {
				return subNode.Outputs[0].Name
			}
			
			// 2. If no explicit output definition, try to get from iteration node's output selector configuration
			if iterationConfig.OutputSelector.NodeID == subNode.ID && iterationConfig.OutputSelector.OutputName != "" {
				return iterationConfig.OutputSelector.OutputName
			}
			
			// 3. Dynamically analyze node configuration, infer main output field
			return g.inferMainOutputFieldName(subNode)
		}
	}
	
	// 4. If none found, use generic fallback value
	return "result"
}

// inferMainOutputFieldName dynamically infers main output field name based on node configuration
func (g *IterationNodeGenerator) inferMainOutputFieldName(node models.Node) string {
	switch node.Type {
    case models.NodeTypeCode:
        // Code node: analyze return statements in code to infer output fields
        if codeConfig, ok := common.AsCodeConfig(node.Config); ok && codeConfig != nil {
			// Simple regex matching to find field names in return statements
			// Example: return {"integrated_result": result} or return {"output": data}
			outputField := g.extractOutputFieldFromCode(codeConfig.Code)
			if outputField != "" {
				return outputField
			}
		}
		
    case models.NodeTypeLLM:
        // LLM node: check prompt template to infer output usage
        if llmConfig, ok := common.AsLLMConfig(node.Config); ok && llmConfig != nil {
			// Analyze system prompt to infer output type
			outputField := g.inferLLMOutputField(llmConfig.Prompt.SystemTemplate)
			if outputField != "" {
				return outputField
			}
		}
		
	case models.NodeTypeClassifier:
		// Classifier node: use standard classification output field
		return "classificationId"
		
	case models.NodeTypeCondition:
		// Condition node: usually no data output, but if there is, use result
		return "result"
	}
	
	// Final generic fallback: examine node title or description, infer possible output field name
	return g.inferOutputFromNodeMeta(node)
}

// extractOutputFieldFromCode extracts main field name returned from code
func (g *IterationNodeGenerator) extractOutputFieldFromCode(code string) string {
	// Simple string matching, looking for fields in return statements
	// Implement a simple pattern matching here
	
	// Match return {"field_name": ...} pattern
	returnPattern := `return\s*\{\s*"([^"]+)"`
	if matches := regexp.MustCompile(returnPattern).FindStringSubmatch(code); len(matches) > 1 {
		return matches[1]
	}
	
	// If not found, return empty string for upper layer to continue trying other methods
	return ""
}

// inferLLMOutputField infers output field usage based on LLM prompt
func (g *IterationNodeGenerator) inferLLMOutputField(systemPrompt string) string {
	// Infer output field name based on prompt content
	// This is a heuristic method that can be extended based on actual situations
	
	if strings.Contains(systemPrompt, "指导") || strings.Contains(systemPrompt, "guidance") {
		return "guidance"
	}
	
	if strings.Contains(systemPrompt, "分析") || strings.Contains(systemPrompt, "analysis") {
		return "analysis"
	}
	
	if strings.Contains(systemPrompt, "建议") || strings.Contains(systemPrompt, "suggestion") {
		return "suggestion"
	}
	
	// Default return empty, let upper layer continue processing
	return ""
}

// inferOutputFromNodeMeta infers output field based on node meta information
func (g *IterationNodeGenerator) inferOutputFromNodeMeta(node models.Node) string {
	// Analyze node title, infer possible output field
	title := strings.ToLower(node.Title)
	
	if strings.Contains(title, "整合") || strings.Contains(title, "集成") {
		return "integrated_result"
	}
	
	if strings.Contains(title, "处理") || strings.Contains(title, "process") {
		return "processed_result"
	}
	
	if strings.Contains(title, "分析") || strings.Contains(title, "analysis") {
		return "analysis_result"
	}
	
	if strings.Contains(title, "生成") || strings.Contains(title, "generate") {
		return "generated_content"
	}
	
	// Final fallback to generic field name
	return "result"
}

// generateIterationOutputs generates complex reference outputs matching exact Coze format
func (g *IterationNodeGenerator) generateIterationOutputs(unifiedNode *models.Node) []map[string]interface{} {
	var outputs []map[string]interface{}

	// FIXED: Reference the last actual processing node, not a virtual end node
	lastProcessingNodeID := g.getLastProcessingNodeID(unifiedNode)

	// CRITICAL: According to correct example, iteration output field name must be result_list
	outputName := "result_list" // Must use result_list to match correct example
	if unifiedNode.Outputs != nil && len(unifiedNode.Outputs) > 0 {
		// If unified DSL explicitly specifies output name, use it first
		if unifiedNode.Outputs[0].Name != "" {
			outputName = unifiedNode.Outputs[0].Name
		}
	}

	// According to correct Coze example, iteration node only needs one output!
	// Removed variable_out, only keep main list output

	// output: output (list type)
	listOutput := map[string]interface{}{
		"input": map[string]interface{}{
			"type": "list", // List type for collected results
			"schema": map[string]interface{}{
				"type": "string", // Element type
			},
			"value": map[string]interface{}{
				"content": map[string]interface{}{
					"blockID": lastProcessingNodeID,                                        // Reference actual last processing node
					"name":    g.getLastProcessingNodeOutputName(unifiedNode, lastProcessingNodeID), // CRITICAL: Dynamically get actual output field name of final node
					"source":  "block-output",
				},
				"rawMeta": map[string]interface{}{
					"type": 1, // Keep as 1 for internal block references
				},
				"type": "ref",
			},
		},
		"name": outputName, // Use dynamic output name
	}
	outputs = append(outputs, listOutput)

	return outputs
}





// mapUnifiedTypeToCozeSchemaType maps unified data type to Coze schema type string
func (g *IterationNodeGenerator) mapUnifiedTypeToCozeSchemaType(unifiedType models.UnifiedDataType) string {
	switch unifiedType {
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "float"
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeArrayString:
		return "string" // Schema type for array elements
	case models.DataTypeArrayObject:
		return "object" // Schema type for array elements
	case models.DataTypeObject:
		return "object"
	default:
		return "string"
	}
}







// generateSchemaIterationInputs generates schema iteration inputs
func (g *IterationNodeGenerator) generateSchemaIterationInputs(unifiedNode *models.Node, iterConfig *models.IterationConfig) map[string]interface{} {
	// Generate schema input parameters with proper RawMeta format
	schemaInputParams := g.generateSchemaInputParameters(unifiedNode)

	// Generate error handling settings for schema
	errorSettings := g.generateSchemaErrorSettings()

	// FIXED: Use the same top-level structure as nodes format
	return map[string]interface{}{
		"inputParameters":    schemaInputParams,                        // camelCase
		"loopCount":          g.generateLoopCountConfig(iterConfig),    // Top level
		"loopType":           "array",                                  // Top level
		"variableParameters": g.generateVariableParameters(iterConfig), // Top level
		"settingOnError":     errorSettings,                            // camelCase
		// Required null fields for Coze compatibility
		"batch":              nil,
		"comment":            nil,
		"databasenode":       nil,
		"exit":               nil,
		"httprequestnode":    nil,
		"inputreceiver":      nil,
		"intentdetector":     nil,
		"knowledge":          nil,
		"llm":                nil,
		"llmparam":           nil,
		"nodebatchinfo":      nil,
		"outputemitter":      nil,
		"pluginapiparam":     nil,
		"qa":                 nil,
		"selector":           nil,
		"subworkflow":        nil,
		"textprocessor":      nil,
		"variableaggregator": nil,
		"variableassigner":   nil,
	}
}

// generateSchemaInputParameters generates schema input parameters with proper RawMeta format
func (g *IterationNodeGenerator) generateSchemaInputParameters(unifiedNode *models.Node) []CozeInputParameter {
	var schemaInputParams []CozeInputParameter

	if unifiedNode.Inputs == nil {
		return schemaInputParams
	}

	for _, input := range unifiedNode.Inputs {
		// Use schema format consistent with Coze official examples
		if input.Reference != nil && input.Reference.Type == models.ReferenceTypeNodeOutput {
			schemaInput := CozeInputParameter{
				Name: input.Name,
				Input: &CozeInputValue{
					Type: "list", // Coze uses "list" type for array inputs
					Schema: &CozeInputSchema{
						Type: g.mapUnifiedTypeToCozeSchemaType(input.Type),
					},
					Value: &CozeInputRef{
						Type: "ref",
						Content: &CozeRefContent{
							BlockID: g.idGenerator.MapToCozeNodeID(input.Reference.NodeID),
							Name:    input.Reference.OutputName,
							Source:  "block-output",
						},
					},
				},
			}
			schemaInputParams = append(schemaInputParams, schemaInput)
		}
	}

	return schemaInputParams
}

// generateSchemaErrorSettings generates error handling settings for schema
func (g *IterationNodeGenerator) generateSchemaErrorSettings() map[string]interface{} {
	return map[string]interface{}{
		"processType": 1,
		"retryTimes":  0,
		"timeoutMs":   300000,
	}
}



// getNodeTitle returns node title with uniqueness
func (g *IterationNodeGenerator) getNodeTitle(unifiedNode *models.Node) string {
	if unifiedNode.Title != "" {
		return unifiedNode.Title
	}
	return "循环"
}

// getNodeDescription returns node description
func (g *IterationNodeGenerator) getNodeDescription(unifiedNode *models.Node) string {
	if unifiedNode.Description != "" {
		return unifiedNode.Description
	}
	return "用于通过设定循环次数和逻辑，重复执行一系列任务"
}

// getNodeIcon returns iteration node icon
func (g *IterationNodeGenerator) getNodeIcon() string {
	return "https://lf3-static.bytednsdoc.com/obj/eden-cn/dvsmryvd_avi_dvsm/ljhwZthlaukjlkulzlp/icon/icon-Loop-v2.jpg"
}

// mapOutputFieldNameForCoze maps output field names from unified DSL to Coze platform format
func (g *IterationNodeGenerator) mapOutputFieldNameForCoze(nodeID, outputName string) string {
	// Map classifier output names from iFlytek format to Coze format
	if outputName == "class_name" {
		// iFlytek classifier outputs "class_name", but Coze uses "classificationId"
		return "classificationId"
	}

	// Default: return original name if no mapping needed
	return outputName
}

// buildDynamicClassifierBranchMappings builds truly dynamic branch mappings based on original iFlytek DSL edge definitions
func (g *IterationNodeGenerator) buildDynamicClassifierBranchMappings(iterationConfig *models.IterationConfig) map[string]string {
	mappings := make(map[string]string)
	
	// Step 1: Build mapping table from intent ID to intent information
	intentIDToInfo := make(map[string]models.ClassifierClass)
	classifierNodeIDs := make(map[string]bool)
	
	for _, node := range iterationConfig.SubWorkflow.Nodes {
		if node.Type == models.NodeTypeClassifier {
			classifierNodeIDs[node.ID] = true
			if classifierConfig, ok := node.Config.(*models.ClassifierConfig); ok {
				for _, class := range classifierConfig.Classes {
					if class.ID != "" {
						intentIDToInfo[class.ID] = class
					}
				}
			}
		}
	}
	
	// Collect intent IDs from classifier output edges from edge definitions, maintaining original order in edge definitions
	intentIDsInOrder := []string{}
	for _, edge := range iterationConfig.SubWorkflow.Edges {
		if classifierNodeIDs[edge.Source] && edge.SourceHandle != "" {
			// Avoid recording the same intent ID repeatedly
			found := false
			for _, existingID := range intentIDsInOrder {
				if existingID == edge.SourceHandle {
					found = true
					break
				}
			}
			if !found {
				intentIDsInOrder = append(intentIDsInOrder, edge.SourceHandle)
			}
		}
	}
	
	// Step 3: Map based on actual intent information
	branchIndex := 0
	for _, intentID := range intentIDsInOrder {
		if intentInfo, exists := intentIDToInfo[intentID]; exists {
			// Judge whether it is default intent based on actual name and description
			isDefault := (intentInfo.Name == "default" && strings.Contains(intentInfo.Description, "默认")) ||
			             strings.Contains(intentInfo.Description, "默认意图") ||
			             strings.Contains(strings.ToLower(intentInfo.Name), "default") ||
			             strings.Contains(strings.ToLower(intentInfo.Name), "其他") ||
			             strings.Contains(strings.ToLower(intentInfo.Name), "fallback")
			
			if isDefault {
				mappings[intentID] = "default"
			} else {
				mappings[intentID] = fmt.Sprintf("branch_%d", branchIndex)
				branchIndex++
			}
		} else {
			// If intent information not found, assign by order (insurance measure)
			mappings[intentID] = fmt.Sprintf("branch_%d", branchIndex)
			branchIndex++
		}
	}
	
	return mappings
}






