package parser

import (
	"ai-agents-transformer/core/interfaces"
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Compile-time interface check
var _ interfaces.DSLParser = (*IFlytekParser)(nil)

// IFlytekParser provides DSL parsing for iFlytek Agent platform
type IFlytekParser struct {
	*common.BaseParser
	variableRefSystem *models.VariableReferenceSystem
	parserFactory     *ParserFactory
	edgeParser        *EdgeParser
	referenceParser   *ReferenceParser
	// Global node output type mapping table for condition node type inference
	nodeOutputTypeMap map[string]map[string]models.UnifiedDataType // nodeID -> outputName -> dataType
}

func NewIFlytekParser() *IFlytekParser {
	vrs := models.NewVariableReferenceSystem()
	return &IFlytekParser{
		BaseParser:        common.NewBaseParser(models.PlatformIFlytek),
		variableRefSystem: vrs,
		parserFactory:     NewParserFactory(),
		edgeParser:        NewEdgeParser(vrs),
		referenceParser:   NewReferenceParser(vrs),
		nodeOutputTypeMap: make(map[string]map[string]models.UnifiedDataType),
	}
}

// Parse parses DSL data into unified format
func (p *IFlytekParser) Parse(data []byte) (*models.UnifiedDSL, error) {
	var root IFlytekRootStructure
	if err := yaml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	unifiedDSL := models.NewUnifiedDSL()

	// Parse metadata
	if err := p.parseMetadata(root.FlowMeta, unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Parse nodes
	if err := p.parseNodes(root.FlowData.Nodes, unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to parse nodes: %w", err)
	}

	// Parse edges
	if err := p.parseEdges(root.FlowData.Edges, unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to parse edges: %w", err)
	}

	// After parsing all edges, organize iteration internal connections
	if err := p.organizeIterationInternalEdges(unifiedDSL); err != nil {
		return nil, fmt.Errorf("failed to organize iteration internal edges: %w", err)
	}

	// Print conversion summary after parsing is complete
	p.printConversionSummary(unifiedDSL)

	return unifiedDSL, nil
}

// ParseFile parses DSL from file
func (p *IFlytekParser) ParseFile(filename string) (*models.UnifiedDSL, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	return p.Parse(data)
}

// Validate validates input data
func (p *IFlytekParser) Validate(data []byte) error {
	var root IFlytekRootStructure
	if err := yaml.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	return p.validateStructure(root)
}

// parseMetadata parses flow metadata
func (p *IFlytekParser) parseMetadata(flowMeta IFlytekFlowMeta, unifiedDSL *models.UnifiedDSL) error {
	// Validate required fields
	if flowMeta.Name == "" {
		return fmt.Errorf("flow name is required")
	}

	// Set basic metadata
	unifiedDSL.Metadata.Name = flowMeta.Name
	unifiedDSL.Metadata.Description = flowMeta.Description

	// Parse UI configuration
	uiConfig, err := p.parseUIConfig(flowMeta)
	if err != nil {
		return fmt.Errorf("failed to parse UI config: %w", err)
	}
	unifiedDSL.Metadata.UIConfig = uiConfig

	// Set iFlytek Platform specific metadata
	unifiedDSL.PlatformMetadata.IFlytek = &models.IFlytekMetadata{
		AvatarIcon:     flowMeta.AvatarIcon,
		AvatarColor:    flowMeta.AvatarColor,
		AdvancedConfig: flowMeta.AdvancedConfig,
		DSLVersion:     flowMeta.DSLVersion,
	}

	return nil
}

// parseUIConfig parses UI configuration.
func (p *IFlytekParser) parseUIConfig(flowMeta IFlytekFlowMeta) (*models.UIConfig, error) {
	if !p.hasAdvancedConfig(flowMeta) {
		return nil, nil
	}

	advancedConfig := p.parseAdvancedConfigJSON(flowMeta.AdvancedConfig)
	if advancedConfig == nil {
		return nil, nil
	}

	uiConfig := &models.UIConfig{}

	p.parsePrologueConfig(advancedConfig, uiConfig)
	p.parseIconConfig(flowMeta, uiConfig)

	if p.isUIConfigEmpty(uiConfig) {
		return nil, nil
	}

	return uiConfig, nil
}

func (p *IFlytekParser) hasAdvancedConfig(flowMeta IFlytekFlowMeta) bool {
	return flowMeta.AdvancedConfig != ""
}

func (p *IFlytekParser) parseAdvancedConfigJSON(advancedConfigStr string) map[string]interface{} {
	var advancedConfig map[string]interface{}
	if err := json.Unmarshal([]byte(advancedConfigStr), &advancedConfig); err != nil {
		return nil
	}
	return advancedConfig
}

func (p *IFlytekParser) parsePrologueConfig(advancedConfig map[string]interface{}, uiConfig *models.UIConfig) {
	prologueInterface, exists := advancedConfig["prologue"]
	if !exists {
		return
	}

	prologue, ok := prologueInterface.(map[string]interface{})
	if !ok {
		return
	}

	p.parseOpeningStatement(prologue, uiConfig)
	p.parseSuggestedQuestions(prologue, uiConfig)
}

func (p *IFlytekParser) parseOpeningStatement(prologue map[string]interface{}, uiConfig *models.UIConfig) {
	enabled, ok := prologue["enabled"].(bool)
	if !ok || !enabled {
		return
	}

	statement, ok := prologue["statement"].(string)
	if ok && statement != "" {
		uiConfig.OpeningStatement = statement
	}
}

func (p *IFlytekParser) parseSuggestedQuestions(prologue map[string]interface{}, uiConfig *models.UIConfig) {
	inputExampleInterface, exists := prologue["inputExample"]
	if !exists {
		return
	}

	inputExampleArray, ok := inputExampleInterface.([]interface{})
	if !ok {
		return
	}

	suggestedQuestions := p.extractSuggestedQuestions(inputExampleArray)
	if len(suggestedQuestions) > 0 {
		uiConfig.SuggestedQuestions = suggestedQuestions
	}
}

func (p *IFlytekParser) extractSuggestedQuestions(inputExampleArray []interface{}) []string {
	var suggestedQuestions []string
	for _, item := range inputExampleArray {
		if question, ok := item.(string); ok && question != "" {
			suggestedQuestions = append(suggestedQuestions, question)
		}
	}
	return suggestedQuestions
}

func (p *IFlytekParser) parseIconConfig(flowMeta IFlytekFlowMeta, uiConfig *models.UIConfig) {
	if flowMeta.AvatarIcon != "" {
		uiConfig.Icon = flowMeta.AvatarIcon
	}
	if flowMeta.AvatarColor != "" {
		uiConfig.IconBackground = flowMeta.AvatarColor
	}
}

func (p *IFlytekParser) isUIConfigEmpty(uiConfig *models.UIConfig) bool {
	return uiConfig.OpeningStatement == "" &&
		len(uiConfig.SuggestedQuestions) == 0 &&
		uiConfig.Icon == "" &&
		uiConfig.IconBackground == ""
}

// parseNodes parses nodes.
func (p *IFlytekParser) parseNodes(nodes []IFlytekNode, unifiedDSL *models.UnifiedDSL) error {
	// Step 1: Parse all nodes
	allNodes, nodeParentMap, err := p.parseAllNodes(nodes)
	if err != nil {
		return err
	}

	// Step 2: Organize node hierarchy
	if err := p.organizeNodeHierarchy(allNodes, nodeParentMap, unifiedDSL); err != nil {
		return fmt.Errorf("failed to organize node hierarchy: %w", err)
	}

	return nil
}

// parseAllNodes parses all iFlytek nodes to unified nodes
func (p *IFlytekParser) parseAllNodes(nodes []IFlytekNode) ([]*models.Node, map[string]string, error) {
	allNodes := make([]*models.Node, 0, len(nodes))
	nodeParentMap := make(map[string]string) // nodeID -> parentID

	for _, iflytekNode := range nodes {
		node, err := p.parseIndividualNode(iflytekNode)
		if err != nil {
			return nil, nil, err
		}

		// All nodes are now converted, including unsupported nodes (converted to code nodes)
		allNodes = append(allNodes, node)

		// Record parent-child relationships
		if iflytekNode.ParentID != "" {
			nodeParentMap[node.ID] = iflytekNode.ParentID
		}
	}

	return allNodes, nodeParentMap, nil
}

// parseIndividualNode parses a single iFlytek node to unified node
func (p *IFlytekParser) parseIndividualNode(iflytekNode IFlytekNode) (*models.Node, error) {
	// Convert to iflytek parser type
	iflytekNodeConverted := p.convertToParserType(iflytekNode)

	// Use fallback parsing to handle unsupported node types
	node, supported, err := p.parserFactory.ParseNodeWithFallback(iflytekNodeConverted, p.variableRefSystem, p)
	if err != nil {
		return nil, fmt.Errorf("failed to parse node %s: %w", iflytekNode.ID, err)
	}

	if !supported {
		// Convert unsupported nodes to code node placeholders
		fmt.Printf("⚠️  Converting unsupported node type '%s' (ID: %s) to code node placeholder\n",
			iflytekNode.Type, iflytekNode.ID)

		return p.convertUnsupportedNodeToCodeNode(iflytekNode)
	}

	// Update node output type mapping table
	p.updateNodeOutputTypeMapping(node)

	return node, nil
}

// convertUnsupportedNodeToCodeNode converts unsupported nodes to code node placeholders
func (p *IFlytekParser) convertUnsupportedNodeToCodeNode(iflytekNode IFlytekNode) (*models.Node, error) {
	// Get node label for type description
	nodeLabel := p.extractNodeLabel(iflytekNode)

	// Use code node parser to create code node
	codeParser := NewCodeNodeParser(p.variableRefSystem)

	// Create modified node with code node type
	modifiedNode := iflytekNode
	modifiedNode.Type = IFlytekNodeTypeCode // "code"

	// Modify node title
	if modifiedNode.Data == nil {
		modifiedNode.Data = make(map[string]interface{})
	}
	modifiedNode.Data["label"] = fmt.Sprintf("暂不兼容的节点-%s（请根据需求手动实现）", nodeLabel)

	// Set default code configuration
	if modifiedNode.Data["nodeParam"] == nil {
		modifiedNode.Data["nodeParam"] = make(map[string]interface{})
	}
	nodeParam := modifiedNode.Data["nodeParam"].(map[string]interface{})
	nodeParam["code"] = fmt.Sprintf(`# 抱歉！当前兼容性工具不支持转换此类节点: %s

# 请根据业务需求手动补充实现逻辑
`, iflytekNode.Type)

	// Create default output if none exist to maintain connections
	if _, hasOutputs := modifiedNode.Data["outputs"]; !hasOutputs {
		modifiedNode.Data["outputs"] = []interface{}{
			map[string]interface{}{
				"name": "result",
				"schema": map[string]interface{}{
					"type":    "string",
					"default": "",
				},
				"required": false,
			},
		}
	}

	// Parse using code node parser
	return codeParser.ParseNode(modifiedNode)
}

// extractNodeLabel extracts node label
func (p *IFlytekParser) extractNodeLabel(iflytekNode IFlytekNode) string {
	if data, ok := iflytekNode.Data["label"].(string); ok && data != "" {
		return data
	}
	return iflytekNode.Type
}

// updateNodeOutputTypeMapping updates node output type mapping table
func (p *IFlytekParser) updateNodeOutputTypeMapping(node *models.Node) {
	if len(node.Outputs) == 0 {
		return
	}

	if p.nodeOutputTypeMap[node.ID] == nil {
		p.nodeOutputTypeMap[node.ID] = make(map[string]models.UnifiedDataType)
	}

	for _, output := range node.Outputs {
		p.nodeOutputTypeMap[node.ID][output.Name] = output.Type
	}
}

// GetOutputType retrieves data type for specified node output (for condition node parsers)
func (p *IFlytekParser) GetOutputType(nodeID, outputName string) models.UnifiedDataType {
	if nodeOutputs, exists := p.nodeOutputTypeMap[nodeID]; exists {
		if dataType, exists := nodeOutputs[outputName]; exists {
			return dataType
		}
	}
	return models.DataTypeString // Default to string type
}

// convertToParserType converts iFlytek node to parser type
func (p *IFlytekParser) convertToParserType(iflytekNode IFlytekNode) IFlytekNode {
	return IFlytekNode{
		ID:               iflytekNode.ID,
		Type:             iflytekNode.Type,
		Width:            iflytekNode.Width,
		Height:           iflytekNode.Height,
		Position:         p.convertPosition(iflytekNode.Position),
		PositionAbsolute: p.convertPosition(iflytekNode.PositionAbsolute),
		Dragging:         iflytekNode.Dragging,
		Selected:         iflytekNode.Selected,
		Data:             iflytekNode.Data,
		ParentID:         iflytekNode.ParentID,
		Extent:           iflytekNode.Extent,
		ZIndex:           iflytekNode.ZIndex,
		Draggable:        iflytekNode.Draggable,
	}
}

// convertPosition converts position structure
func (p *IFlytekParser) convertPosition(pos IFlytekPosition) IFlytekPosition {
	return IFlytekPosition{X: pos.X, Y: pos.Y}
}

// organizeNodeHierarchy organizes node hierarchy.
func (p *IFlytekParser) organizeNodeHierarchy(allNodes []*models.Node, nodeParentMap map[string]string, unifiedDSL *models.UnifiedDSL) error {
	// Step 1: Create node mapping and separate nodes by hierarchy
	nodeMap, topLevelNodes, childNodesByParent := p.categorizeNodesByHierarchy(allNodes, nodeParentMap)

	// Step 2: Organize child nodes into parent workflows
	if err := p.organizeChildNodesIntoParentWorkflows(childNodesByParent, nodeMap, unifiedDSL); err != nil {
		return err
	}

	// Step 3: Add top-level nodes to main workflow
	p.addTopLevelNodesToWorkflow(topLevelNodes, unifiedDSL)

	return nil
}

// categorizeNodesByHierarchy separates nodes into hierarchy levels and creates mapping
func (p *IFlytekParser) categorizeNodesByHierarchy(allNodes []*models.Node, nodeParentMap map[string]string) (map[string]*models.Node, []*models.Node, map[string][]*models.Node) {
	// Create node mapping table
	nodeMap := make(map[string]*models.Node)
	for _, node := range allNodes {
		nodeMap[node.ID] = node
	}

	// Separate top-level nodes and child nodes
	topLevelNodes := make([]*models.Node, 0)
	childNodesByParent := make(map[string][]*models.Node)

	for _, node := range allNodes {
		if parentID, hasParent := nodeParentMap[node.ID]; hasParent {
			// This is a child node
			childNodesByParent[parentID] = append(childNodesByParent[parentID], node)
		} else {
			// This is a top-level node
			topLevelNodes = append(topLevelNodes, node)
		}
	}

	return nodeMap, topLevelNodes, childNodesByParent
}

// organizeChildNodesIntoParentWorkflows organizes child nodes into parent workflows
func (p *IFlytekParser) organizeChildNodesIntoParentWorkflows(childNodesByParent map[string][]*models.Node, nodeMap map[string]*models.Node, unifiedDSL *models.UnifiedDSL) error {
	for parentID, childNodes := range childNodesByParent {
		if err := p.organizeChildNodesForParent(parentID, childNodes, nodeMap, unifiedDSL); err != nil {
			return err
		}
	}
	return nil
}

// organizeChildNodesForParent organizes child nodes for a specific parent
func (p *IFlytekParser) organizeChildNodesForParent(parentID string, childNodes []*models.Node, nodeMap map[string]*models.Node, unifiedDSL *models.UnifiedDSL) error {
	parentNode := nodeMap[parentID]
	if parentNode == nil {
		return fmt.Errorf("parent node %s not found", parentID)
	}

	// Only iteration nodes have sub_workflow
	if parentNode.Type == models.NodeTypeIteration {
		return p.setupIterationSubWorkflow(parentNode, parentID, childNodes, unifiedDSL)
	}

	return nil
}

// setupIterationSubWorkflow sets up sub-workflow for iteration nodes
func (p *IFlytekParser) setupIterationSubWorkflow(parentNode *models.Node, parentID string, childNodes []*models.Node, unifiedDSL *models.UnifiedDSL) error {
	iterConfigPtr, ok := parentNode.Config.(*models.IterationConfig)
	if !ok {
		configType := fmt.Sprintf("%T", parentNode.Config)
		return fmt.Errorf("iteration node %s has invalid config type: %s", parentID, configType)
	}

	// Create a configuration copy
	newIterConfig := *iterConfigPtr

	// Add child nodes to sub_workflow
	p.addChildNodesToSubWorkflow(&newIterConfig, childNodes)

	// Parse iteration internal edges
	internalEdges, err := p.extractIterationInternalEdges(parentID, childNodes, unifiedDSL.Workflow.Edges)
	if err != nil {
		return fmt.Errorf("failed to extract internal edges for iteration %s: %w", parentID, err)
	}
	newIterConfig.SubWorkflow.Edges = internalEdges

	// Update configuration
	parentNode.Config = &newIterConfig
	return nil
}

// addChildNodesToSubWorkflow adds child nodes to iteration sub-workflow
func (p *IFlytekParser) addChildNodesToSubWorkflow(iterConfig *models.IterationConfig, childNodes []*models.Node) {
	iterConfig.SubWorkflow.Nodes = make([]models.Node, len(childNodes))
	for i, childNode := range childNodes {
		iterConfig.SubWorkflow.Nodes[i] = *childNode
	}
}

// addTopLevelNodesToWorkflow adds top-level nodes to main workflow
func (p *IFlytekParser) addTopLevelNodesToWorkflow(topLevelNodes []*models.Node, unifiedDSL *models.UnifiedDSL) {
	for _, node := range topLevelNodes {
		unifiedDSL.Workflow.Nodes = append(unifiedDSL.Workflow.Nodes, *node)
	}
}

// extractIterationInternalEdges extracts iteration internal connection relationships.
func (p *IFlytekParser) extractIterationInternalEdges(iterationID string, childNodes []*models.Node, allEdges []models.Edge) ([]models.Edge, error) {
	// Create child node ID set
	childNodeIDs := make(map[string]bool)
	for _, childNode := range childNodes {
		childNodeIDs[childNode.ID] = true
	}

	// Extract iteration internal connections (source and target nodes are both child nodes)
	var internalEdges []models.Edge
	for _, edge := range allEdges {
		// Check if it's an iteration internal connection
		sourceIsChild := childNodeIDs[edge.Source]
		targetIsChild := childNodeIDs[edge.Target]

		if sourceIsChild && targetIsChild {
			// This is an iteration internal connection
			internalEdges = append(internalEdges, edge)
		}
	}

	return internalEdges, nil
}

// collectAllValidNodeIDs collects all valid node IDs from unified DSL, including iteration internal nodes
func (p *IFlytekParser) collectAllValidNodeIDs(unifiedDSL *models.UnifiedDSL) map[string]bool {
	existingNodeIDs := make(map[string]bool)

	// Add main workflow nodes
	for _, node := range unifiedDSL.Workflow.Nodes {
		existingNodeIDs[node.ID] = true

		// If this is an iteration node, also add its internal nodes
		if node.Type == models.NodeTypeIteration {
			if iterConfig, ok := node.Config.(*models.IterationConfig); ok {
				for _, internalNode := range iterConfig.SubWorkflow.Nodes {
					existingNodeIDs[internalNode.ID] = true
				}
			}
		}
	}

	return existingNodeIDs
}

// parseEdges parses connections.
func (p *IFlytekParser) parseEdges(edges []IFlytekEdge, unifiedDSL *models.UnifiedDSL) error {
	// Create a set of existing node IDs for edge validation, including iteration internal nodes
	existingNodeIDs := p.collectAllValidNodeIDs(unifiedDSL)

	// Convert to interface{} list for edge parser
	edgeInterfaces := make([]interface{}, 0, len(edges))
	for _, edge := range edges {
		// Skip edges that reference non-existent (unsupported) nodes
		if !existingNodeIDs[edge.Source] || !existingNodeIDs[edge.Target] {
			fmt.Printf("⚠️  Skipping edge with unsupported nodes: %s -> %s\n", edge.Source, edge.Target)
			continue
		}

		edgeInterfaces = append(edgeInterfaces, map[string]interface{}{
			"id":           edge.ID,
			"source":       edge.Source,
			"target":       edge.Target,
			"sourceHandle": edge.SourceHandle,
			"targetHandle": edge.TargetHandle,
			"type":         edge.Type,
			"markerEnd":    edge.MarkerEnd,
			"data":         edge.Data,
		})
	}

	parsedEdges, err := p.edgeParser.ParseEdges(edgeInterfaces)
	if err != nil {
		return fmt.Errorf("failed to parse edges: %w", err)
	}

	unifiedDSL.Workflow.Edges = append(unifiedDSL.Workflow.Edges, parsedEdges...)

	return nil
}

// validateStructure validates structure.
func (p *IFlytekParser) validateStructure(root IFlytekRootStructure) error {
	// Validate metadata
	if root.FlowMeta.Name == "" {
		return fmt.Errorf("flow name is required")
	}

	// Validate at least one node exists
	if len(root.FlowData.Nodes) == 0 {
		return fmt.Errorf("at least one node is required")
	}

	// Validate nodes
	for _, node := range root.FlowData.Nodes {
		if node.ID == "" {
			return fmt.Errorf("node ID is required")
		}
		if node.Type == "" {
			return fmt.Errorf("node type is required for node %s", node.ID)
		}
	}

	return nil
}

// organizeIterationInternalEdges organizes iteration internal connection relationships after parsing all connections.
func (p *IFlytekParser) organizeIterationInternalEdges(unifiedDSL *models.UnifiedDSL) error {
	// Collect all iteration internal node IDs
	allIterationInternalNodeIDs := make(map[string]bool)

	// Traverse all nodes to find iteration nodes
	for i := range unifiedDSL.Workflow.Nodes {
		node := &unifiedDSL.Workflow.Nodes[i]
		if node.Type == models.NodeTypeIteration {
			if iterConfigPtr, ok := node.Config.(*models.IterationConfig); ok {
				// Create a configuration copy
				newIterConfig := *iterConfigPtr

				// Collect all child node IDs within this iteration node
				for _, childNode := range newIterConfig.SubWorkflow.Nodes {
					allIterationInternalNodeIDs[childNode.ID] = true
				}

				// Extract iteration internal connection relationships
				if len(newIterConfig.SubWorkflow.Nodes) > 0 {
					internalEdges, err := p.extractIterationInternalEdges(node.ID,
						convertNodesToPointers(newIterConfig.SubWorkflow.Nodes),
						unifiedDSL.Workflow.Edges)
					if err != nil {
						return fmt.Errorf("failed to extract internal edges for iteration %s: %w", node.ID, err)
					}
					newIterConfig.SubWorkflow.Edges = internalEdges

					// Update configuration
					node.Config = &newIterConfig
				}
			}
		}
	}

	// Remove iteration internal edge connections from main workflow edge list
	filteredEdges := make([]models.Edge, 0, len(unifiedDSL.Workflow.Edges))
	for _, edge := range unifiedDSL.Workflow.Edges {
		// Keep edges in main workflow if source and target nodes are not iteration internal nodes
		sourceIsInternal := allIterationInternalNodeIDs[edge.Source]
		targetIsInternal := allIterationInternalNodeIDs[edge.Target]

		if !sourceIsInternal && !targetIsInternal {
			// This is a main workflow edge (both ends are not iteration internal nodes)
			filteredEdges = append(filteredEdges, edge)
		}
		// Note: We skip mixed edges (one end is iteration internal node, one end is external node)
		// These edges should be modeled through iteration node inputs/outputs in actual iFlytek DSL
	}

	// Update main workflow edge list
	unifiedDSL.Workflow.Edges = filteredEdges

	return nil
}

// convertNodesToPointers converts node slice to pointer slice (helper function).
func convertNodesToPointers(nodes []models.Node) []*models.Node {
	pointers := make([]*models.Node, len(nodes))
	for i := range nodes {
		pointers[i] = &nodes[i]
	}
	return pointers
}

// printConversionSummary prints detailed conversion statistics
func (p *IFlytekParser) printConversionSummary(unifiedDSL *models.UnifiedDSL) {
	totalNodes := len(unifiedDSL.Workflow.Nodes)
	fmt.Printf("✅ Conversion Summary: All %d nodes processed successfully\n", totalNodes)

	// Count nodes converted to code placeholders by checking titles
	convertedCount := 0
	for _, node := range unifiedDSL.Workflow.Nodes {
		if strings.Contains(node.Title, "暂不兼容的节点-") {
			convertedCount++
		}
	}

	if convertedCount > 0 {
		fmt.Printf("ℹ️  %d unsupported nodes were converted to code node placeholders\n", convertedCount)
		fmt.Printf("ℹ️  Please manually adjust these placeholder nodes as needed\n")
	}
}

// IFlytekRootStructure represents iFlytek SparkAgent root structure.
type IFlytekRootStructure struct {
	FlowMeta IFlytekFlowMeta `yaml:"flowMeta"`
	FlowData IFlytekFlowData `yaml:"flowData"`
}

// IFlytekFlowMeta represents iFlytek SparkAgent flowMeta structure.
type IFlytekFlowMeta struct {
	Name           string `yaml:"name"`
	Description    string `yaml:"description"`
	AvatarIcon     string `yaml:"avatarIcon"`
	AvatarColor    string `yaml:"avatarColor"`
	AdvancedConfig string `yaml:"advancedConfig"`
	DSLVersion     string `yaml:"dslVersion"`
}

// IFlytekFlowData represents iFlytek SparkAgent flowData structure.
type IFlytekFlowData struct {
	Nodes []IFlytekNode `yaml:"nodes"`
	Edges []IFlytekEdge `yaml:"edges"`
}
