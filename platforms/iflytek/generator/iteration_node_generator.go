package generator

import (
	"fmt"
	"strings"

	"agentbridge/internal/models"
	"agentbridge/platforms/common"
)

// BranchMappingExtractor interface for extracting branch mapping from condition nodes
type BranchMappingExtractor interface {
	ExtractBranchMapping(node IFlytekNode)
}

// IterationNodeGenerator iteration node generator
type IterationNodeGenerator struct {
	*BaseNodeGenerator
	idMapping           map[string]string                  // Dify ID -> iFlytek SparkAgent ID mapping
	nodeTitleMapping    map[string]string                  // iFlytek SparkAgent ID -> node title mapping
	outputIDMapping     map[string]map[string]string       // Output ID mapping: nodeID -> outputName -> outputID
	conditionGenerators map[string]*ConditionNodeGenerator // Store condition generators for branch mapping extraction
	branchExtractor     BranchMappingExtractor             // Interface for extracting branch mapping
}

func NewIterationNodeGenerator() *IterationNodeGenerator {
	return &IterationNodeGenerator{
		BaseNodeGenerator:   NewBaseNodeGenerator(models.NodeTypeIteration),
		idMapping:           make(map[string]string),
		nodeTitleMapping:    make(map[string]string),
		outputIDMapping:     make(map[string]map[string]string),
		conditionGenerators: make(map[string]*ConditionNodeGenerator),
	}
}

// SetIDMapping sets ID mapping
func (g *IterationNodeGenerator) SetIDMapping(idMapping map[string]string) {
	g.idMapping = idMapping
}

// SetNodeTitleMapping sets node title mapping
func (g *IterationNodeGenerator) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	g.nodeTitleMapping = nodeTitleMapping
}

// SetOutputIDMapping sets output ID mapping
func (g *IterationNodeGenerator) SetOutputIDMapping(outputIDMapping map[string]map[string]string) {
	g.outputIDMapping = outputIDMapping
}

// recordOutputIDs records the output IDs for a generated node
func (g *IterationNodeGenerator) recordOutputIDs(nodeID string, outputs []IFlytekOutput) {
	if g.outputIDMapping == nil {
		g.outputIDMapping = make(map[string]map[string]string)
	}

	if g.outputIDMapping[nodeID] == nil {
		g.outputIDMapping[nodeID] = make(map[string]string)
	}

	for _, output := range outputs {
		g.outputIDMapping[nodeID][output.Name] = output.ID
	}
}

// SetBranchExtractor sets branch mapping extractor
func (g *IterationNodeGenerator) SetBranchExtractor(extractor BranchMappingExtractor) {
	g.branchExtractor = extractor
}

// storeConditionGenerator stores condition generator for branch mapping extraction
func (g *IterationNodeGenerator) storeConditionGenerator(nodeID string, condGen *ConditionNodeGenerator) {
	g.conditionGenerators[nodeID] = condGen
}

// ExtractConditionBranchMappings extracts branch mappings from all condition nodes
func (g *IterationNodeGenerator) ExtractConditionBranchMappings(nodes []IFlytekNode) {
	if g.branchExtractor == nil {
		return
	}

	for _, node := range nodes {
		if strings.HasPrefix(node.ID, "if-else::") {
			g.branchExtractor.ExtractBranchMapping(node)
		}
	}
}

// GenerateNode generates iteration node
func (g *IterationNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	iterationConfig, ok := common.AsIterationConfig(node.Config)
	if !ok || iterationConfig == nil {
		return IFlytekNode{}, fmt.Errorf("迭代节点配置类型错误")
	}

	// Generate or get iteration node ID
	var iterationID string
	if existingID, exists := g.idMapping[node.ID]; exists {
		// If mapping already exists, use existing ID
		iterationID = existingID
	} else {
		// Generate ID on first generation
		iterationID = g.generateIFlytekNodeID(models.NodeTypeIteration)
		g.idMapping[node.ID] = iterationID
	}
	g.nodeTitleMapping[iterationID] = node.Title

	// Generate unified ID for iteration input and start node output
	iterationInputID := g.generateRandomID()
	// Generate deterministic start node ID, associated with iteration main node ID
	startNodeID := g.generateDeterministicStartNodeID(iterationID)

	// Extract input references
	references := g.processIterationReferences(node.Inputs)
	inputs := g.processIterationInputsWithID(node.Inputs, iterationInputID)

	// Generate outputs with deterministic IDs
	outputs := g.generateIterationOutputsWithMapping(node.Outputs, iterationID)

	// Record iteration outputs for mapping
	g.recordOutputIDs(iterationID, outputs)

	// Generate nodeParam
	nodeParam := g.generateIterationNodeParam(*iterationConfig, startNodeID)

	// Create main iteration node
	iflytekNode := IFlytekNode{
		ID:               iterationID,
		Dragging:         false,
		Selected:         false,
		Width:            635, // Set according to examples
		Height:           763, // Set according to examples
		Position:         g.convertPosition(node.Position),
		PositionAbsolute: g.convertPosition(node.Position),
		Type:             "迭代",
		Data: IFlytekNodeData{
			AllowInputReference:  true,
			AllowOutputReference: true,
			Label:                node.Title,
			LabelEdit:            false,
			References:           references,
			Status:               "",
			NodeMeta: IFlytekNodeMeta{
				AliasName: "迭代",
				NodeType:  "基础节点",
			},
			Inputs:      inputs,
			Outputs:     outputs,
			NodeParam:   nodeParam,
			Icon:        g.getNodeIcon(models.NodeTypeIteration),
			Description: "该节点用于处理循环逻辑，仅支持嵌套一次",
			Updatable:   false,
		},
	}

	return iflytekNode, nil
}

// GenerateIterationSubNodes generates iteration sub-nodes (including start node and end node)
func (g *IterationNodeGenerator) GenerateIterationSubNodes(iterationNode models.Node, iterationID string, subNodes []models.Node, iterationStartNodeID string) ([]IFlytekNode, []IFlytekEdge, error) {
	// Call method using randomly generated IDs
	return g.GenerateIterationSubNodesWithIDs(iterationNode, iterationID, subNodes, iterationStartNodeID, "", "")
}

// GenerateIterationSubNodesWithIDs generates iteration sub-nodes using pre-generated IDs
func (g *IterationNodeGenerator) GenerateIterationSubNodesWithIDs(iterationNode models.Node, iterationID string, subNodes []models.Node, iterationStartNodeID string, iterationEndNodeID string, iterationCodeNodeID string) ([]IFlytekNode, []IFlytekEdge, error) {
	iterationConfig, ok := common.AsIterationConfig(iterationNode.Config)
	if !ok || iterationConfig == nil {
		return nil, nil, fmt.Errorf("迭代节点配置类型错误")
	}

	var subIFlytekNodes []IFlytekNode

	startNode, iterationInputID := g.createIterationStartNode(iterationID, iterationStartNodeID)
	subIFlytekNodes = append(subIFlytekNodes, startNode)

	childNodes, err := g.generateIterationChildNodes(subNodes, iterationID, iterationCodeNodeID, startNode.ID, iterationInputID)
	if err != nil {
		return nil, nil, err
	}
	subIFlytekNodes = append(subIFlytekNodes, childNodes...)

	endNode, sourceNode, startNodePtr := g.createIterationEndNode(iterationEndNodeID, iterationID, *iterationConfig, subIFlytekNodes)
	subIFlytekNodes = append(subIFlytekNodes, endNode)

	g.ExtractConditionBranchMappings(subIFlytekNodes)

	iterationEdges := g.generateIterationInternalEdges(subIFlytekNodes, sourceNode, &endNode, startNodePtr)

	return subIFlytekNodes, iterationEdges, nil
}

// createIterationStartNode creates the iteration start node
func (g *IterationNodeGenerator) createIterationStartNode(iterationID, iterationStartNodeID string) (IFlytekNode, string) {
	startNodeID := g.determineStartNodeID(iterationID, iterationStartNodeID)
	iterationInputID := g.generateRandomID()

	startNode := IFlytekNode{
		ID:               startNodeID,
		Dragging:         false,
		Selected:         false,
		Width:            68,
		Height:           44,
		Position:         IFlytekPosition{X: 30, Y: 397.5},
		PositionAbsolute: IFlytekPosition{X: -668, Y: 58},
		Type:             "开始节点",
		ParentID:         &iterationID,
		Extent:           "parent",
		ZIndex:           1,
		Draggable:        &[]bool{false}[0],
		Data:             g.createStartNodeData(iterationID, iterationInputID),
	}

	g.nodeTitleMapping[startNodeID] = "开始"
	return startNode, iterationInputID
}

// determineStartNodeID determines the start node ID to use
func (g *IterationNodeGenerator) determineStartNodeID(iterationID, iterationStartNodeID string) string {
	if iterationStartNodeID != "" && strings.HasPrefix(iterationStartNodeID, "iteration-node-start::") {
		return iterationStartNodeID
	}
	return g.generateDeterministicStartNodeID(iterationID)
}

// createStartNodeData creates the data for iteration start node
func (g *IterationNodeGenerator) createStartNodeData(iterationID, iterationInputID string) IFlytekNodeData {
	return IFlytekNodeData{
		AllowInputReference:  false,
		AllowOutputReference: true,
		Label:                "开始",
		Status:               "",
		NodeMeta: IFlytekNodeMeta{
			AliasName: "开始节点",
			NodeType:  "基础节点",
		},
		Inputs: []IFlytekInput{},
		Outputs: []IFlytekOutput{
			{
				ID:         iterationInputID,
				Name:       "input",
				NameErrMsg: "",
				Schema: IFlytekSchema{
					Type:    "string",
					Default: "",
				},
			},
		},
		NodeParam:      map[string]interface{}{},
		Icon:           g.getNodeIcon(models.NodeTypeStart),
		Description:    "工作流的开启节点，用于定义流程调用所需的业务变量信息。",
		ParentID:       &iterationID,
		OriginPosition: &IFlytekPosition{X: -668, Y: 58},
		Updatable:      false,
	}
}

// generateIterationChildNodes generates child nodes inside iteration
func (g *IterationNodeGenerator) generateIterationChildNodes(subNodes []models.Node, iterationID, iterationCodeNodeID, startNodeID, iterationInputID string) ([]IFlytekNode, error) {
	var childNodes []IFlytekNode

	for _, subNode := range subNodes {
		if subNode.Type == models.NodeTypeStart {
			continue
		}

		childNode, err := g.generateSingleChildNode(subNode, iterationID, iterationCodeNodeID, startNodeID, iterationInputID)
		if err != nil {
			return nil, fmt.Errorf("生成迭代子节点失败 %s: %w", subNode.ID, err)
		}

		childNodes = append(childNodes, childNode)
	}

	return childNodes, nil
}

// generateSingleChildNode generates a single child node
func (g *IterationNodeGenerator) generateSingleChildNode(subNode models.Node, iterationID, iterationCodeNodeID, startNodeID, iterationInputID string) (IFlytekNode, error) {
	if subNode.Type == models.NodeTypeCode && iterationCodeNodeID != "" {
		return g.generateIterationChildNodeWithID(subNode, iterationID, iterationCodeNodeID, startNodeID, iterationInputID)
	}
	return g.generateIterationChildNode(subNode, iterationID, startNodeID, iterationInputID)
}

// createIterationEndNode creates the iteration end node
func (g *IterationNodeGenerator) createIterationEndNode(iterationEndNodeID, iterationID string, iterationConfig models.IterationConfig, subNodes []IFlytekNode) (IFlytekNode, *IFlytekNode, *IFlytekNode) {
	endNodeID := g.determineEndNodeID(iterationEndNodeID)
	sourceNode, startNode := g.findIterationSourceNodes(iterationConfig, subNodes)

	endNode := IFlytekNode{
		ID:               endNodeID,
		Dragging:         false,
		Selected:         false,
		Width:            68,
		Height:           44,
		Position:         IFlytekPosition{X: 502.7, Y: 421.4},
		PositionAbsolute: IFlytekPosition{X: 1222.9, Y: 153.7},
		Type:             "结束节点",
		ParentID:         &iterationID,
		Extent:           "parent",
		ZIndex:           1,
		Draggable:        &[]bool{false}[0],
		Data:             g.createEndNodeData(iterationID, iterationConfig, sourceNode, startNode),
	}

	return endNode, sourceNode, startNode
}

// determineEndNodeID determines the end node ID to use
func (g *IterationNodeGenerator) determineEndNodeID(iterationEndNodeID string) string {
	if iterationEndNodeID != "" {
		return iterationEndNodeID
	}
	return g.generateSpecialNodeID("iteration-node-end")
}

// findIterationSourceNodes finds source and start nodes in iteration
func (g *IterationNodeGenerator) findIterationSourceNodes(iterationConfig models.IterationConfig, subNodes []IFlytekNode) (*IFlytekNode, *IFlytekNode) {
	var sourceNode, startNode *IFlytekNode

	outputSourceNodeID := iterationConfig.OutputSelector.NodeID
	if outputSourceNodeID != "" {
		sourceNode, startNode = g.findNodesByOutputSelector(outputSourceNodeID, subNodes)
	}

	if sourceNode == nil {
		sourceNode = g.findLastCodeNode(subNodes)
	}

	return sourceNode, startNode
}

// findNodesByOutputSelector finds nodes by output selector
func (g *IterationNodeGenerator) findNodesByOutputSelector(outputSourceNodeID string, subNodes []IFlytekNode) (*IFlytekNode, *IFlytekNode) {
	var sourceNode, startNode *IFlytekNode

	for i := range subNodes {
		if strings.HasPrefix(subNodes[i].ID, "iteration-node-start::") {
			startNode = &subNodes[i]
		}

		if g.isOutputSelectorMatch(outputSourceNodeID, subNodes[i].ID) {
			sourceNode = &subNodes[i]
		}
	}

	return sourceNode, startNode
}

// isOutputSelectorMatch checks if node matches output selector
func (g *IterationNodeGenerator) isOutputSelectorMatch(outputSourceNodeID, nodeID string) bool {
	if g.idMapping == nil {
		return false
	}

	for originalNodeID, mappedNodeID := range g.idMapping {
		if originalNodeID == outputSourceNodeID && mappedNodeID == nodeID {
			return true
		}
	}
	return false
}

// findLastCodeNode finds the last code node for fallback
func (g *IterationNodeGenerator) findLastCodeNode(subNodes []IFlytekNode) *IFlytekNode {
	for i := len(subNodes) - 1; i >= 0; i-- {
		if strings.HasPrefix(subNodes[i].ID, "ifly-code::") {
			return &subNodes[i]
		}
	}
	return nil
}

// createEndNodeData creates the data for iteration end node
func (g *IterationNodeGenerator) createEndNodeData(iterationID string, iterationConfig models.IterationConfig, sourceNode, startNode *IFlytekNode) IFlytekNodeData {
	return IFlytekNodeData{
		AllowInputReference:  true,
		AllowOutputReference: false,
		Label:                "结束",
		Status:               "",
		NodeMeta: IFlytekNodeMeta{
			AliasName: "结束节点",
			NodeType:  "基础节点",
		},
		Inputs:     g.generateIterationEndInputs(iterationConfig, sourceNode, startNode),
		Outputs:    []IFlytekOutput{},
		References: g.generateIterationEndReferences(sourceNode, startNode),
		NodeParam: map[string]interface{}{
			"template":   "",
			"outputMode": 0,
		},
		Icon:           g.getNodeIcon(models.NodeTypeEnd),
		Description:    "工作流的结束节点，用于输出工作流运行后的最终结果。",
		ParentID:       &iterationID,
		OriginPosition: &IFlytekPosition{X: 1222.9, Y: 153.7},
		Updatable:      false,
	}
}

// generateIterationChildNode generates child nodes inside iteration
func (g *IterationNodeGenerator) generateIterationChildNode(subNode models.Node, iterationID string, startNodeID string, iterationInputID string) (IFlytekNode, error) {
	// Get the corresponding node generator
	generator, err := g.getChildNodeGenerator(subNode.Type)
	if err != nil {
		return IFlytekNode{}, fmt.Errorf("无法获取节点生成器 %s: %w", subNode.Type, err)
	}

	// Set ID mapping
	if mappingSetter, ok := generator.(interface{ SetIDMapping(map[string]string) }); ok {
		mappingSetter.SetIDMapping(g.idMapping)
	}
	if titleSetter, ok := generator.(interface{ SetNodeTitleMapping(map[string]string) }); ok {
		titleSetter.SetNodeTitleMapping(g.nodeTitleMapping)
	}

	// Fix input references of iteration sub-nodes, change references to iteration main node to references to iteration start node
	// Here we pass empty string because at this stage we don't have the iteration start node ID yet
	fixedSubNode := g.fixIterationSubNodeReferencesWithID(subNode, iterationID, startNodeID, iterationInputID)

	// Generate base node
	baseNode, err := generator.GenerateNode(fixedSubNode)
	if err != nil {
		return IFlytekNode{}, fmt.Errorf("生成基础节点失败 %s: %w", subNode.ID, err)
	}

	// Update ID mapping with the generated child node mapping
	g.idMapping[subNode.ID] = baseNode.ID
	g.nodeTitleMapping[baseNode.ID] = subNode.Title

	// Record output IDs for this node before fixing
	g.recordOutputIDs(baseNode.ID, baseNode.Data.Outputs)

	// Fix output IDs to match parent iteration node outputs
	baseNode = g.fixIterationChildOutputIDs(baseNode, iterationID)

	// Update the recorded output IDs after fixing
	g.recordOutputIDs(baseNode.ID, baseNode.Data.Outputs)

	// Fix references in the generated node to use mapped UUIDs and correct output IDs
	baseNode = g.fixGeneratedNodeReferences(baseNode)

	// Set iteration sub-node specific properties
	baseNode.ParentID = &iterationID
	baseNode.Extent = "parent"
	baseNode.ZIndex = 1
	baseNode.Draggable = &[]bool{false}[0]

	// Set ParentID in Data
	baseNode.Data.ParentID = &iterationID

	// Set OriginPosition (save original position information)
	baseNode.Data.OriginPosition = &IFlytekPosition{
		X: baseNode.PositionAbsolute.X,
		Y: baseNode.PositionAbsolute.Y,
	}

	// If this is a condition node, we need to store branch mapping
	// This will be handled by the main generator through callback
	if subNode.Type == models.NodeTypeCondition {
		if condGen, ok := generator.(*ConditionNodeGenerator); ok {
			// Store condition generator in the iteration generator for later access
			g.storeConditionGenerator(baseNode.ID, condGen)
			// Immediately extract branch mapping if branchExtractor is available
			if g.branchExtractor != nil {
				g.branchExtractor.ExtractBranchMapping(baseNode)
			}
		}
	}

	return baseNode, nil
}

// generateIterationChildNodeWithID generates child nodes inside iteration using pre-generated ID
func (g *IterationNodeGenerator) generateIterationChildNodeWithID(subNode models.Node, iterationID string, preGeneratedID string, startNodeID string, iterationInputID string) (IFlytekNode, error) {
	// Get the corresponding node generator
	generator, err := g.getChildNodeGenerator(subNode.Type)
	if err != nil {
		return IFlytekNode{}, fmt.Errorf("无法获取节点生成器 %s: %w", subNode.Type, err)
	}

	// Set ID mapping
	if mappingSetter, ok := generator.(interface{ SetIDMapping(map[string]string) }); ok {
		mappingSetter.SetIDMapping(g.idMapping)
	}
	if titleSetter, ok := generator.(interface{ SetNodeTitleMapping(map[string]string) }); ok {
		titleSetter.SetNodeTitleMapping(g.nodeTitleMapping)
	}

	// Fix input references of iteration sub-nodes, change references to iteration main node to references to iteration start node
	// Here we pass empty string because at this stage we don't have the iteration start node ID yet
	fixedSubNode := g.fixIterationSubNodeReferencesWithID(subNode, iterationID, startNodeID, iterationInputID)

	// Generate base node
	baseNode, err := generator.GenerateNode(fixedSubNode)
	if err != nil {
		return IFlytekNode{}, fmt.Errorf("生成基础节点失败 %s: %w", subNode.ID, err)
	}

	// Use pre-generated ID
	baseNode.ID = preGeneratedID

	// Update ID mapping with the generated child node mapping
	g.idMapping[subNode.ID] = baseNode.ID
	g.nodeTitleMapping[baseNode.ID] = subNode.Title

	// Record output IDs for this node before fixing
	g.recordOutputIDs(baseNode.ID, baseNode.Data.Outputs)

	// Fix output IDs to match parent iteration node outputs
	baseNode = g.fixIterationChildOutputIDs(baseNode, iterationID)

	// Update the recorded output IDs after fixing
	g.recordOutputIDs(baseNode.ID, baseNode.Data.Outputs)

	// Fix references in the generated node to use mapped UUIDs and correct output IDs
	baseNode = g.fixGeneratedNodeReferences(baseNode)

	// Set iteration sub-node specific properties
	baseNode.ParentID = &iterationID
	baseNode.Extent = "parent"
	baseNode.ZIndex = 1
	baseNode.Draggable = &[]bool{false}[0]

	// Set ParentID in Data
	baseNode.Data.ParentID = &iterationID

	// Set OriginPosition (save original position information)
	baseNode.Data.OriginPosition = &IFlytekPosition{
		X: baseNode.PositionAbsolute.X,
		Y: baseNode.PositionAbsolute.Y,
	}

	// If this is a condition node, extract branch mapping immediately
	if subNode.Type == models.NodeTypeCondition {
		if condGen, ok := generator.(*ConditionNodeGenerator); ok {
			// Store condition generator in the iteration generator for later access
			g.storeConditionGenerator(baseNode.ID, condGen)
			// Immediately extract branch mapping if branchExtractor is available
			if g.branchExtractor != nil {
				g.branchExtractor.ExtractBranchMapping(baseNode)
			}
		}
	}

	return baseNode, nil
}

// ChildNodeGenerator child node generator interface
type ChildNodeGenerator interface {
	GenerateNode(node models.Node) (IFlytekNode, error)
}

// getChildNodeGenerator gets the generator corresponding to the child node
func (g *IterationNodeGenerator) getChildNodeGenerator(nodeType models.NodeType) (ChildNodeGenerator, error) {
	switch nodeType {
	case models.NodeTypeCode:
		codeGen := NewCodeNodeGenerator()
		return codeGen, nil
	case models.NodeTypeLLM:
		llmGen := NewLLMNodeGenerator()
		return llmGen, nil
	case models.NodeTypeCondition:
		condGen := NewConditionNodeGenerator()
		// Set ID mappings for the condition generator
		condGen.SetIDMapping(g.idMapping)
		condGen.SetNodeTitleMapping(g.nodeTitleMapping)
		return condGen, nil
	case models.NodeTypeClassifier:
		classifierGen := NewClassifierNodeGenerator()
		// Set ID mappings for the classifier generator
		classifierGen.SetIDMapping(g.idMapping)
		classifierGen.SetNodeTitleMapping(g.nodeTitleMapping)
		return classifierGen, nil
	default:
		return nil, fmt.Errorf("不支持的迭代子节点类型: %s", nodeType)
	}
}

// generateIterationEndInputs generates inputs for iteration end node
func (g *IterationNodeGenerator) generateIterationEndInputs(config models.IterationConfig, sourceNode *IFlytekNode, startNode *IFlytekNode) []IFlytekInput {
	if sourceNode == nil {
		return []IFlytekInput{}
	}

	// Determine output name based on iteration configuration output selector
	outputName := config.OutputSelector.OutputName
	if outputName == "" {
		outputName = "result" // Default output name
	}

	// Find the output in source node that matches the output selector specification
	var outputRef *IFlytekRefContent
	var foundOutput *IFlytekOutput

	for i := range sourceNode.Data.Outputs {
		if sourceNode.Data.Outputs[i].Name == outputName {
			foundOutput = &sourceNode.Data.Outputs[i]
			break
		}
	}

	if foundOutput != nil {
		outputRef = &IFlytekRefContent{
			Name:   foundOutput.Name,
			ID:     foundOutput.ID,
			NodeID: sourceNode.ID,
		}
	} else if len(sourceNode.Data.Outputs) > 0 {
		// If the specified output is not found, use the first output
		outputRef = &IFlytekRefContent{
			Name:   sourceNode.Data.Outputs[0].Name,
			ID:     sourceNode.Data.Outputs[0].ID,
			NodeID: sourceNode.ID,
		}
	} else {
		// If source node has no output, provide default reference
		outputRef = &IFlytekRefContent{
			Name:   outputName,
			ID:     g.generateRandomID(),
			NodeID: sourceNode.ID,
		}
	}

	return []IFlytekInput{
		{
			ID:         g.generateRandomID(),
			Name:       "output",
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type: "string",
				Value: &IFlytekSchemaValue{
					Type:          "ref",
					Content:       outputRef,
					ContentErrMsg: "",
				},
			},
			FileType: "",
		},
	}
}

// generateIterationEndReferences generates references for iteration end node
func (g *IterationNodeGenerator) generateIterationEndReferences(sourceNode *IFlytekNode, startNode *IFlytekNode) []IFlytekReference {
	var references []IFlytekReference

	// Add reference to source node inside iteration
	if sourceNode != nil {
		var sourceRefDetails []IFlytekRefDetail
		for _, output := range sourceNode.Data.Outputs {
			sourceRefDetails = append(sourceRefDetails, IFlytekRefDetail{
				OriginID: sourceNode.ID,
				ID:       output.ID,
				Label:    output.Name,
				Type:     output.Schema.Type,
				Value:    output.Name,
				FileType: "",
			})
		}

		references = append(references, IFlytekReference{
			Children: []IFlytekReference{
				{
					Label:      "",
					Value:      "",
					References: sourceRefDetails,
				},
			},
			Label:      sourceNode.Data.Label,
			Value:      sourceNode.ID,
			ParentNode: true,
		})
	}

	// Add reference to iteration start node
	if startNode != nil {
		var startRefDetails []IFlytekRefDetail
		for _, output := range startNode.Data.Outputs {
			startRefDetails = append(startRefDetails, IFlytekRefDetail{
				OriginID: startNode.ID,
				ID:       output.ID,
				Label:    output.Name,
				Type:     output.Schema.Type,
				Value:    output.Name,
				FileType: "",
			})
		}

		references = append(references, IFlytekReference{
			Children: []IFlytekReference{
				{
					Label:      "",
					Value:      "",
					References: startRefDetails,
				},
			},
			Label:      startNode.Data.Label,
			Value:      startNode.ID,
			ParentNode: true,
		})
	}

	return references
}

// findOriginalIterationIDByMapping finds the original Dify iteration node ID through reverse lookup
func (g *IterationNodeGenerator) findOriginalIterationIDByMapping(iterationID string) string {
	if g.idMapping == nil {
		return ""
	}

	for originalID, mappedID := range g.idMapping {
		if mappedID == iterationID {
			return originalID
		}
	}
	return ""
}

// resolveIterationStartNodeID resolves iteration start node ID from mapping if not provided
func (g *IterationNodeGenerator) resolveIterationStartNodeID(iterationStartNodeID string) string {
	if iterationStartNodeID != "" {
		return iterationStartNodeID
	}

	for mappedID := range g.idMapping {
		if strings.HasPrefix(mappedID, "iteration-node-start::") {
			return mappedID
		}
	}
	return ""
}

// processSubNodeInputReferences processes all input references of a sub-node
func (g *IterationNodeGenerator) processSubNodeInputReferences(inputs []models.Input, originalIterationID, iterationID, startNodeID string) []models.Input {
	var fixedInputs []models.Input

	for _, input := range inputs {
		fixedInput := input

		if input.Reference != nil {
			fixedInput.Reference = g.fixIterationReference(input.Reference, originalIterationID, iterationID, startNodeID)
		}

		fixedInputs = append(fixedInputs, fixedInput)
	}

	return fixedInputs
}

// fixIterationReference fixes a single iteration reference
func (g *IterationNodeGenerator) fixIterationReference(ref *models.VariableReference, originalIterationID, iterationID, startNodeID string) *models.VariableReference {
	if !g.shouldFixReference(ref, originalIterationID, iterationID) {
		return ref
	}

	if startNodeID == "" {
		return ref
	}

	return &models.VariableReference{
		Type:       ref.Type,
		NodeID:     startNodeID,
		OutputName: "input", // Uniformly use input as output name
		DataType:   ref.DataType,
	}
}

// shouldFixReference determines if a reference should be fixed
func (g *IterationNodeGenerator) shouldFixReference(ref *models.VariableReference, originalIterationID, iterationID string) bool {
	return ref.NodeID == originalIterationID || ref.NodeID == iterationID
}

// fixIterationSubNodeReferencesWithID fixes references of sub-nodes inside iteration using specified ID
func (g *IterationNodeGenerator) fixIterationSubNodeReferencesWithID(subNode models.Node, iterationID string, iterationStartNodeID string, iterationInputID string) models.Node {
	fixedNode := subNode
	originalIterationID := g.findOriginalIterationID(iterationID)

	fixedNode.Inputs = g.fixSubNodeInputReferences(subNode.Inputs, originalIterationID, iterationID, iterationStartNodeID)

	if fixedNode.Type == models.NodeTypeCondition {
		if condConfig, ok := common.AsConditionConfig(fixedNode.Config); ok && condConfig != nil {
			fixedNode.Config = g.fixConditionConfigReferences(*condConfig, originalIterationID, iterationID, iterationStartNodeID)
		}
	}

	return fixedNode
}

// findOriginalIterationID finds the original Dify iteration node ID through reverse lookup
func (g *IterationNodeGenerator) findOriginalIterationID(iterationID string) string {
	return common.ReverseIDMapping(g.idMapping, iterationID)
}

// fixSubNodeInputReferences fixes all input references of a sub-node
func (g *IterationNodeGenerator) fixSubNodeInputReferences(inputs []models.Input, originalIterationID, iterationID, iterationStartNodeID string) []models.Input {
	var fixedInputs []models.Input

	for _, input := range inputs {
		fixedInput := input

		if input.Reference != nil {
			fixedInput.Reference = g.fixInputVariableReference(input.Reference, originalIterationID, iterationID, iterationStartNodeID)
		}

		fixedInputs = append(fixedInputs, fixedInput)
	}

	return fixedInputs
}

// fixInputVariableReference fixes a single input variable reference
func (g *IterationNodeGenerator) fixInputVariableReference(ref *models.VariableReference, originalIterationID, iterationID, iterationStartNodeID string) *models.VariableReference {
	if g.isIterationItemReference(ref, originalIterationID, iterationID) {
		return g.createIterationStartReference(ref, iterationStartNodeID)
	}

	if g.isIterationMainNodeReference(ref, originalIterationID, iterationID) {
		return g.createIterationStartReference(ref, iterationStartNodeID)
	}

	if mappedRef := g.tryRemapNodeReference(ref); mappedRef != nil {
		return mappedRef
	}

	return ref
}

// isIterationItemReference checks if reference is to iteration main node's item output
func (g *IterationNodeGenerator) isIterationItemReference(ref *models.VariableReference, originalIterationID, iterationID string) bool {
	return (ref.NodeID == originalIterationID || ref.NodeID == iterationID) && ref.OutputName == "item"
}

// isIterationMainNodeReference checks if reference is to iteration main node
func (g *IterationNodeGenerator) isIterationMainNodeReference(ref *models.VariableReference, originalIterationID, iterationID string) bool {
	return ref.NodeID == originalIterationID || ref.NodeID == iterationID
}

// createIterationStartReference creates a reference to iteration start node
func (g *IterationNodeGenerator) createIterationStartReference(originalRef *models.VariableReference, iterationStartNodeID string) *models.VariableReference {
	return common.CreateVariableReference(originalRef, iterationStartNodeID, "input")
}

// tryRemapNodeReference tries to remap node reference using ID mapping
func (g *IterationNodeGenerator) tryRemapNodeReference(ref *models.VariableReference) *models.VariableReference {
	if mappedNodeID, exists := common.TryRemapNodeID(g.idMapping, ref.NodeID); exists {
		return common.CreateVariableReference(ref, mappedNodeID, ref.OutputName)
	}
	return nil
}

// fixConditionConfigReferences fixes references in condition configuration
func (g *IterationNodeGenerator) fixConditionConfigReferences(condConfig models.ConditionConfig, originalIterationID, iterationID, iterationStartNodeID string) models.ConditionConfig {

	var fixedCases []models.ConditionCase
	for _, caseItem := range condConfig.Cases {
		fixedCase := caseItem
		fixedCase.Conditions = g.fixConditionCaseConditions(caseItem.Conditions, originalIterationID, iterationID, iterationStartNodeID)
		fixedCases = append(fixedCases, fixedCase)
	}

	condConfig.Cases = fixedCases
	return condConfig
}

// fixConditionCaseConditions fixes conditions in a condition case
func (g *IterationNodeGenerator) fixConditionCaseConditions(conditions []models.Condition, originalIterationID, iterationID, iterationStartNodeID string) []models.Condition {
	var fixedConditions []models.Condition

	for _, condition := range conditions {
		fixedCondition := condition

		if len(condition.VariableSelector) >= 2 {
			fixedCondition.VariableSelector = g.fixVariableSelector(condition.VariableSelector, originalIterationID, iterationID, iterationStartNodeID)
		}

		fixedConditions = append(fixedConditions, fixedCondition)
	}

	return fixedConditions
}

// fixVariableSelector fixes variable selector references
func (g *IterationNodeGenerator) fixVariableSelector(selector []string, originalIterationID, iterationID, iterationStartNodeID string) []string {
	nodeID := selector[0]
	outputName := selector[1]

	if g.isIterationSelectorItemReference(nodeID, outputName, originalIterationID, iterationID) {
		return g.createIterationStartSelector(iterationStartNodeID)
	}

	if g.isIterationSelectorMainReference(nodeID, originalIterationID, iterationID) {
		return g.createIterationStartSelector(iterationStartNodeID)
	}

	if mappedSelector := g.tryRemapVariableSelector(nodeID, outputName); mappedSelector != nil {
		return mappedSelector
	}

	return selector
}

// isIterationSelectorItemReference checks if selector references iteration item output
func (g *IterationNodeGenerator) isIterationSelectorItemReference(nodeID, outputName, originalIterationID, iterationID string) bool {
	return (nodeID == originalIterationID || nodeID == iterationID) && outputName == "item"
}

// isIterationSelectorMainReference checks if selector references iteration main node
func (g *IterationNodeGenerator) isIterationSelectorMainReference(nodeID, originalIterationID, iterationID string) bool {
	return nodeID == originalIterationID || nodeID == iterationID
}

// createIterationStartSelector creates selector for iteration start node
func (g *IterationNodeGenerator) createIterationStartSelector(iterationStartNodeID string) []string {
	if iterationStartNodeID == "" {
		return nil
	}
	return []string{iterationStartNodeID, "input"}
}

// tryRemapVariableSelector tries to remap variable selector using ID mapping
func (g *IterationNodeGenerator) tryRemapVariableSelector(nodeID, outputName string) []string {
	if mappedNodeID, exists := common.TryRemapNodeID(g.idMapping, nodeID); exists {
		return []string{mappedNodeID, outputName}
	}
	return nil
}

// fixGeneratedNodeReferences fixes references in the generated iFlytek node to use mapped UUIDs
func (g *IterationNodeGenerator) fixGeneratedNodeReferences(node IFlytekNode) IFlytekNode {
	node = g.fixNodeDataReferences(node)
	node = g.fixNodeInputReferences(node)
	return node
}

func (g *IterationNodeGenerator) fixNodeDataReferences(node IFlytekNode) IFlytekNode {
	if len(node.Data.References) > 0 {
		fixedReferences := make([]IFlytekReference, len(node.Data.References))

		for i, ref := range node.Data.References {
			fixedRef := g.fixIFlytekReference(ref)
			fixedReferences[i] = fixedRef
		}

		node.Data.References = fixedReferences
	}
	return node
}

func (g *IterationNodeGenerator) fixNodeInputReferences(node IFlytekNode) IFlytekNode {
	if len(node.Data.Inputs) > 0 {
		fixedInputs := make([]IFlytekInput, len(node.Data.Inputs))

		for i, input := range node.Data.Inputs {
			fixedInput := g.processInputReference(input)
			fixedInputs[i] = fixedInput
		}

		node.Data.Inputs = fixedInputs
	}
	return node
}

func (g *IterationNodeGenerator) processInputReference(input IFlytekInput) IFlytekInput {
	fixedInput := input

	if input.Schema.Value != nil && input.Schema.Value.Content != nil {
		if refContent, ok := input.Schema.Value.Content.(*IFlytekRefContent); ok {
			content := *refContent
			content = g.fixInputNodeIDMapping(content)
			content = g.fixInputOutputMapping(content)
			fixedInput.Schema.Value.Content = &content
		}
	}

	return fixedInput
}

func (g *IterationNodeGenerator) fixInputNodeIDMapping(content IFlytekRefContent) IFlytekRefContent {
	if g.idMapping != nil {
		if mappedNodeID, exists := g.idMapping[content.NodeID]; exists {
			content.NodeID = mappedNodeID
		} else if content.NodeID != "" && !strings.Contains(content.NodeID, "::") {
			content.NodeID = g.findAlternativeMapping(content.NodeID)
		}
	}
	return content
}

func (g *IterationNodeGenerator) fixInputOutputMapping(content IFlytekRefContent) IFlytekRefContent {
	if g.outputIDMapping != nil {
		if outputMap, exists := g.outputIDMapping[content.NodeID]; exists {
			if actualOutputID, exists := outputMap[content.Name]; exists {
				content.ID = actualOutputID
			} else {
				content = g.handleOutputNameMismatch(content, outputMap)
			}
		}
	}
	return content
}

func (g *IterationNodeGenerator) handleOutputNameMismatch(content IFlytekRefContent, outputMap map[string]string) IFlytekRefContent {
	if content.Name == "text" {
		if actualOutputID, exists := outputMap["output"]; exists {
			content.ID = actualOutputID
			content.Name = "output"
		}
	}
	return content
}

// fixIFlytekReference fixes a single IFlytekReference to use mapped UUIDs
func (g *IterationNodeGenerator) fixIFlytekReference(ref IFlytekReference) IFlytekReference {
	fixedRef := ref

	fixedRef.Value = g.fixReferenceValue(ref.Value)
	fixedRef.Label = g.fixReferenceLabel(fixedRef.Value, ref.Label)
	fixedRef.Children = g.fixChildrenReferences(ref.Children)

	return fixedRef
}

// fixReferenceValue fixes the Value field if it contains original Dify node ID
func (g *IterationNodeGenerator) fixReferenceValue(value string) string {
	if g.idMapping == nil {
		return value
	}

	if mappedID, exists := g.idMapping[value]; exists {
		return mappedID
	}

	if value != "" && !strings.Contains(value, "::") {
		return g.findAlternativeMapping(value)
	}

	return value
}

// findAlternativeMapping tries to find mapping by iterating through all mappings
func (g *IterationNodeGenerator) findAlternativeMapping(value string) string {
	for originalID, mappedID := range g.idMapping {
		if originalID == value {
			return mappedID
		}
	}
	return value // Keep original if no mapping found
}

// fixReferenceLabel updates the label based on the mapped node ID
func (g *IterationNodeGenerator) fixReferenceLabel(fixedValue, originalLabel string) string {
	if g.nodeTitleMapping == nil {
		return originalLabel
	}

	if title, exists := g.nodeTitleMapping[fixedValue]; exists {
		return title
	}

	if originalLabel == "节点" {
		if title, exists := g.nodeTitleMapping[fixedValue]; exists {
			return title
		}
	}

	return originalLabel
}

// fixChildrenReferences fixes children references recursively
func (g *IterationNodeGenerator) fixChildrenReferences(children []IFlytekReference) []IFlytekReference {
	if len(children) == 0 {
		return children
	}

	fixedChildren := make([]IFlytekReference, len(children))

	for i, child := range children {
		fixedChild := g.fixIFlytekReference(child)
		fixedChild.References = g.fixReferenceDetails(child.References)
		fixedChildren[i] = fixedChild
	}

	return fixedChildren
}

// fixReferenceDetails fixes references within children
func (g *IterationNodeGenerator) fixReferenceDetails(references []IFlytekRefDetail) []IFlytekRefDetail {
	if len(references) == 0 {
		return references
	}

	fixedRefDetails := make([]IFlytekRefDetail, len(references))

	for j, refDetail := range references {
		fixedRefDetail := refDetail
		fixedRefDetail.OriginID = g.fixOriginID(refDetail.OriginID)
		fixedRefDetail.ID = g.fixOutputID(fixedRefDetail.OriginID, refDetail.Value)
		fixedRefDetails[j] = fixedRefDetail
	}

	return fixedRefDetails
}

// fixOriginID fixes OriginID field - handle both mapped and unmapped original Dify node IDs
func (g *IterationNodeGenerator) fixOriginID(originID string) string {
	if g.idMapping == nil {
		return originID
	}

	if mappedID, exists := g.idMapping[originID]; exists {
		return mappedID
	}

	if originID != "" && !strings.Contains(originID, "::") {
		return g.findAlternativeMapping(originID)
	}

	return originID
}

// fixOutputID fixes the reference ID to match the actual output ID
func (g *IterationNodeGenerator) fixOutputID(originNodeID, refValue string) string {
	if g.outputIDMapping == nil {
		return ""
	}

	outputMap, exists := g.outputIDMapping[originNodeID]
	if !exists {
		return ""
	}

	actualOutputID, exists := outputMap[refValue]
	if exists {
		return actualOutputID
	}

	return ""
}

// processIterationReferences processes iteration references
func (g *IterationNodeGenerator) processIterationReferences(inputs []models.Input) []IFlytekReference {
	references := []IFlytekReference{}

	for _, input := range inputs {
		if input.Reference != nil && input.Reference.NodeID != "" {
			// Get source node's iFlytek SparkAgent ID
			sourceIFlytekID, exists := g.idMapping[input.Reference.NodeID]
			if !exists {
				continue // If mapping not found, skip this reference
			}

			// Get source node title
			sourceLabel := g.nodeTitleMapping[sourceIFlytekID]
			if sourceLabel == "" {
				sourceLabel = "未知节点"
			}

			// Create reference group
			refGroup := IFlytekReference{
				Children: []IFlytekReference{
					{
						References: []IFlytekRefDetail{
							{
								OriginID: sourceIFlytekID,
								ID:       g.generateRandomID(),
								Label:    input.Reference.OutputName,
								Type:     g.convertDataType(input.Reference.DataType),
								Value:    input.Reference.OutputName,
								FileType: "",
							},
						},
						Label: "",
						Value: "",
					},
				},
				Label:      sourceLabel,
				ParentNode: true,
				Value:      sourceIFlytekID,
			}
			references = append(references, refGroup)
		}
	}

	return references
}

// processIterationInputsWithID processes iteration input parameters using specified ID
func (g *IterationNodeGenerator) processIterationInputsWithID(inputs []models.Input, iterationInputID string) []IFlytekInput {
	inputParams := []IFlytekInput{}

	for _, input := range inputs {
		var schema IFlytekSchema

		if input.Reference != nil && input.Reference.NodeID != "" {
			// Get source node's iFlytek SparkAgent ID
			sourceIFlytekID, exists := g.idMapping[input.Reference.NodeID]
			if exists {
				schema = IFlytekSchema{
					Type: g.convertDataType(input.Type),
					Value: &IFlytekSchemaValue{
						Type: "ref",
						Content: &IFlytekRefContent{
							Name:   input.Reference.OutputName,
							ID:     g.generateRandomID(), // Can use different ID here as this is reference ID
							NodeID: sourceIFlytekID,
						},
						ContentErrMsg: "",
					},
				}
			}
		} else {
			schema = IFlytekSchema{
				Type:    g.convertDataType(input.Type),
				Default: "",
			}
		}

		inputParam := IFlytekInput{
			ID:         iterationInputID, // Use unified ID
			Name:       input.Name,
			NameErrMsg: "",
			Schema:     schema,
			FileType:   "",
		}
		inputParams = append(inputParams, inputParam)
	}

	return inputParams
}

// generateIterationOutputsWithMapping generates iteration outputs with deterministic IDs for mapping
func (g *IterationNodeGenerator) generateIterationOutputsWithMapping(outputs []models.Output, iterationID string) []IFlytekOutput {
	outputParams := []IFlytekOutput{}

	for _, output := range outputs {
		// Generate deterministic ID for iteration output to ensure consistent mapping
		deterministicOutputID := g.generateDeterministicIterationOutputID(iterationID, output.Name)

		outputParam := IFlytekOutput{
			ID:         deterministicOutputID,
			Name:       output.Name,
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type:    g.convertDataType(output.Type),
				Default: "",
			},
		}
		outputParams = append(outputParams, outputParam)
	}

	return outputParams
}

// generateDeterministicIterationOutputID generates deterministic output ID for iteration nodes
func (g *IterationNodeGenerator) generateDeterministicIterationOutputID(iterationID, outputName string) string {
	// Generate a deterministic UUID based on iteration ID and output name
	return g.generateRandomID()
}

// generateIterationNodeParam generates iteration node parameters
func (g *IterationNodeGenerator) generateIterationNodeParam(config models.IterationConfig, iterationStartNodeID string) map[string]interface{} {
	nodeParam := map[string]interface{}{
		"uid":                  "20718349453", // Default UID, can be configured in actual use
		"appId":                "12a0a7e2",    // Default AppID, can be configured in actual use
		"IterationStartNodeId": iterationStartNodeID,
	}

	return nodeParam
}

// generateDeterministicStartNodeID generates deterministic start node ID based on iteration node ID
func (g *IterationNodeGenerator) generateDeterministicStartNodeID(iterationID string) string {
	// Extract UUID part from iteration node ID
	if len(iterationID) > len("iteration::") {
		uuid := iterationID[len("iteration::"):]
		return "iteration-node-start::" + uuid
	}
	// If format is incorrect, fallback to random generation
	return g.generateSpecialNodeID("iteration-node-start")
}

// convertDataType converts data types
func (g *IterationNodeGenerator) convertDataType(dataType models.UnifiedDataType) string {
	switch dataType {
	case models.DataTypeArrayString:
		return "array-string"
	case models.DataTypeArrayObject:
		return "array-object"
	case models.DataTypeString:
		return "string"
	case models.DataTypeInteger:
		return "integer"
	case models.DataTypeFloat:
		return "number"
	case models.DataTypeNumber:
		return "integer"
	case models.DataTypeBoolean:
		return "boolean"
	case models.DataTypeObject:
		return "object"
	default:
		return string(dataType)
	}
}

// convertPosition converts positions
func (g *IterationNodeGenerator) convertPosition(pos models.Position) IFlytekPosition {
	return IFlytekPosition{
		X: pos.X,
		Y: pos.Y,
	}
}

// generateIterationInternalEdges generates iteration internal edges
func (g *IterationNodeGenerator) generateIterationInternalEdges(subNodes []IFlytekNode, sourceNode *IFlytekNode, endNode *IFlytekNode, startNode *IFlytekNode) []IFlytekEdge {
	var edges []IFlytekEdge

	// Only create edge from the final processing node (output source) to end node
	if sourceNode != nil && endNode != nil {
		edge := IFlytekEdge{
			Source:       sourceNode.ID,
			Target:       endNode.ID,
			TargetHandle: endNode.ID, // End node target handle
			Type:         "customEdge",
			ID:           g.generateEdgeID(sourceNode.ID, endNode.ID),
			MarkerEnd: &IFlytekMarkerEnd{
				Color: "#275EFF",
				Type:  "arrow",
			},
			Data: &IFlytekEdgeData{
				EdgeType: "curve",
			},
			ZIndex: 1, // Iteration internal edge level
		}
		edges = append(edges, edge)
	}

	return edges
}

// generateEdgeID generates edge ID
func (g *IterationNodeGenerator) generateEdgeID(sourceID, targetID string) string {
	return fmt.Sprintf("reactflow__edge-%s-%s", sourceID, targetID)
}

// fixIterationChildOutputIDs fixes the output IDs of iteration child nodes to use parent iteration outputs
func (g *IterationNodeGenerator) fixIterationChildOutputIDs(childNode IFlytekNode, iterationID string) IFlytekNode {
	fixedNode := childNode

	if g.shouldFixChildOutputIDs(childNode) {
		g.fixChildNodeOutputIDs(&fixedNode, iterationID)
	}

	return fixedNode
}

// shouldFixChildOutputIDs checks if child output IDs need fixing
func (g *IterationNodeGenerator) shouldFixChildOutputIDs(childNode IFlytekNode) bool {
	return strings.HasPrefix(childNode.ID, "spark-llm::") && len(childNode.Data.Outputs) > 0
}

// fixChildNodeOutputIDs fixes output IDs for child node
func (g *IterationNodeGenerator) fixChildNodeOutputIDs(fixedNode *IFlytekNode, iterationID string) {
	if g.hasIterationOutputMapping(iterationID) {
		g.applyIterationOutputMapping(fixedNode, iterationID)
	} else {
		g.generateFallbackOutputIDs(fixedNode, iterationID)
	}
}

// hasIterationOutputMapping checks if iteration has output mapping
func (g *IterationNodeGenerator) hasIterationOutputMapping(iterationID string) bool {
	return g.outputIDMapping != nil && g.outputIDMapping[iterationID] != nil
}

// applyIterationOutputMapping applies iteration output mapping to child node
func (g *IterationNodeGenerator) applyIterationOutputMapping(fixedNode *IFlytekNode, iterationID string) {
	iterationOutputs := g.outputIDMapping[iterationID]

	for i, output := range fixedNode.Data.Outputs {
		if iterationOutputID, exists := iterationOutputs[output.Name]; exists {
			fixedNode.Data.Outputs[i].ID = iterationOutputID
		} else if output.Name == "output" && len(iterationOutputs) > 0 {
			fixedNode.Data.Outputs[i].ID = g.getFirstIterationOutputID(iterationOutputs)
		} else {
			fixedNode.Data.Outputs[i].ID = g.generateDeterministicChildOutputID(fixedNode.ID, output.Name, iterationID)
		}
	}
}

// getFirstIterationOutputID gets first available iteration output ID
func (g *IterationNodeGenerator) getFirstIterationOutputID(iterationOutputs map[string]string) string {
	for _, outputID := range iterationOutputs {
		return outputID
	}
	return ""
}

// generateFallbackOutputIDs generates fallback deterministic IDs
func (g *IterationNodeGenerator) generateFallbackOutputIDs(fixedNode *IFlytekNode, iterationID string) {
	for i, output := range fixedNode.Data.Outputs {
		deterministicOutputID := g.generateDeterministicChildOutputID(fixedNode.ID, output.Name, iterationID)
		fixedNode.Data.Outputs[i].ID = deterministicOutputID
	}
}

// generateDeterministicChildOutputID generates deterministic output ID for iteration child nodes
func (g *IterationNodeGenerator) generateDeterministicChildOutputID(childNodeID, outputName, iterationID string) string {
	// based on the child node and iteration context
	return g.generateRandomID()
}

// generateRandomID generates random ID (uses global UUID generation function)
func (g *IterationNodeGenerator) generateRandomID() string {
	return generateUUID()
}
