package generator

import (
	"encoding/json"
	"fmt"
	"github.com/iflytek/agentbridge/core/interfaces"
	"github.com/iflytek/agentbridge/internal/models"
	"github.com/iflytek/agentbridge/platforms/common"
	"strings"

	"gopkg.in/yaml.v3"
)

// Platform provider constants
const (
	ProviderCoze  = "coze"
	ProviderDify  = "dify"
	ProviderEmpty = ""
)

// Classifier mapping constants
const (
	DefaultSourceHandle = "default"
	DefaultIntentKey    = "__default__"
)

// compile-time interface verification
var _ interfaces.DSLGenerator = (*IFlytekGenerator)(nil)

// BranchMapping contains branch mapping information
type BranchMapping struct {
	TrueBranchID  string            // true branch ID (for backward compatibility)
	FalseBranchID string            // false/else branch ID (for backward compatibility)
	BranchIDs     map[string]string // case ID -> branch_one_of ID mapping (for multi-branch support)
}

// ClassifierMapping contains classifier mapping information
type ClassifierMapping struct {
	IntentIDs         []string          // Intent ID list stored in order
	DefaultIntentID   string            // Default intent ID
	FirstIntentTarget string            // First intent's target node
	ClassIDToIntentID map[string]string // Complete mapping from class ID/number handle to intent ID
}

// IFlytekGenerator iFlytek Agent DSL generator
type IFlytekGenerator struct {
	*common.BaseGenerator
	factory                 *NodeGeneratorFactory
	idMapping               map[string]string                   // Dify ID -> iFlytek SparkAgent ID mapping
	nodeTitleMapping        map[string]string                   // iFlytek SparkAgent ID -> node title mapping
	conditionBranchMapping  map[string]*BranchMapping           // Condition node ID -> branch ID mapping
	classifierIntentMapping map[string]*ClassifierMapping       // Classifier node ID -> intent mapping
	classifierGenerators    map[string]*ClassifierNodeGenerator // Classifier generator cache
	iterationSubNodeMapping map[string]map[string]string        // Iteration main node ID -> sub-node type -> sub-node ID mapping
	currentDSL              *models.UnifiedDSL                  // Current DSL being processed
	sourcePlatform          models.PlatformType                 // Source platform identification
}

func NewIFlytekGenerator() *IFlytekGenerator {
	return &IFlytekGenerator{
		BaseGenerator:           common.NewBaseGenerator(models.PlatformIFlytek),
		factory:                 NewNodeGeneratorFactory(),
		idMapping:               make(map[string]string),
		nodeTitleMapping:        make(map[string]string),
		conditionBranchMapping:  make(map[string]*BranchMapping),
		classifierIntentMapping: make(map[string]*ClassifierMapping),
		classifierGenerators:    make(map[string]*ClassifierNodeGenerator),
		iterationSubNodeMapping: make(map[string]map[string]string),
	}
}

// Generate generates iFlytek SparkAgent DSL from unified format
func (g *IFlytekGenerator) Generate(unifiedDSL *models.UnifiedDSL) ([]byte, error) {
	// Store DSL for use in generators
	g.currentDSL = unifiedDSL

	// Identify source platform
	g.sourcePlatform = g.identifySourcePlatform(unifiedDSL)

	// Validate input
	if err := g.Validate(unifiedDSL); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Build iFlytek SparkAgent DSL
	iflytekDSL := IFlytekDSL{
		FlowMeta: g.generateFlowMeta(unifiedDSL),
		FlowData: IFlytekFlowData{
			Nodes: []IFlytekNode{},
			Edges: []IFlytekEdge{},
		},
	}

	// No longer need to mark and skip nodes - all Dify nodes should be converted

	// Generate nodes
	if err := g.generateNodes(unifiedDSL.Workflow.Nodes, &iflytekDSL); err != nil {
		return nil, fmt.Errorf("failed to generate nodes: %w", err)
	}

	// Before generating edges, first analyze classifier target node mapping
	g.analyzeClassifierTargets(unifiedDSL.Workflow.Edges)

	// Generate connections
	if err := g.generateEdges(unifiedDSL.Workflow.Edges, &iflytekDSL); err != nil {
		return nil, fmt.Errorf("failed to generate edges: %w", err)
	}

	// Add default intent connections ONLY for Dify to Spark conversion (let default intent connect to first intent's target node)
	// Skip this for Coze platform as it already has proper default intent mapping
	if g.sourcePlatform == models.PlatformDify {
		g.generateDefaultIntentEdges(unifiedDSL.Workflow.Edges, &iflytekDSL)
	}

	// Serialize to YAML
	data, err := yaml.Marshal(iflytekDSL)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return data, nil
}

// identifySourcePlatform identifies the source platform from unified DSL
func (g *IFlytekGenerator) identifySourcePlatform(unifiedDSL *models.UnifiedDSL) models.PlatformType {
	// Check platform metadata for source platform identification
	if unifiedDSL.PlatformMetadata.Coze != nil {
		return models.PlatformCoze
	}

	if unifiedDSL.PlatformMetadata.Dify != nil {
		return models.PlatformDify
	}

	if unifiedDSL.PlatformMetadata.IFlytek != nil {
		return models.PlatformIFlytek
	}

	// Fallback: try to identify from node structure patterns
	// This is a backup method if platform metadata is not available
	sourcePlatform := g.identifyFromNodePatterns(unifiedDSL)
	return sourcePlatform
}

// identifyFromNodePatterns identifies platform from node structure patterns
func (g *IFlytekGenerator) identifyFromNodePatterns(unifiedDSL *models.UnifiedDSL) models.PlatformType {
	// Look for platform-specific patterns in node configurations
	for _, node := range unifiedDSL.Workflow.Nodes {
		// Check for Coze-specific patterns
		if g.hasCozePatterns(node) {
			return models.PlatformCoze
		}

		// Check for Dify-specific patterns
		if g.hasDifyPatterns(node) {
			return models.PlatformDify
		}
	}

	// Default to Dify if no specific patterns found (for backward compatibility)
	return models.PlatformDify
}

// hasCozePatterns checks for Coze-specific patterns in node configuration
func (g *IFlytekGenerator) hasCozePatterns(node models.Node) bool {
	// Check for Coze-specific platform config
	if node.PlatformConfig.Coze != nil && len(node.PlatformConfig.Coze) > 0 {
		return true
	}

	// Check for Coze-specific node structure patterns
	// Coze typically has more structured metadata and specific field patterns
	if node.Type == models.NodeTypeClassifier {
		if classifierConfig, ok := common.AsClassifierConfig(node.Config); ok && classifierConfig != nil {
			// Coze classifiers usually have more structured model configurations
			if classifierConfig.Model.Provider == ProviderCoze {
				return true
			}
		}
	}

	return false
}

// hasDifyPatterns checks for Dify-specific patterns in node configuration
func (g *IFlytekGenerator) hasDifyPatterns(node models.Node) bool {
	// Check for Dify-specific platform config
	if node.PlatformConfig.Dify != nil && len(node.PlatformConfig.Dify) > 0 {
		return true
	}

	// Check for Dify-specific node structure patterns
	if node.Type == models.NodeTypeClassifier {
		if classifierConfig, ok := common.AsClassifierConfig(node.Config); ok && classifierConfig != nil {
			// Dify classifiers usually have simpler configurations
			if classifierConfig.Model.Provider == ProviderDify || classifierConfig.Model.Provider == ProviderEmpty {
				return true
			}
		}
	}

	return false
}

// Validate validates if unified DSL meets iFlytek SparkAgent requirements
func (g *IFlytekGenerator) Validate(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL cannot be nil")
	}

	if unifiedDSL.Metadata.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(unifiedDSL.Workflow.Nodes) == 0 {
		return fmt.Errorf("workflow must have at least one node")
	}

	return nil
}

// generateFlowMeta generates flow metadata
func (g *IFlytekGenerator) generateFlowMeta(unifiedDSL *models.UnifiedDSL) IFlytekFlowMeta {
	meta := IFlytekFlowMeta{
		Name:        unifiedDSL.Metadata.Name,
		Description: unifiedDSL.Metadata.Description,
		DSLVersion:  "v1",
	}

	// Restore iFlytek Platform specific configurations
	if unifiedDSL.PlatformMetadata.IFlytek != nil {
		iflytekMeta := unifiedDSL.PlatformMetadata.IFlytek
		meta.AvatarIcon = iflytekMeta.AvatarIcon
		meta.AvatarColor = iflytekMeta.AvatarColor
		meta.AdvancedConfig = iflytekMeta.AdvancedConfig
		meta.DSLVersion = iflytekMeta.DSLVersion
	} else {
		// If no iFlytek specific configuration, generate from UI configuration
		meta.AvatarColor = "#FFEAD5" // Default color
		meta.AdvancedConfig = g.generateAdvancedConfig(unifiedDSL.Metadata.UIConfig)

		// If there is Dify configuration, try to convert icon
		if unifiedDSL.PlatformMetadata.Dify != nil && unifiedDSL.PlatformMetadata.Dify.Icon != "" {
			meta.AvatarIcon = unifiedDSL.PlatformMetadata.Dify.Icon
		} else {
			// Use default hardcoded avatar icon for iFlytek platform
			meta.AvatarIcon = "https://oss-beijing-m8.openstorage.cn/SparkBotProd/icon/common/emojiitem_00_10@2x.png"
		}
	}

	return meta
}

// generateAdvancedConfig generates advanced configuration
func (g *IFlytekGenerator) generateAdvancedConfig(uiConfig *models.UIConfig) string {
	if uiConfig == nil {
		return `{"prologue":{"enabled":true,"inputExample":["","",""]},"needGuide":false}`
	}

	config := map[string]interface{}{
		"prologue": map[string]interface{}{
			"enabled": false,
		},
		"needGuide": false,
	}

	// If there is an opening statement or suggested questions, enable prologue
	if uiConfig.OpeningStatement != "" || len(uiConfig.SuggestedQuestions) > 0 {
		prologue := map[string]interface{}{
			"enabled": true,
		}

		if uiConfig.OpeningStatement != "" {
			prologue["statement"] = uiConfig.OpeningStatement
		}

		if len(uiConfig.SuggestedQuestions) > 0 {
			// Max 3 suggested questions
			questions := uiConfig.SuggestedQuestions
			if len(questions) > 3 {
				questions = questions[:3]
			}

			// Ensure there are 3 elements (fill with empty strings)
			for len(questions) < 3 {
				questions = append(questions, "")
			}

			prologue["inputExample"] = questions
		} else {
			prologue["inputExample"] = []string{"", "", ""}
		}

		config["prologue"] = prologue
	}

	// Serialize to JSON string
	jsonData, err := json.Marshal(config)
	if err != nil {
		return `{"prologue":{"enabled":true,"inputExample":["","",""]},"needGuide":false}`
	}

	return string(jsonData)
}

// isIterationSubNode checks if a node is a sub-node within an iteration
func (g *IFlytekGenerator) isIterationSubNode(node models.Node) bool {
	checkers := g.getIterationCheckFunctions()

	if checker, exists := checkers[node.Type]; exists {
		return checker(node)
	}

	return false
}

// getIterationCheckFunctions returns node type to iteration check function mapping
func (g *IFlytekGenerator) getIterationCheckFunctions() map[models.NodeType]func(models.Node) bool {
	return map[models.NodeType]func(models.Node) bool{
		models.NodeTypeStart:      g.checkStartNodeIteration,
		models.NodeTypeCode:       g.checkCodeNodeIteration,
		models.NodeTypeLLM:        g.checkLLMNodeIteration,
		models.NodeTypeCondition:  g.checkConditionNodeIteration,
		models.NodeTypeClassifier: g.checkClassifierNodeIteration,
	}
}

// checkStartNodeIteration checks if start node is in iteration
func (g *IFlytekGenerator) checkStartNodeIteration(node models.Node) bool {
	if startConfig, ok := common.AsStartConfig(node.Config); ok && startConfig != nil {
		return startConfig.IsInIteration
	}
	return false
}

// checkCodeNodeIteration checks if code node is in iteration
func (g *IFlytekGenerator) checkCodeNodeIteration(node models.Node) bool {
	if codeConfig, ok := common.AsCodeConfig(node.Config); ok && codeConfig != nil {
		return codeConfig.IsInIteration
	}
	return false
}

// checkLLMNodeIteration checks if LLM node is in iteration
func (g *IFlytekGenerator) checkLLMNodeIteration(node models.Node) bool {
	if llmConfig, ok := common.AsLLMConfig(node.Config); ok && llmConfig != nil {
		return llmConfig.IsInIteration
	}
	return false
}

// checkConditionNodeIteration checks if condition node is in iteration
func (g *IFlytekGenerator) checkConditionNodeIteration(node models.Node) bool {
	if conditionConfig, ok := common.AsConditionConfig(node.Config); ok && conditionConfig != nil {
		return conditionConfig.IsInIteration
	}
	return false
}

// checkClassifierNodeIteration checks if classifier node is in iteration
func (g *IFlytekGenerator) checkClassifierNodeIteration(node models.Node) bool {
	if classifierConfig, ok := common.AsClassifierConfig(node.Config); ok && classifierConfig != nil {
		return classifierConfig.IsInIteration
	}
	return false
}

// establishIterationSubNodeMappings establishes ID mappings for iteration sub-nodes
func (g *IFlytekGenerator) establishIterationSubNodeMappings(iterationNode models.Node, generatedSubNodes []IFlytekNode, originalSubNodes []models.Node) {
	// Establish mapping for the original iteration start node
	for _, originalNode := range originalSubNodes {
		if originalNode.Type == models.NodeTypeStart {
			// Find the generated iteration start node (the first one, and its ID starts with iteration-node-start::)
			for _, generatedNode := range generatedSubNodes {
				if strings.HasPrefix(generatedNode.ID, "iteration-node-start::") {
					g.idMapping[originalNode.ID] = generatedNode.ID
					g.nodeTitleMapping[generatedNode.ID] = originalNode.Title
					break
				}
			}
		}
	}

	// Establish mappings for other iteration sub-nodes (code nodes, LLM nodes, etc.)
	for _, originalNode := range originalSubNodes {
		if originalNode.Type != models.NodeTypeStart {
			// Match generated nodes based on node type and title
			for _, generatedNode := range generatedSubNodes {
				if g.isMatchingIterationSubNode(originalNode, generatedNode) {
					g.idMapping[originalNode.ID] = generatedNode.ID
					g.nodeTitleMapping[generatedNode.ID] = originalNode.Title
					break
				}
			}
		}
	}
}

// isMatchingIterationSubNode checks if an original node and a generated node match
func (g *IFlytekGenerator) isMatchingIterationSubNode(originalNode models.Node, generatedNode IFlytekNode) bool {
	// Match based on node type
	switch originalNode.Type {
	case models.NodeTypeCode:
		return strings.HasPrefix(generatedNode.ID, "ifly-code::") &&
			generatedNode.Data.Label == originalNode.Title
	case models.NodeTypeLLM:
		return strings.HasPrefix(generatedNode.ID, "spark-llm::") &&
			generatedNode.Data.Label == originalNode.Title
	case models.NodeTypeCondition:
		return strings.HasPrefix(generatedNode.ID, "if-else::") &&
			generatedNode.Data.Label == originalNode.Title
	case models.NodeTypeClassifier:
		return strings.HasPrefix(generatedNode.ID, "decision-making::") &&
			generatedNode.Data.Label == originalNode.Title
	default:
		return false
	}
}

// generateNodes generates nodes
func (g *IFlytekGenerator) generateNodes(nodes []models.Node, iflytekDSL *IFlytekDSL) error {
	// First round: generate all nodes and establish ID mappings
	if err := g.performFirstRoundNodeGeneration(nodes, iflytekDSL); err != nil {
		return err
	}

	// Second round: regenerate nodes with references
	if err := g.performSecondRoundRegeneration(nodes, iflytekDSL); err != nil {
		return err
	}

	// Third round: final refinement of node references
	if err := g.performThirdRoundRefinement(nodes, iflytekDSL); err != nil {
		return err
	}

	return nil
}

// performFirstRoundNodeGeneration handles the first round of node generation
func (g *IFlytekGenerator) performFirstRoundNodeGeneration(nodes []models.Node, iflytekDSL *IFlytekDSL) error {
	for _, node := range nodes {
		if g.isIterationSubNode(node) {
			continue
		}

		if err := g.generateAndProcessSingleNode(node, nodes, iflytekDSL); err != nil {
			return err
		}
	}
	return nil
}

// generateAndProcessSingleNode generates a single node and handles its special processing
func (g *IFlytekGenerator) generateAndProcessSingleNode(node models.Node, allNodes []models.Node, iflytekDSL *IFlytekDSL) error {
	// Generate basic node
	generator, err := g.factory.GetGenerator(node.Type)
	if err != nil {
		return fmt.Errorf("failed to get generator for node %s: %w", node.ID, err)
	}

	// Set DSL context and mappings for condition generators to enable proper type inference
	if condGen, ok := generator.(*ConditionNodeGenerator); ok {
		condGen.SetUnifiedDSL(g.currentDSL)
		condGen.SetIDMapping(g.idMapping)
		condGen.SetNodeTitleMapping(g.nodeTitleMapping)
	}

	iflytekNode, err := generator.GenerateNode(node)
	if err != nil {
		return fmt.Errorf("failed to generate node %s: %w", node.ID, err)
	}

	// Establish basic mappings
	g.establishNodeMappings(node, iflytekNode)

	// Handle node type-specific processing
	if err := g.handleNodeTypeSpecificProcessing(node, iflytekNode, allNodes, generator, iflytekDSL); err != nil {
		return err
	}

	// Add node to DSL
	iflytekDSL.FlowData.Nodes = append(iflytekDSL.FlowData.Nodes, iflytekNode)
	return nil
}

// establishNodeMappings establishes basic node mappings
func (g *IFlytekGenerator) establishNodeMappings(node models.Node, iflytekNode IFlytekNode) {
	g.idMapping[node.ID] = iflytekNode.ID
	g.nodeTitleMapping[iflytekNode.ID] = node.Title
}

// handleNodeTypeSpecificProcessing handles processing specific to different node types
func (g *IFlytekGenerator) handleNodeTypeSpecificProcessing(node models.Node, iflytekNode IFlytekNode, allNodes []models.Node, generator NodeGenerator, iflytekDSL *IFlytekDSL) error {
	switch node.Type {
	case models.NodeTypeCondition:
		g.handleConditionNodeSpecialProcessing(generator, iflytekNode)
	case models.NodeTypeClassifier:
		g.handleClassifierNodeSpecialProcessing(generator, iflytekNode)
	case models.NodeTypeIteration:
		return g.handleIterationNodeSpecialProcessing(node, iflytekNode, allNodes, generator, iflytekDSL)
	}
	return nil
}

// handleConditionNodeSpecialProcessing handles condition node special processing
func (g *IFlytekGenerator) handleConditionNodeSpecialProcessing(generator NodeGenerator, iflytekNode IFlytekNode) {
	if conditionGen, ok := generator.(*ConditionNodeGenerator); ok {
		// Get branch ID mapping from the condition generator
		branchIDMapping := conditionGen.GetBranchIDMapping()

		// Create BranchMapping structure for the main generator
		mapping := &BranchMapping{
			BranchIDs: make(map[string]string),
		}

		// Copy all mappings from condition generator
		for key, value := range branchIDMapping {
			mapping.BranchIDs[key] = value
		}

		// Set compatibility mappings for true/false
		g.setCompatibilityMappings(mapping, branchIDMapping)

		// Save the mapping
		g.conditionBranchMapping[iflytekNode.ID] = mapping
	}
}

// setCompatibilityMappings sets backward compatibility mappings for true/false handles
func (g *IFlytekGenerator) setCompatibilityMappings(mapping *BranchMapping, branchIDMapping map[string]string) {
	// Set TrueBranchID and FalseBranchID for backward compatibility
	if branchID, exists := branchIDMapping["1"]; exists {
		mapping.TrueBranchID = branchID
		mapping.BranchIDs["true"] = branchID
	}

	if branchID, exists := branchIDMapping["__default__"]; exists {
		mapping.FalseBranchID = branchID
		mapping.BranchIDs["false"] = branchID
	}
}

// handleClassifierNodeSpecialProcessing handles classifier node special processing
func (g *IFlytekGenerator) handleClassifierNodeSpecialProcessing(generator NodeGenerator, iflytekNode IFlytekNode) {
	if classifierGen, ok := generator.(*ClassifierNodeGenerator); ok {
		g.classifierGenerators[iflytekNode.ID] = classifierGen
	}
	g.extractClassifierMapping(iflytekNode)
}

// handleIterationNodeSpecialProcessing handles iteration node special processing
func (g *IFlytekGenerator) handleIterationNodeSpecialProcessing(node models.Node, iflytekNode IFlytekNode, allNodes []models.Node, generator NodeGenerator, iflytekDSL *IFlytekDSL) error {
	iterationGen, ok := generator.(*IterationNodeGenerator)
	if !ok {
		return nil
	}

	// Find iteration sub-nodes
	iterationSubNodes := g.findIterationSubNodes(allNodes, node.ID)

	// Extract iteration start node ID
	iterationStartNodeID := g.extractIterationStartNodeID(iflytekNode)

	// Generate iteration components
	subNodes, iterationEdges, err := iterationGen.GenerateIterationSubNodes(node, iflytekNode.ID, iterationSubNodes, iterationStartNodeID)
	if err != nil {
		return fmt.Errorf("failed to generate iteration sub-nodes for %s: %w", node.ID, err)
	}

	// Add components to DSL and establish mappings
	g.addIterationComponentsToDSL(iflytekDSL, subNodes, iterationEdges)
	g.registerIterationGeneratorsForSubNodes(iterationSubNodes, subNodes)
	g.establishIterationSubNodeMappings(node, subNodes, iterationSubNodes)

	return nil
}

// extractIterationStartNodeID extracts iteration start node ID from node parameters
func (g *IFlytekGenerator) extractIterationStartNodeID(iflytekNode IFlytekNode) string {
	if nodeParam, ok := iflytekNode.Data.NodeParam["IterationStartNodeId"].(string); ok {
		return nodeParam
	}
	return ""
}

// addIterationComponentsToDSL adds iteration sub-nodes and edges to DSL
func (g *IFlytekGenerator) addIterationComponentsToDSL(iflytekDSL *IFlytekDSL, subNodes []IFlytekNode, iterationEdges []IFlytekEdge) {
	iflytekDSL.FlowData.Nodes = append(iflytekDSL.FlowData.Nodes, subNodes...)
	iflytekDSL.FlowData.Edges = append(iflytekDSL.FlowData.Edges, iterationEdges...)
}

// registerIterationGeneratorsForSubNodes registers generators for iteration sub-nodes
func (g *IFlytekGenerator) registerIterationGeneratorsForSubNodes(iterationSubNodes []models.Node, subNodes []IFlytekNode) {
	g.registerIterationSubNodeClassifiers(iterationSubNodes, subNodes)
	g.registerIterationSubNodeConditions(iterationSubNodes, subNodes)
}

// performSecondRoundRegeneration handles the second round of node regeneration
func (g *IFlytekGenerator) performSecondRoundRegeneration(nodes []models.Node, iflytekDSL *IFlytekDSL) error {
	// Set up mappings in factory
	g.factory.SetIDMapping(g.idMapping)
	g.factory.SetNodeTitleMapping(g.nodeTitleMapping)

	// Regenerate nodes that need references
	for _, node := range nodes {
		if g.shouldSkipNodeInSecondRound(node) {
			continue
		}

		if g.needsRegeneration(node) {
			if err := g.regenerateNodeWithReferences(node, iflytekDSL); err != nil {
				return err
			}
		}
	}

	return nil
}

// shouldSkipNodeInSecondRound determines if a node should be skipped in second round
func (g *IFlytekGenerator) shouldSkipNodeInSecondRound(node models.Node) bool {
	return g.isIterationSubNode(node)
}

// regenerateNodeWithReferences regenerates a node with updated references
func (g *IFlytekGenerator) regenerateNodeWithReferences(node models.Node, iflytekDSL *IFlytekDSL) error {
	generator, err := g.factory.GetGenerator(node.Type)
	if err != nil {
		return fmt.Errorf("failed to get generator for node %s: %w", node.ID, err)
	}

	iflytekNode, err := generator.GenerateNode(node)
	if err != nil {
		return fmt.Errorf("failed to regenerate node %s: %w", node.ID, err)
	}

	// Preserve original ID and handle special processing
	iflytekNode.ID = g.idMapping[node.ID]
	g.handleRegenerationSpecialProcessing(node, iflytekNode, generator)

	// Replace node in DSL
	g.replaceNodeInDSL(iflytekDSL, iflytekNode)
	return nil
}

// handleRegenerationSpecialProcessing handles special processing during regeneration
func (g *IFlytekGenerator) handleRegenerationSpecialProcessing(node models.Node, iflytekNode IFlytekNode, generator NodeGenerator) {
	switch node.Type {
	case models.NodeTypeCondition:
		g.handleConditionNodeSpecialProcessing(generator, iflytekNode)
	case models.NodeTypeClassifier:
		if classifierGen, ok := generator.(*ClassifierNodeGenerator); ok {
			g.classifierGenerators[iflytekNode.ID] = classifierGen
		}
		g.extractClassifierMapping(iflytekNode)
	}
}

// replaceNodeInDSL replaces a node in the DSL by ID lookup
func (g *IFlytekGenerator) replaceNodeInDSL(iflytekDSL *IFlytekDSL, iflytekNode IFlytekNode) {
	for j, existingNode := range iflytekDSL.FlowData.Nodes {
		if existingNode.ID == iflytekNode.ID {
			iflytekDSL.FlowData.Nodes[j] = iflytekNode
			break
		}
	}
}

// performThirdRoundRefinement handles the third round of node refinement
func (g *IFlytekGenerator) performThirdRoundRefinement(nodes []models.Node, iflytekDSL *IFlytekDSL) error {
	nodeTypesToRefine := []models.NodeType{
		models.NodeTypeEnd, models.NodeTypeLLM, models.NodeTypeCondition,
		models.NodeTypeCode, models.NodeTypeIteration,
	}

	for _, node := range nodes {
		if g.shouldSkipNodeInThirdRound(node, nodeTypesToRefine) {
			continue
		}

		if err := g.refineNodeWithFinalReferences(node, iflytekDSL); err != nil {
			// Continue on error for this round (non-critical)
			continue
		}
	}

	return nil
}

// shouldSkipNodeInThirdRound determines if a node should be skipped in third round
func (g *IFlytekGenerator) shouldSkipNodeInThirdRound(node models.Node, nodeTypesToRefine []models.NodeType) bool {
	if g.isIterationSubNode(node) {
		return true
	}

	for _, nodeType := range nodeTypesToRefine {
		if node.Type == nodeType {
			return false
		}
	}
	return true
}

// refineNodeWithFinalReferences performs final refinement of node references
func (g *IFlytekGenerator) refineNodeWithFinalReferences(node models.Node, iflytekDSL *IFlytekDSL) error {
	generator, err := g.factory.GetGenerator(node.Type)
	if err != nil {
		return err
	}

	iflytekNode, err := generator.GenerateNode(node)
	if err != nil {
		return err
	}

	// Preserve ID and handle final mappings
	iflytekNode.ID = g.idMapping[node.ID]
	g.handleFinalRefinementMappings(node, iflytekNode, generator)

	// Replace in DSL
	g.replaceNodeInDSL(iflytekDSL, iflytekNode)
	return nil
}

// handleFinalRefinementMappings handles mappings during final refinement
func (g *IFlytekGenerator) handleFinalRefinementMappings(node models.Node, iflytekNode IFlytekNode, generator NodeGenerator) {
	switch node.Type {
	case models.NodeTypeCondition:
		g.handleConditionNodeSpecialProcessing(generator, iflytekNode)
	case models.NodeTypeClassifier:
		g.extractClassifierMapping(iflytekNode)
	}
}

// registerIterationSubNodeClassifiers registers classifier generators for iteration sub-nodes
func (g *IFlytekGenerator) registerIterationSubNodeClassifiers(originalSubNodes []models.Node, generatedSubNodes []IFlytekNode) {
	for _, generatedNode := range generatedSubNodes {
		if !g.isClassifierNode(generatedNode) {
			continue
		}

		matchedNode := g.findMatchingOriginalClassifierNode(originalSubNodes, generatedNode)
		if matchedNode == nil {
			continue
		}

		g.establishClassifierNodeMapping(matchedNode, generatedNode)

		generator := g.createClassifierGenerator()
		if generator == nil {
			continue
		}

		g.configureGeneratorMappings(generator)
		g.processClassifierNodeGeneration(generator, generatedNode)
	}
}

func (g *IFlytekGenerator) isClassifierNode(generatedNode IFlytekNode) bool {
	return strings.HasPrefix(generatedNode.ID, "decision-making::")
}

func (g *IFlytekGenerator) findMatchingOriginalClassifierNode(originalSubNodes []models.Node, generatedNode IFlytekNode) *models.Node {
	for i := range originalSubNodes {
		if originalSubNodes[i].Type == models.NodeTypeClassifier &&
			originalSubNodes[i].Title == generatedNode.Data.Label {
			return &originalSubNodes[i]
		}
	}
	return nil
}

func (g *IFlytekGenerator) establishClassifierNodeMapping(matchedNode *models.Node, generatedNode IFlytekNode) {
	g.idMapping[matchedNode.ID] = generatedNode.ID
	g.nodeTitleMapping[generatedNode.ID] = matchedNode.Title
}

func (g *IFlytekGenerator) createClassifierGenerator() interface{} {
	generator, err := g.factory.GetGenerator(models.NodeTypeClassifier)
	if err != nil {
		return nil
	}
	return generator
}

func (g *IFlytekGenerator) configureGeneratorMappings(generator interface{}) {
	if mappingSetter, ok := generator.(interface{ SetIDMapping(map[string]string) }); ok {
		mappingSetter.SetIDMapping(g.idMapping)
	}
	if titleSetter, ok := generator.(interface{ SetNodeTitleMapping(map[string]string) }); ok {
		titleSetter.SetNodeTitleMapping(g.nodeTitleMapping)
	}
}

func (g *IFlytekGenerator) processClassifierNodeGeneration(generator interface{}, generatedNode IFlytekNode) {
	classifierGen, ok := generator.(*ClassifierNodeGenerator)
	if !ok {
		return
	}

	intentChains := g.extractIntentChainsFromNode(generatedNode)
	if intentChains == nil {
		return
	}

	classIDToIntentID := g.buildClassIDToIntentIDMapping(intentChains)
	classifierGen.SetClassIDToIntentIDMapping(classIDToIntentID)
	g.classifierGenerators[generatedNode.ID] = classifierGen

	g.updateClassifierIntentMapping(generatedNode.ID, intentChains, classIDToIntentID)
}

func (g *IFlytekGenerator) extractIntentChainsFromNode(generatedNode IFlytekNode) []map[string]interface{} {
	nodeParam, ok := generatedNode.Data.NodeParam["intentChains"]
	if !ok {
		return nil
	}

	intentChains, ok := nodeParam.([]map[string]interface{})
	if !ok {
		return nil
	}

	return intentChains
}

func (g *IFlytekGenerator) buildClassIDToIntentIDMapping(intentChains []map[string]interface{}) map[string]string {
	classIDToIntentID := make(map[string]string)

	for i, intentChain := range intentChains {
		intentID, hasID := intentChain["id"].(string)
		intentType, hasType := intentChain["intentType"].(int)

		if !hasID || !hasType {
			continue
		}

		switch intentType {
		case 2: // Normal classification intent
			difyNumberHandle := fmt.Sprintf("%d", i+1)
			classIDToIntentID[difyNumberHandle] = intentID
		case 1: // Default intent
			classIDToIntentID["__default__"] = intentID
		}
	}

	return classIDToIntentID
}

func (g *IFlytekGenerator) updateClassifierIntentMapping(nodeID string, intentChains []map[string]interface{}, classIDToIntentID map[string]string) {
	if g.classifierIntentMapping == nil {
		g.classifierIntentMapping = make(map[string]*ClassifierMapping)
	}

	intentIDs, defaultIntentID := g.extractIntentIDsFromChains(intentChains)

	newMapping := &ClassifierMapping{
		IntentIDs:         intentIDs,
		DefaultIntentID:   defaultIntentID,
		ClassIDToIntentID: classIDToIntentID,
	}
	g.classifierIntentMapping[nodeID] = newMapping
}

func (g *IFlytekGenerator) extractIntentIDsFromChains(intentChains []map[string]interface{}) ([]string, string) {
	var intentIDs []string
	var defaultIntentID string

	for _, intentChain := range intentChains {
		intentID, hasID := intentChain["id"].(string)
		intentType, hasType := intentChain["intentType"].(int)

		if !hasID || !hasType {
			continue
		}

		switch intentType {
		case 2: // Normal classification intent
			intentIDs = append(intentIDs, intentID)
		case 1: // Default intent
			defaultIntentID = intentID
		}
	}

	return intentIDs, defaultIntentID
}

// registerIterationSubNodeConditions registers condition generators for iteration sub-nodes
func (g *IFlytekGenerator) registerIterationSubNodeConditions(originalSubNodes []models.Node, generatedSubNodes []IFlytekNode) {
	for _, generatedNode := range generatedSubNodes {
		if !g.isConditionNode(generatedNode) {
			continue
		}

		matchedNode := g.findMatchingOriginalConditionNode(originalSubNodes, generatedNode)
		if matchedNode == nil {
			continue
		}

		g.establishConditionNodeMapping(matchedNode, generatedNode)

		generator := g.createConditionGenerator()
		if generator == nil {
			continue
		}

		g.configureGeneratorMappings(generator)
		g.processConditionNodeGeneration(generator, generatedNode, matchedNode)
	}
}

func (g *IFlytekGenerator) isConditionNode(generatedNode IFlytekNode) bool {
	return strings.HasPrefix(generatedNode.ID, "if-else::")
}

func (g *IFlytekGenerator) findMatchingOriginalConditionNode(originalSubNodes []models.Node, generatedNode IFlytekNode) *models.Node {
	for i := range originalSubNodes {
		if originalSubNodes[i].Type == models.NodeTypeCondition &&
			originalSubNodes[i].Title == generatedNode.Data.Label {
			return &originalSubNodes[i]
		}
	}
	return nil
}

func (g *IFlytekGenerator) establishConditionNodeMapping(matchedNode *models.Node, generatedNode IFlytekNode) {
	g.idMapping[matchedNode.ID] = generatedNode.ID
	g.nodeTitleMapping[generatedNode.ID] = matchedNode.Title
}

func (g *IFlytekGenerator) createConditionGenerator() interface{} {
	generator, err := g.factory.GetGenerator(models.NodeTypeCondition)
	if err != nil {
		return nil
	}
	return generator
}

func (g *IFlytekGenerator) processConditionNodeGeneration(generator interface{}, generatedNode IFlytekNode, matchedNode *models.Node) {
	conditionGen, ok := generator.(*ConditionNodeGenerator)
	if !ok {
		return
	}

	cases := g.extractCasesFromNode(generatedNode)
	if cases == nil {
		return
	}

	branchIDMapping := g.buildBranchIDMapping(cases, matchedNode)
	conditionGen.SetBranchIDMapping(branchIDMapping)
	g.extractBranchMappingWithCaseIDs(generatedNode, branchIDMapping)
}

func (g *IFlytekGenerator) extractCasesFromNode(generatedNode IFlytekNode) []map[string]interface{} {
	nodeParam, ok := generatedNode.Data.NodeParam["cases"]
	if !ok {
		return nil
	}

	cases, ok := nodeParam.([]map[string]interface{})
	if !ok {
		return nil
	}

	return cases
}

func (g *IFlytekGenerator) buildBranchIDMapping(cases []map[string]interface{}, matchedNode *models.Node) map[string]string {
	branchIDMapping := make(map[string]string)

	condConfig, ok := common.AsConditionConfig(matchedNode.Config)
	if !ok || condConfig == nil {
		return branchIDMapping
	}

	g.buildOriginalCaseIDToIndexMapping(*condConfig)
	g.extractBranchIDsFromCases(cases, branchIDMapping, *condConfig)

	return branchIDMapping
}

func (g *IFlytekGenerator) buildOriginalCaseIDToIndexMapping(condConfig models.ConditionConfig) map[string]int {
	originalCaseIDToIndex := make(map[string]int)
	for i, caseItem := range condConfig.Cases {
		originalCaseIDToIndex[caseItem.CaseID] = i
	}
	return originalCaseIDToIndex
}

func (g *IFlytekGenerator) extractBranchIDsFromCases(cases []map[string]interface{}, branchIDMapping map[string]string, condConfig models.ConditionConfig) {
	for _, caseItem := range cases {
		branchID, hasID := caseItem["id"].(string)
		level, hasLevel := caseItem["level"].(int)

		if !hasID || !hasLevel {
			continue
		}

		if level == 999 {
			g.handleDefaultBranch(branchIDMapping, branchID)
		} else {
			g.handleNormalBranch(branchIDMapping, branchID, level, condConfig)
		}
	}
}

func (g *IFlytekGenerator) handleDefaultBranch(branchIDMapping map[string]string, branchID string) {
	branchIDMapping["__default__"] = branchID
	branchIDMapping["false"] = branchID
}

func (g *IFlytekGenerator) handleNormalBranch(branchIDMapping map[string]string, branchID string, level int, condConfig models.ConditionConfig) {
	levelKey := fmt.Sprintf("%d", level)
	branchIDMapping[levelKey] = branchID

	if level > 0 && level-1 < len(condConfig.Cases) {
		originalCaseID := condConfig.Cases[level-1].CaseID
		branchIDMapping[originalCaseID] = branchID
	}

	if level == 1 {
		branchIDMapping["true"] = branchID // backward compatibility
	}
}

// generateEdges generates connection relationships
func (g *IFlytekGenerator) generateEdges(edges []models.Edge, iflytekDSL *IFlytekDSL) error {
	for _, edge := range edges {
		// Use mapped node IDs
		sourceID := g.idMapping[edge.Source]
		targetID := g.idMapping[edge.Target]

		// If mapping does not exist, use original ID
		if sourceID == "" {
			sourceID = edge.Source
		}
		if targetID == "" {
			targetID = edge.Target
		}

		// Handle special iteration start edges from Coze
		if edge.PlatformConfig.IFlytek != nil {
			if isIterationStartEdge, ok := edge.PlatformConfig.IFlytek["isIterationStartEdge"].(bool); ok && isIterationStartEdge {
				// This edge should connect from iteration start node to target node
				// sourceID at this point is already the mapped iFlytek iteration node ID
				if sourceID != "" && strings.HasPrefix(sourceID, "iteration::") {
					// Extract UUID and construct iteration start node ID
					uuid := sourceID[len("iteration::"):]
					iterationStartNodeID := "iteration-node-start::" + uuid
					sourceID = iterationStartNodeID
				}
			}
		}

		// Handle default intent target redirection
		finalTargetID := targetID
		sourceHandle := g.convertSourceHandle(edge.SourceHandle, edge.Source)

		// All edges maintain original connections, no redirection

		iflytekEdge := IFlytekEdge{
			Source:       sourceID,
			Target:       finalTargetID,
			SourceHandle: sourceHandle,
			TargetHandle: edge.TargetHandle,
			Type:         g.convertEdgeType(edge.Type),
			ID:           g.generateEdgeIDWithHandle(sourceID, sourceHandle, finalTargetID),
		}

		// Generate default arrow marker
		iflytekEdge.MarkerEnd = &IFlytekMarkerEnd{
			Color: "#275EFF",
			Type:  "arrow",
		}

		// Generate edge data
		iflytekEdge.Data = &IFlytekEdgeData{
			EdgeType: "curve",
		}

		// Restore platform specific configurations
		if edge.PlatformConfig.IFlytek != nil {
			config := edge.PlatformConfig.IFlytek
			// Handle MarkerEnd configuration
			if markerEndData, ok := config["markerEnd"]; ok {
				if markerEndMap, ok := markerEndData.(map[string]interface{}); ok {
					color, _ := markerEndMap["color"].(string)
					markerType, _ := markerEndMap["type"].(string)
					iflytekEdge.MarkerEnd = &IFlytekMarkerEnd{
						Color: color,
						Type:  markerType,
					}
				}
			}
			// Handle EdgeType configuration
			if edgeType, ok := config["edgeType"].(string); ok && edgeType != "" {
				iflytekEdge.Data.EdgeType = edgeType
			}
		}

		iflytekDSL.FlowData.Edges = append(iflytekDSL.FlowData.Edges, iflytekEdge)
	}

	return nil
}

// generateEdgeID generates iFlytek SparkAgent edge ID

// generateEdgeIDWithHandle generates iFlytek SparkAgent edge ID with source handle
func (g *IFlytekGenerator) generateEdgeIDWithHandle(sourceID, sourceHandle, targetID string) string {
	if sourceHandle != "" && sourceHandle != "source" {
		return fmt.Sprintf("reactflow__edge-%s%s-%s", sourceID, sourceHandle, targetID)
	}
	return fmt.Sprintf("reactflow__edge-%s-%s", sourceID, targetID)
}

// convertEdgeType converts connection type
func (g *IFlytekGenerator) convertEdgeType(edgeType models.EdgeType) string {
	switch edgeType {
	case models.EdgeTypeDefault:
		return "customEdge"
	case models.EdgeTypeConditional:
		return "customEdge"
	default:
		return "customEdge"
	}
}

// convertSourceHandle converts source handle, handles special cases for branch nodes and classifier nodes
func (g *IFlytekGenerator) convertSourceHandle(sourceHandle, sourceNodeID string) string {
	mappedSourceID := g.getMappedSourceNodeID(sourceNodeID)

	if convertedHandle := g.handleStartNodeSource(sourceHandle); convertedHandle != "" {
		return convertedHandle
	}

	if convertedHandle := g.handleConditionBranchSource(sourceHandle, mappedSourceID); convertedHandle != "" {
		return convertedHandle
	}

	if convertedHandle := g.handleMultiLevelConditionSource(sourceHandle, mappedSourceID); convertedHandle != "" {
		return convertedHandle
	}

	if convertedHandle := g.handleClassifierNumberSource(sourceHandle, mappedSourceID); convertedHandle != "" {
		return convertedHandle
	}

	if convertedHandle := g.handleClassifierIntentMappingSource(sourceHandle, mappedSourceID); convertedHandle != "" {
		return convertedHandle
	}

	if convertedHandle := g.handleClassifierIntentSource(sourceHandle, mappedSourceID); convertedHandle != "" {
		return convertedHandle
	}

	return sourceHandle
}

// getMappedSourceNodeID gets the mapped source node ID or returns original if not found
func (g *IFlytekGenerator) getMappedSourceNodeID(sourceNodeID string) string {
	mappedSourceID := g.idMapping[sourceNodeID]
	if mappedSourceID == "" {
		return sourceNodeID
	}
	return mappedSourceID
}

// handleStartNodeSource handles start node source conversion
func (g *IFlytekGenerator) handleStartNodeSource(sourceHandle string) string {
	if sourceHandle == "start" {
		// For start node, return "source" as the handle, which is standard for iFlytek SparkAgent
		return "source"
	}
	return ""
}

// handleConditionBranchSource handles condition branch node source conversion
func (g *IFlytekGenerator) handleConditionBranchSource(sourceHandle, mappedSourceID string) string {
	if sourceHandle != "true" && sourceHandle != "false" {
		return ""
	}

	if g.conditionBranchMapping == nil {
		return sourceHandle
	}

	branchMapping, exists := g.conditionBranchMapping[mappedSourceID]
	if !exists {
		return sourceHandle
	}

	if convertedHandle := g.tryDirectBranchMapping(sourceHandle, branchMapping); convertedHandle != "" {
		return convertedHandle
	}

	if branchID, exists := branchMapping.BranchIDs[sourceHandle]; exists {
		return branchID
	}

	return sourceHandle
}

// tryDirectBranchMapping tries direct true/false branch mapping
func (g *IFlytekGenerator) tryDirectBranchMapping(sourceHandle string, branchMapping *BranchMapping) string {
	if sourceHandle == "true" && branchMapping.TrueBranchID != "" {
		return branchMapping.TrueBranchID
	}
	if sourceHandle == "false" && branchMapping.FalseBranchID != "" {
		return branchMapping.FalseBranchID
	}
	return ""
}

// handleMultiLevelConditionSource handles multi-level condition branch node source conversion
func (g *IFlytekGenerator) handleMultiLevelConditionSource(sourceHandle, mappedSourceID string) string {
	if g.conditionBranchMapping == nil {
		return ""
	}

	branchMapping, exists := g.conditionBranchMapping[mappedSourceID]
	if !exists {
		return ""
	}

	// First try direct case ID lookup (handles Dify UUID format case IDs)
	if branchID, exists := branchMapping.BranchIDs[sourceHandle]; exists {
		return branchID
	}

	// For backward compatibility with numeric handles
	if g.isNumericHandle(sourceHandle) {
		return g.lookupNumericHandle(sourceHandle, branchMapping)
	}

	return ""
}

// isNumericHandle checks if the source handle is numeric (like "1", "2", "3", etc.)
func (g *IFlytekGenerator) isNumericHandle(sourceHandle string) bool {
	if len(sourceHandle) == 0 {
		return false
	}

	for _, r := range sourceHandle {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// lookupNumericHandle looks up branch ID by numeric handle
func (g *IFlytekGenerator) lookupNumericHandle(sourceHandle string, branchMapping *BranchMapping) string {
	if branchID, exists := branchMapping.BranchIDs[sourceHandle]; exists {
		return branchID
	}
	return ""
}

// handleClassifierNumberSource handles classifier number handles conversion
func (g *IFlytekGenerator) handleClassifierNumberSource(sourceHandle, mappedSourceID string) string {
	classifierGen, exists := g.classifierGenerators[mappedSourceID]
	if !exists {
		return ""
	}

	classIDToIntentID := classifierGen.GetClassIDToIntentIDMapping()
	if intentID, found := classIDToIntentID[sourceHandle]; found {
		return intentID
	}

	return ""
}

// lookupClassifierIntentID performs unified classifier intent ID lookup with default handling
func (g *IFlytekGenerator) lookupClassifierIntentID(sourceHandle string, classIDToIntentID map[string]string) string {
	// Handle special case: "default" should map to "__default__" key
	if sourceHandle == DefaultSourceHandle {
		if intentID, found := classIDToIntentID[DefaultIntentKey]; found {
			return intentID
		}
	}

	// Try direct mapping
	if intentID, found := classIDToIntentID[sourceHandle]; found {
		return intentID
	}

	return ""
}

// handleClassifierIntentMappingSource handles classifier intent mapping source conversion
func (g *IFlytekGenerator) handleClassifierIntentMappingSource(sourceHandle, mappedSourceID string) string {
	// First try to use classifierGenerators mapping (same as handleClassifierNumberSource)
	if classifierGen, exists := g.classifierGenerators[mappedSourceID]; exists {
		classIDToIntentID := classifierGen.GetClassIDToIntentIDMapping()
		if intentID := g.lookupClassifierIntentID(sourceHandle, classIDToIntentID); intentID != "" {
			return intentID
		}
	}

	// Fallback to old classifierIntentMapping system
	if g.classifierIntentMapping == nil {
		return ""
	}

	classifierMapping, exists := g.classifierIntentMapping[mappedSourceID]
	if !exists {
		return ""
	}

	if convertedHandle := g.tryClassIDToIntentIDMapping(sourceHandle, classifierMapping); convertedHandle != "" {
		return convertedHandle
	}

	return g.tryPositionalIntentMapping(sourceHandle, classifierMapping)
}

// tryClassIDToIntentIDMapping tries ClassIDToIntentID mapping
func (g *IFlytekGenerator) tryClassIDToIntentIDMapping(sourceHandle string, classifierMapping *ClassifierMapping) string {
	if classifierMapping.ClassIDToIntentID == nil {
		return ""
	}

	return g.lookupClassifierIntentID(sourceHandle, classifierMapping.ClassIDToIntentID)
}

// tryPositionalIntentMapping tries positional intent mapping for numbered handles
func (g *IFlytekGenerator) tryPositionalIntentMapping(sourceHandle string, classifierMapping *ClassifierMapping) string {
	switch sourceHandle {
	case "1":
		if len(classifierMapping.IntentIDs) > 0 {
			return classifierMapping.IntentIDs[0]
		}
	case "2":
		if len(classifierMapping.IntentIDs) > 1 {
			return classifierMapping.IntentIDs[1]
		}
	case "3":
		if len(classifierMapping.IntentIDs) > 2 {
			return classifierMapping.IntentIDs[2]
		}
	case "4":
		if len(classifierMapping.IntentIDs) > 3 {
			return classifierMapping.IntentIDs[3]
		}
	}
	return ""
}

// handleClassifierIntentSource handles classifier intent source conversion
func (g *IFlytekGenerator) handleClassifierIntentSource(sourceHandle, mappedSourceID string) string {
	if g.classifierIntentMapping == nil {
		return ""
	}

	classifierMapping, exists := g.classifierIntentMapping[mappedSourceID]
	if !exists {
		return ""
	}

	// For default intent, return default intent ID directly, do not redirect
	if sourceHandle == classifierMapping.DefaultIntentID {
		return sourceHandle
	}

	return ""
}

// ExtractBranchMapping implements BranchMappingExtractor interface
func (g *IFlytekGenerator) ExtractBranchMapping(iflytekNode IFlytekNode) {
	g.extractBranchMapping(iflytekNode)
}

// extractBranchMapping extracts branch mapping from generated condition nodes
func (g *IFlytekGenerator) extractBranchMapping(iflytekNode IFlytekNode) {
	// Extract cases from node parameter
	cases := g.extractCasesFromNodeParam(iflytekNode.Data.NodeParam)
	if cases == nil {
		return
	}

	// Create and populate branch mapping
	mapping := g.createBranchMappingFromCases(cases)

	// Save mapping
	g.conditionBranchMapping[iflytekNode.ID] = mapping
}

// extractCasesFromNodeParam extracts cases from node parameters
func (g *IFlytekGenerator) extractCasesFromNodeParam(nodeParam map[string]interface{}) []interface{} {
	if nodeParam == nil {
		return nil
	}

	casesInterface, exists := nodeParam["cases"]
	if !exists {
		return nil
	}

	return g.convertCasesToInterfaceSlice(casesInterface)
}

// convertCasesToInterfaceSlice converts cases to []interface{} format
func (g *IFlytekGenerator) convertCasesToInterfaceSlice(casesInterface interface{}) []interface{} {
	// Try to convert to []interface{} type
	if casesList, ok := casesInterface.([]interface{}); ok {
		return casesList
	}

	// Try to convert []map[string]interface{} to []interface{}
	if casesMapList, ok := casesInterface.([]map[string]interface{}); ok {
		cases := make([]interface{}, len(casesMapList))
		for i, caseMap := range casesMapList {
			cases[i] = caseMap
		}
		return cases
	}

	return nil
}

// createBranchMappingFromCases creates branch mapping from cases
func (g *IFlytekGenerator) createBranchMappingFromCases(cases []interface{}) *BranchMapping {
	mapping := &BranchMapping{
		BranchIDs: make(map[string]string),
	}

	for _, caseInterface := range cases {
		g.processCaseForBranchMapping(caseInterface, mapping)
	}

	return mapping
}

// processCaseForBranchMapping processes a single case for branch mapping
func (g *IFlytekGenerator) processCaseForBranchMapping(caseInterface interface{}, mapping *BranchMapping) {
	caseMap, ok := caseInterface.(map[string]interface{})
	if !ok {
		return
	}

	level, branchID := g.extractLevelAndBranchID(caseMap)
	if level == 0 || branchID == "" {
		return
	}

	g.storeBranchIDByLevel(mapping, level, branchID)
}

// extractLevelAndBranchID extracts level and branch ID from case map
func (g *IFlytekGenerator) extractLevelAndBranchID(caseMap map[string]interface{}) (int, string) {
	levelInterface, exists := caseMap["level"]
	if !exists {
		return 0, ""
	}

	level, ok := levelInterface.(int)
	if !ok {
		return 0, ""
	}

	idInterface, exists := caseMap["id"]
	if !exists {
		return 0, ""
	}

	branchID, ok := idInterface.(string)
	if !ok {
		return 0, ""
	}

	return level, branchID
}

// storeBranchIDByLevel stores branch ID by level with backward compatibility
func (g *IFlytekGenerator) storeBranchIDByLevel(mapping *BranchMapping, level int, branchID string) {
	switch level {
	case 1:
		g.storeLevel1BranchID(mapping, branchID)
	case 999:
		g.storeDefaultBranchID(mapping, branchID)
	default:
		g.storeMultiLevelBranchID(mapping, level, branchID)
	}
}

// storeLevel1BranchID stores level 1 (true) branch ID
func (g *IFlytekGenerator) storeLevel1BranchID(mapping *BranchMapping, branchID string) {
	mapping.TrueBranchID = branchID // backward compatibility
	mapping.BranchIDs["true"] = branchID
	mapping.BranchIDs["1"] = branchID // also map to string "1"
}

// storeDefaultBranchID stores default (false) branch ID
func (g *IFlytekGenerator) storeDefaultBranchID(mapping *BranchMapping, branchID string) {
	mapping.FalseBranchID = branchID // backward compatibility
	mapping.BranchIDs["false"] = branchID
	mapping.BranchIDs["__default__"] = branchID
}

// storeMultiLevelBranchID stores multi-level branch ID
func (g *IFlytekGenerator) storeMultiLevelBranchID(mapping *BranchMapping, level int, branchID string) {
	levelKey := fmt.Sprintf("%d", level)
	mapping.BranchIDs[levelKey] = branchID

	// Store specific levels for backward compatibility
	g.storeSpecificLevelMapping(mapping, level, branchID)
}

// storeSpecificLevelMapping stores specific level mappings for backward compatibility
func (g *IFlytekGenerator) storeSpecificLevelMapping(mapping *BranchMapping, level int, branchID string) {
	switch level {
	case 2:
		mapping.BranchIDs["2"] = branchID
	case 3:
		mapping.BranchIDs["3"] = branchID
	case 4:
		mapping.BranchIDs["4"] = branchID
	}
}

// extractBranchMappingWithCaseIDs extracts branch mapping from generated condition nodes and preserves case ID mappings
func (g *IFlytekGenerator) extractBranchMappingWithCaseIDs(iflytekNode IFlytekNode, additionalMappings map[string]string) {
	cases := g.extractBranchCasesFromNode(iflytekNode)
	if cases == nil {
		return
	}

	mapping := g.createBranchMapping(additionalMappings)
	g.processCasesForMapping(cases, mapping)
	g.conditionBranchMapping[iflytekNode.ID] = mapping
}

func (g *IFlytekGenerator) extractBranchCasesFromNode(iflytekNode IFlytekNode) []interface{} {
	nodeParam := iflytekNode.Data.NodeParam
	if nodeParam == nil {
		return nil
	}

	casesInterface, exists := nodeParam["cases"]
	if !exists {
		return nil
	}

	return g.convertToCasesInterface(casesInterface)
}

func (g *IFlytekGenerator) convertToCasesInterface(casesInterface interface{}) []interface{} {
	if casesList, ok := casesInterface.([]interface{}); ok {
		return casesList
	}

	if casesMapList, ok := casesInterface.([]map[string]interface{}); ok {
		cases := make([]interface{}, len(casesMapList))
		for i, caseMap := range casesMapList {
			cases[i] = caseMap
		}
		return cases
	}

	return nil
}

func (g *IFlytekGenerator) createBranchMapping(additionalMappings map[string]string) *BranchMapping {
	mapping := &BranchMapping{
		BranchIDs: make(map[string]string),
	}

	if additionalMappings != nil {
		for caseID, branchID := range additionalMappings {
			mapping.BranchIDs[caseID] = branchID
		}
	}

	return mapping
}

func (g *IFlytekGenerator) processCasesForMapping(cases []interface{}, mapping *BranchMapping) {
	for _, caseInterface := range cases {
		caseData := g.extractCaseData(caseInterface)
		if caseData == nil {
			continue
		}

		g.storeBranchMapping(caseData.Level, caseData.BranchID, mapping)
	}
}

type CaseData struct {
	Level    int
	BranchID string
}

func (g *IFlytekGenerator) extractCaseData(caseInterface interface{}) *CaseData {
	caseMap, ok := caseInterface.(map[string]interface{})
	if !ok {
		return nil
	}

	levelInterface, exists := caseMap["level"]
	if !exists {
		return nil
	}

	level, ok := levelInterface.(int)
	if !ok {
		return nil
	}

	idInterface, exists := caseMap["id"]
	if !exists {
		return nil
	}

	branchID, ok := idInterface.(string)
	if !ok {
		return nil
	}

	return &CaseData{Level: level, BranchID: branchID}
}

func (g *IFlytekGenerator) storeBranchMapping(level int, branchID string, mapping *BranchMapping) {
	switch level {
	case 1:
		g.storeLevelOneBranch(branchID, mapping)
	case 999:
		g.storeDefaultBranch(branchID, mapping)
	default:
		g.storeMultiLevelBranch(level, branchID, mapping)
	}
}

func (g *IFlytekGenerator) storeLevelOneBranch(branchID string, mapping *BranchMapping) {
	mapping.TrueBranchID = branchID
	mapping.BranchIDs["true"] = branchID
	mapping.BranchIDs["1"] = branchID
}

func (g *IFlytekGenerator) storeDefaultBranch(branchID string, mapping *BranchMapping) {
	mapping.FalseBranchID = branchID
	mapping.BranchIDs["false"] = branchID
	mapping.BranchIDs["__default__"] = branchID
}

func (g *IFlytekGenerator) storeMultiLevelBranch(level int, branchID string, mapping *BranchMapping) {
	levelKey := fmt.Sprintf("%d", level)
	mapping.BranchIDs[levelKey] = branchID

	// Backward compatibility for common levels
	if level >= 2 && level <= 4 {
		mapping.BranchIDs[levelKey] = branchID
	}
}

// analyzeClassifierTargets analyzes classifier target node mapping
func (g *IFlytekGenerator) analyzeClassifierTargets(edges []models.Edge) {
	classifierEdges := g.groupEdgesByClassifier(edges)
	g.updateClassifierMappings(classifierEdges)
}

// groupEdgesByClassifier groups edges by classifier node
func (g *IFlytekGenerator) groupEdgesByClassifier(edges []models.Edge) map[string][]models.Edge {
	classifierEdges := make(map[string][]models.Edge)

	for _, edge := range edges {
		if classifierID := g.findClassifierID(edge.Source); classifierID != "" {
			classifierEdges[classifierID] = append(classifierEdges[classifierID], edge)
		}
	}

	return classifierEdges
}

// findClassifierID finds classifier ID from edge source
func (g *IFlytekGenerator) findClassifierID(edgeSource string) string {
	for originalID, mappedID := range g.idMapping {
		if edgeSource == originalID && g.isClassifierNodeID(mappedID) {
			return mappedID
		}
	}
	return ""
}

// isClassifierNodeID checks if the ID belongs to a classifier node
func (g *IFlytekGenerator) isClassifierNodeID(mappedID string) bool {
	const classifierPrefix = "decision-making::"
	return mappedID != "" && len(mappedID) > len(classifierPrefix) &&
		mappedID[:len(classifierPrefix)] == classifierPrefix
}

// updateClassifierMappings updates classifier mappings
func (g *IFlytekGenerator) updateClassifierMappings(classifierEdges map[string][]models.Edge) {
	for classifierID, edges := range classifierEdges {
		if len(edges) >= 2 { // At least 2 edges needed
			mapping := g.getOrCreateClassifierMapping(classifierID)
			g.setFirstIntentTarget(mapping, edges)
		}
	}
}

// getOrCreateClassifierMapping gets existing classifier mapping or creates one
func (g *IFlytekGenerator) getOrCreateClassifierMapping(classifierID string) *ClassifierMapping {
	if existing, exists := g.classifierIntentMapping[classifierID]; exists {
		return existing
	}

	mapping := &ClassifierMapping{
		IntentIDs: make([]string, 0),
	}
	g.classifierIntentMapping[classifierID] = mapping
	return mapping
}

// setFirstIntentTarget sets the first intent target for mapping
func (g *IFlytekGenerator) setFirstIntentTarget(mapping *ClassifierMapping, edges []models.Edge) {
	if len(edges) == 0 {
		return
	}

	firstTarget := edges[0].Target
	if mappedTarget := g.idMapping[firstTarget]; mappedTarget != "" {
		mapping.FirstIntentTarget = mappedTarget
	} else {
		mapping.FirstIntentTarget = firstTarget
	}
}

// extractClassifierMapping extracts intent mapping from generated classifier nodes
func (g *IFlytekGenerator) extractClassifierMapping(iflytekNode IFlytekNode) {
	intentChains := g.getIntentChainsFromNode(iflytekNode)
	if intentChains == nil {
		return
	}

	mapping := g.getOrCreateClassifierMapping(iflytekNode.ID)
	if g.isMappingComplete(mapping) {
		return
	}

	g.processIntentChains(mapping, intentChains)
}

// getIntentChainsFromNode extracts intent chains from node parameters
func (g *IFlytekGenerator) getIntentChainsFromNode(iflytekNode IFlytekNode) []map[string]interface{} {
	if iflytekNode.Data.NodeParam == nil {
		return nil
	}

	intentChainsInterface, exists := iflytekNode.Data.NodeParam["intentChains"]
	if !exists {
		return nil
	}

	intentChains, ok := intentChainsInterface.([]map[string]interface{})
	if !ok {
		return nil
	}

	return intentChains
}

// isMappingComplete checks if mapping is already complete
func (g *IFlytekGenerator) isMappingComplete(mapping *ClassifierMapping) bool {
	return mapping.DefaultIntentID != "" && len(mapping.IntentIDs) > 0
}

// processIntentChains processes intent chains and updates mapping
func (g *IFlytekGenerator) processIntentChains(mapping *ClassifierMapping, intentChains []map[string]interface{}) {
	for _, intentChain := range intentChains {
		g.processIntentChain(mapping, intentChain)
	}
}

// processIntentChain processes a single intent chain
func (g *IFlytekGenerator) processIntentChain(mapping *ClassifierMapping, intentChain map[string]interface{}) {
	intentType := g.extractIntentType(intentChain)
	intentID := g.extractIntentID(intentChain)

	if intentID == "" {
		return
	}

	g.updateMappingWithIntent(mapping, intentType, intentID)
}

// extractIntentType extracts intent type from intent chain
func (g *IFlytekGenerator) extractIntentType(intentChain map[string]interface{}) int {
	intentType, exists := intentChain["intentType"]
	if !exists {
		return 0
	}

	intentTypeInt, ok := intentType.(int)
	if !ok {
		return 0
	}

	return intentTypeInt
}

// extractIntentID extracts intent ID from intent chain
func (g *IFlytekGenerator) extractIntentID(intentChain map[string]interface{}) string {
	intentID, exists := intentChain["id"].(string)
	if !exists {
		return ""
	}
	return intentID
}

// updateMappingWithIntent updates mapping based on intent type
func (g *IFlytekGenerator) updateMappingWithIntent(mapping *ClassifierMapping, intentType int, intentID string) {
	switch intentType {
	case 2: // Normal classification intent
		mapping.IntentIDs = append(mapping.IntentIDs, intentID)
	case 1: // Default intent
		mapping.DefaultIntentID = intentID
	}
}

// isDefaultIntentEdge checks if it's a default intent edge
func (g *IFlytekGenerator) isDefaultIntentEdge(sourceID, sourceHandle string) bool {
	// Check if the source node is a classifier
	if !strings.HasPrefix(sourceID, "decision-making::") {
		return false
	}

	// Check if default intent ID exists in mapping and sourceHandle matches
	if mapping, exists := g.classifierIntentMapping[sourceID]; exists {
		return sourceHandle == mapping.DefaultIntentID
	}

	return false
}

// getFirstIntentTarget gets the target node of the first intent of a classifier
func (g *IFlytekGenerator) getFirstIntentTarget(classifierID string) string {
	if mapping, exists := g.classifierIntentMapping[classifierID]; exists {
		return mapping.FirstIntentTarget
	}
	return ""
}

// generateDefaultIntentEdges generates edges for default intents of classifiers, connecting to the target node of the last intent
func (g *IFlytekGenerator) generateDefaultIntentEdges(edges []models.Edge, iflytekDSL *IFlytekDSL) {
	// Generate connection edges for default intents of each classifier node
	for classifierID, classifierGen := range g.classifierGenerators {
		classIDToIntentID := classifierGen.GetClassIDToIntentIDMapping()

		// Get default intent ID
		defaultIntentID, hasDefault := classIDToIntentID["__default__"]
		if !hasDefault {
			continue
		}

		// Find the target node of the last classification connection
		var lastClassTargetNode string

		for _, edge := range edges {
			if g.idMapping[edge.Source] == classifierID {
				// Keep updating to the last matching edge's target
				if mapped := g.idMapping[edge.Target]; mapped != "" {
					lastClassTargetNode = mapped
				} else {
					lastClassTargetNode = edge.Target
				}
			}
		}

		if lastClassTargetNode == "" {
			continue
		}

		// Create edge for default intent
		defaultEdge := IFlytekEdge{
			ID:           g.generateEdgeIDWithHandle(classifierID, defaultIntentID, lastClassTargetNode),
			Source:       classifierID,
			Target:       lastClassTargetNode,
			SourceHandle: defaultIntentID,
			TargetHandle: "",
			Type:         "customEdge",
			MarkerEnd: &IFlytekMarkerEnd{
				Color: "#275EFF",
				Type:  "arrow",
			},
			Data: &IFlytekEdgeData{
				EdgeType: "curve",
			},
		}

		iflytekDSL.FlowData.Edges = append(iflytekDSL.FlowData.Edges, defaultEdge)
	}
}

// findIterationSubNodes finds sub-nodes belonging to a specified iteration node
func (g *IFlytekGenerator) findIterationSubNodes(allNodes []models.Node, iterationID string) []models.Node {
	var subNodes []models.Node

	for _, node := range allNodes {
		if g.isNodeInIteration(node, iterationID) {
			subNodes = append(subNodes, node)
		}
	}

	return subNodes
}

// isNodeInIteration checks if a node belongs to a specific iteration
func (g *IFlytekGenerator) isNodeInIteration(node models.Node, iterationID string) bool {
	if node.Config == nil {
		return false
	}

	return g.checkIterationMembership(node.Config, iterationID)
}

// checkIterationMembership checks iteration membership based on config type
func (g *IFlytekGenerator) checkIterationMembership(config interface{}, iterationID string) bool {
	switch cfg := config.(type) {
	case models.StartConfig:
		return g.isStartNodeInIteration(cfg, iterationID)
	case models.CodeConfig:
		return g.isConfigInIteration(cfg.IsInIteration, cfg.IterationID, iterationID)
	case models.LLMConfig:
		return g.isConfigInIteration(cfg.IsInIteration, cfg.IterationID, iterationID)
	case models.ConditionConfig:
		return g.isConfigInIteration(cfg.IsInIteration, cfg.IterationID, iterationID)
	case models.ClassifierConfig:
		return g.isConfigInIteration(cfg.IsInIteration, cfg.IterationID, iterationID)
	}
	return false
}

// isStartNodeInIteration checks if start node belongs to iteration
func (g *IFlytekGenerator) isStartNodeInIteration(config models.StartConfig, iterationID string) bool {
	return config.IsInIteration && config.ParentID == iterationID
}

// isConfigInIteration checks if config indicates iteration membership
func (g *IFlytekGenerator) isConfigInIteration(isInIteration bool, configIterationID, targetIterationID string) bool {
	return isInIteration && configIterationID == targetIterationID
}

// processIterationSubNodes processes iteration sub-nodes, sets the correct parentId
func (g *IFlytekGenerator) processIterationSubNodes(nodes []models.Node, iflytekDSL *IFlytekDSL) error {
	iterationMap := g.buildIterationMap()
	processedIterations := make(map[string]bool)

	if err := g.generateIterationSubNodesForEach(nodes, iflytekDSL, iterationMap, processedIterations); err != nil {
		return err
	}

	g.removeDuplicateIterationSubNodes(iflytekDSL)
	g.setParentIDsForSubNodes(nodes, iflytekDSL, iterationMap)

	return nil
}

// buildIterationMap builds a map of Dify ID to iFlytek ID for iteration nodes
func (g *IFlytekGenerator) buildIterationMap() map[string]string {
	iterationMap := make(map[string]string)

	for difyID, iflytekID := range g.idMapping {
		if g.isIterationNodeID(iflytekID) {
			iterationMap[difyID] = iflytekID
		}
	}

	return iterationMap
}

// isIterationNodeID checks if the ID belongs to an iteration node
func (g *IFlytekGenerator) isIterationNodeID(iflytekID string) bool {
	return len(iflytekID) > len("iteration::") && iflytekID[:len("iteration::")] == "iteration::"
}

// generateIterationSubNodesForEach generates sub-nodes for each iteration
func (g *IFlytekGenerator) generateIterationSubNodesForEach(nodes []models.Node, iflytekDSL *IFlytekDSL, iterationMap map[string]string, processedIterations map[string]bool) error {
	for difyID, iflytekID := range iterationMap {
		if processedIterations[iflytekID] {
			continue
		}

		if g.hasExistingSubNodes(iflytekDSL, iflytekID) {
			processedIterations[iflytekID] = true
			continue
		}

		if err := g.processSignleIterationNode(nodes, iflytekDSL, difyID, iflytekID, processedIterations); err != nil {
			return err
		}
	}

	return nil
}

// hasExistingSubNodes checks if iteration already has sub-nodes
func (g *IFlytekGenerator) hasExistingSubNodes(iflytekDSL *IFlytekDSL, iflytekID string) bool {
	for _, existingNode := range iflytekDSL.FlowData.Nodes {
		if existingNode.ParentID != nil && *existingNode.ParentID == iflytekID {
			return true
		}
	}
	return false
}

// processSignleIterationNode processes a single iteration node
func (g *IFlytekGenerator) processSignleIterationNode(nodes []models.Node, iflytekDSL *IFlytekDSL, difyID, iflytekID string, processedIterations map[string]bool) error {
	originalIterationNode := g.findOriginalIterationNode(nodes, difyID)
	if originalIterationNode.ID == "" {
		return nil
	}

	iterationGen := g.setupIterationGenerator()
	if iterationGen == nil {
		return nil
	}

	if err := g.generateAndAddSubNodes(originalIterationNode, iflytekID, nodes, difyID, iflytekDSL, iterationGen); err != nil {
		return err
	}

	processedIterations[iflytekID] = true
	return nil
}

// findOriginalIterationNode finds the original iteration node
func (g *IFlytekGenerator) findOriginalIterationNode(nodes []models.Node, difyID string) models.Node {
	for _, node := range nodes {
		if node.ID == difyID && node.Type == models.NodeTypeIteration {
			return node
		}
	}
	return models.Node{}
}

// setupIterationGenerator sets up iteration generator with required mappings
func (g *IFlytekGenerator) setupIterationGenerator() *IterationNodeGenerator {
	iterationGenerator, err := g.factory.GetGenerator(models.NodeTypeIteration)
	if err != nil {
		return nil
	}

	iterationGen, ok := iterationGenerator.(*IterationNodeGenerator)
	if !ok {
		return nil
	}

	iterationGen.SetIDMapping(g.idMapping)
	iterationGen.SetNodeTitleMapping(g.nodeTitleMapping)
	iterationGen.SetBranchExtractor(g)

	return iterationGen
}

// generateAndAddSubNodes generates and adds sub-nodes to DSL
func (g *IFlytekGenerator) generateAndAddSubNodes(originalIterationNode models.Node, iflytekID string, nodes []models.Node, difyID string, iflytekDSL *IFlytekDSL, iterationGen *IterationNodeGenerator) error {
	nodeIDs := g.generateIterationNodeIDs(iflytekID)
	iterationSubNodes := g.findIterationSubNodes(nodes, difyID)

	generatedSubNodes, generatedIterationEdges, err := iterationGen.GenerateIterationSubNodesWithIDs(
		originalIterationNode, iflytekID, iterationSubNodes, nodeIDs.startID, nodeIDs.endID, nodeIDs.codeID)
	if err != nil {
		return fmt.Errorf("failed to generate iteration sub-nodes: %w", err)
	}

	g.addGeneratedNodesToDSL(iflytekDSL, generatedSubNodes, generatedIterationEdges)
	g.updateIterationNodeParam(iflytekDSL, iflytekID, nodeIDs.startID)

	internalEdges := g.generateIterationInternalEdges(generatedSubNodes, iflytekID, originalIterationNode)
	iflytekDSL.FlowData.Edges = append(iflytekDSL.FlowData.Edges, internalEdges...)

	return nil
}

// iterationNodeIDs holds generated node IDs for iteration
type iterationNodeIDs struct {
	startID string
	endID   string
	codeID  string
}

// generateIterationNodeIDs generates all required node IDs for iteration
func (g *IFlytekGenerator) generateIterationNodeIDs(iflytekID string) iterationNodeIDs {
	return iterationNodeIDs{
		startID: g.generateDeterministicStartNodeID(iflytekID),
		endID:   g.generateDeterministicEndNodeID(iflytekID),
		codeID:  g.generateDeterministicCodeNodeID(iflytekID),
	}
}

// addGeneratedNodesToDSL adds generated nodes and edges to DSL
func (g *IFlytekGenerator) addGeneratedNodesToDSL(iflytekDSL *IFlytekDSL, generatedSubNodes []IFlytekNode, generatedIterationEdges []IFlytekEdge) {
	iflytekDSL.FlowData.Nodes = append(iflytekDSL.FlowData.Nodes, generatedSubNodes...)

	for _, iterationEdge := range generatedIterationEdges {
		iflytekDSL.FlowData.Edges = append(iflytekDSL.FlowData.Edges, iterationEdge)
	}
}

// setParentIDsForSubNodes sets parent IDs for iteration sub-nodes
func (g *IFlytekGenerator) setParentIDsForSubNodes(nodes []models.Node, iflytekDSL *IFlytekDSL, iterationMap map[string]string) {
	for i, node := range iflytekDSL.FlowData.Nodes {
		if node.ParentID != nil {
			continue
		}

		parentIterationID := g.findParentIterationForNode(node, nodes, iterationMap)
		if parentIterationID != "" {
			g.updateNodeWithParentID(&iflytekDSL.FlowData.Nodes[i], parentIterationID)
		}
	}
}

// findParentIterationForNode finds the parent iteration ID for a node
func (g *IFlytekGenerator) findParentIterationForNode(node IFlytekNode, nodes []models.Node, iterationMap map[string]string) string {
	for _, originalNode := range nodes {
		if g.idMapping[originalNode.ID] == node.ID && g.isIterationSubNode(originalNode) {
			return g.getParentIterationID(originalNode, iterationMap)
		}
	}
	return ""
}

// updateNodeWithParentID updates node with parent ID and related properties
func (g *IFlytekGenerator) updateNodeWithParentID(node *IFlytekNode, parentIterationID string) {
	node.ParentID = &parentIterationID
	node.Extent = "parent"
	node.ZIndex = 1

	draggableFalse := false
	node.Draggable = &draggableFalse
	node.Data.ParentID = &parentIterationID
	node.Data.OriginPosition = &node.PositionAbsolute
}

// removeDuplicateIterationSubNodes removes duplicate iteration sub-nodes
func (g *IFlytekGenerator) removeDuplicateIterationSubNodes(iflytekDSL *IFlytekDSL) {
	// Record seen node types and parent node combinations to avoid duplicates
	seenSubNodeTypes := make(map[string]map[string]bool) // parentID -> nodeType -> bool
	var uniqueNodes []IFlytekNode
	removedNodeIDs := make(map[string]bool)

	for _, node := range iflytekDSL.FlowData.Nodes {
		// If it's an iteration sub-node (with parentId), check for duplicates
		if node.ParentID != nil {
			parentID := *node.ParentID
			nodeType := node.Type

			// Initialize record for parent node
			if seenSubNodeTypes[parentID] == nil {
				seenSubNodeTypes[parentID] = make(map[string]bool)
			}

			// Check if a node of the same type already exists
			// Note: Iteration sub-nodes (start node, code node, end node) are required, should not be removed
			if !seenSubNodeTypes[parentID][nodeType] {
				seenSubNodeTypes[parentID][nodeType] = true
				uniqueNodes = append(uniqueNodes, node)
			} else {
				// For iteration sub-nodes, we should not remove duplicates, as each iteration needs a complete set of sub-nodes
				// Only remove if it's truly a duplicate (e.g., a node is added multiple times)
				if g.isTrulyDuplicateIterationSubNode(node, iflytekDSL.FlowData.Nodes) {
					removedNodeIDs[node.ID] = true
				} else {
					uniqueNodes = append(uniqueNodes, node)
				}
			}
		} else {
			// Non-iteration sub-nodes are added directly
			uniqueNodes = append(uniqueNodes, node)
		}
	}

	// Update node list
	iflytekDSL.FlowData.Nodes = uniqueNodes

	// Also need to update edges, remove edges pointing to deleted nodes
	var uniqueEdges []IFlytekEdge
	for _, edge := range iflytekDSL.FlowData.Edges {
		if !removedNodeIDs[edge.Source] && !removedNodeIDs[edge.Target] {
			uniqueEdges = append(uniqueEdges, edge)
		}
	}

	iflytekDSL.FlowData.Edges = uniqueEdges
}

// isTrulyDuplicateIterationSubNode checks if it's a truly duplicate iteration sub-node
func (g *IFlytekGenerator) isTrulyDuplicateIterationSubNode(node IFlytekNode, allNodes []IFlytekNode) bool {
	if node.ParentID == nil {
		return false
	}

	parentID := *node.ParentID
	nodeType := node.Type
	nodeID := node.ID

	// Calculate the number of nodes of the same type under the same parent node
	count := 0
	for _, otherNode := range allNodes {
		if otherNode.ParentID != nil && *otherNode.ParentID == parentID && otherNode.Type == nodeType {
			count++
			// If multiple nodes of the same type are found, check if it's truly a duplicate
			if count > 1 {
				// For iteration sub-nodes, we allow multiple nodes of each type
				// Only consider it a duplicate if the ID is exactly the same
				if otherNode.ID == nodeID {
					return true
				}
			}
		}
	}

	return false
}

// generateIterationInternalEdges generates internal edges for iteration
func (g *IFlytekGenerator) generateIterationInternalEdges(subNodes []IFlytekNode, iterationID string, originalIterationNode models.Node) []IFlytekEdge {
	iterationConfig, ok := g.parseIterationConfig(originalIterationNode)
	if !ok {
		return []IFlytekEdge{}
	}

	startNode, endNode, sourceNode := g.findIterationNodes(subNodes, iterationConfig.OutputSelector.NodeID)

	var edges []IFlytekEdge
	if startToSourceEdge := g.createStartToSourceEdge(startNode, sourceNode, iterationID); startToSourceEdge != nil {
		edges = append(edges, *startToSourceEdge)
	}

	if sourceToEndEdge := g.createSourceToEndEdge(sourceNode, endNode); sourceToEndEdge != nil {
		edges = append(edges, *sourceToEndEdge)
	}

	return edges
}

// parseIterationConfig parses iteration configuration
func (g *IFlytekGenerator) parseIterationConfig(originalIterationNode models.Node) (models.IterationConfig, bool) {
	iterationConfig, ok := common.AsIterationConfig(originalIterationNode.Config)
	return *iterationConfig, ok && iterationConfig != nil
}

// findIterationNodes finds start, end, and source nodes in iteration
func (g *IFlytekGenerator) findIterationNodes(subNodes []IFlytekNode, outputSourceNodeID string) (*IFlytekNode, *IFlytekNode, *IFlytekNode) {
	var startNode, endNode, sourceNode *IFlytekNode

	// First pass: find start and end nodes, and try to find source node by mapping
	for i := range subNodes {
		if g.isIterationStartNodeByID(&subNodes[i]) {
			startNode = &subNodes[i]
		} else if g.isIterationEndNode(&subNodes[i]) {
			endNode = &subNodes[i]
		} else if sourceNode == nil {
			sourceNode = g.tryFindSourceNode(&subNodes[i], outputSourceNodeID)
		}
	}

	// Second pass: if source node not found, try fallback strategies
	if sourceNode == nil {
		sourceNode = g.findSourceNodeFallback(subNodes, outputSourceNodeID)
	}

	return startNode, endNode, sourceNode
}

// isIterationStartNodeByID checks if node is iteration start node by ID
func (g *IFlytekGenerator) isIterationStartNodeByID(node *IFlytekNode) bool {
	return strings.HasPrefix(node.ID, "iteration-node-start::")
}

// isIterationEndNode checks if node is iteration end node
func (g *IFlytekGenerator) isIterationEndNode(node *IFlytekNode) bool {
	return strings.HasPrefix(node.ID, "iteration-node-end::")
}

// tryFindSourceNode tries to find source node by output selector mapping
func (g *IFlytekGenerator) tryFindSourceNode(node *IFlytekNode, outputSourceNodeID string) *IFlytekNode {
	if outputSourceNodeID != "" && g.idMapping != nil {
		for originalNodeID, mappedNodeID := range g.idMapping {
			if originalNodeID == outputSourceNodeID && mappedNodeID == node.ID {
				return node
			}
		}
	}

	// Fallback to code node for backward compatibility
	if strings.HasPrefix(node.ID, "ifly-code::") {
		return node
	}

	return nil
}

// findSourceNodeFallback finds source node using fallback strategies
func (g *IFlytekGenerator) findSourceNodeFallback(subNodes []IFlytekNode, outputSourceNodeID string) *IFlytekNode {
	if outputSourceNodeID == "" {
		return nil
	}

	for i := range subNodes {
		if strings.Contains(subNodes[i].ID, outputSourceNodeID) {
			return &subNodes[i]
		}
	}

	return nil
}

// createStartToSourceEdge creates edge from start node to source node
func (g *IFlytekGenerator) createStartToSourceEdge(startNode, sourceNode *IFlytekNode, iterationID string) *IFlytekEdge {
	if startNode == nil || sourceNode == nil {
		return nil
	}

	sourceID := g.fixStartNodeSourceID(startNode.ID, iterationID)

	return &IFlytekEdge{
		Source:       sourceID,
		Target:       sourceNode.ID,
		SourceHandle: "source",
		TargetHandle: "target",
		Type:         "customEdge",
		ID:           g.generateEdgeIDWithHandle(sourceID, "source", sourceNode.ID),
		MarkerEnd:    g.createArrowMarkerEnd(),
		Data:         g.createCurveEdgeData(),
		ZIndex:       1,
	}
}

// createSourceToEndEdge creates edge from source node to end node
func (g *IFlytekGenerator) createSourceToEndEdge(sourceNode, endNode *IFlytekNode) *IFlytekEdge {
	if sourceNode == nil || endNode == nil {
		return nil
	}

	return &IFlytekEdge{
		Source:       sourceNode.ID,
		Target:       endNode.ID,
		SourceHandle: "source",
		TargetHandle: endNode.ID,
		Type:         "customEdge",
		ID:           g.generateEdgeIDWithHandle(sourceNode.ID, "source", endNode.ID),
		MarkerEnd:    g.createArrowMarkerEnd(),
		Data:         g.createCurveEdgeData(),
		ZIndex:       1,
	}
}

// fixStartNodeSourceID fixes abnormal start node source ID
func (g *IFlytekGenerator) fixStartNodeSourceID(sourceID, iterationID string) string {
	if strings.Contains(sourceID, "start") && !strings.HasPrefix(sourceID, "iteration-node-start::") {
		if g.iterationSubNodeMapping[iterationID] != nil {
			if correctStartID, exists := g.iterationSubNodeMapping[iterationID]["start"]; exists {
				return correctStartID
			}
		}
	}
	return sourceID
}

// createArrowMarkerEnd creates arrow marker end for edges
func (g *IFlytekGenerator) createArrowMarkerEnd() *IFlytekMarkerEnd {
	return &IFlytekMarkerEnd{
		Color: "#275EFF",
		Type:  "arrow",
	}
}

// createCurveEdgeData creates curve edge data
func (g *IFlytekGenerator) createCurveEdgeData() *IFlytekEdgeData {
	return &IFlytekEdgeData{
		EdgeType: "curve",
	}
}

// generateDeterministicStartNodeID generates a deterministic start node ID based on iteration node ID
func (g *IFlytekGenerator) generateDeterministicStartNodeID(iterationID string) string {
	// Generate a deterministic start node ID based on iteration node ID, consistent with IterationNodeGenerator
	// Extract UUID part from iteration node ID
	var startNodeID string
	if len(iterationID) > len("iteration::") {
		uuid := iterationID[len("iteration::"):]
		startNodeID = "iteration-node-start::" + uuid
	} else {
		// If format is incorrect, fallback to random generation
		newUUID := generateRandomUUID()
		startNodeID = "iteration-node-start::" + newUUID
	}

	// Add the newly generated ID to the mapping to ensure subsequent references can find the correct ID
	// Here we create an inverse mapping from iteration start node ID to iteration main node ID
	// This way, when other nodes need to reference the iteration start node, they can find the correct ID through this mapping
	if g.iterationSubNodeMapping == nil {
		g.iterationSubNodeMapping = make(map[string]map[string]string)
	}
	if g.iterationSubNodeMapping[iterationID] == nil {
		g.iterationSubNodeMapping[iterationID] = make(map[string]string)
	}
	g.iterationSubNodeMapping[iterationID]["start"] = startNodeID

	return startNodeID
}

// generateDeterministicEndNodeID generates a deterministic end node ID based on iteration node ID
func (g *IFlytekGenerator) generateDeterministicEndNodeID(iterationID string) string {
	// Generate a deterministic end node ID based on iteration node ID, consistent with IterationNodeGenerator
	// Extract UUID part from iteration node ID
	var endNodeID string
	if len(iterationID) > len("iteration::") {
		uuid := iterationID[len("iteration::"):]
		endNodeID = "iteration-node-end::" + uuid
	} else {
		// If not parsable, generate a UUID
		endNodeID = "iteration-node-end::" + generateRandomUUID()
	}

	// Add the newly generated ID to the mapping
	if g.iterationSubNodeMapping[iterationID] == nil {
		g.iterationSubNodeMapping[iterationID] = make(map[string]string)
	}
	g.iterationSubNodeMapping[iterationID]["end"] = endNodeID

	return endNodeID
}

// generateDeterministicCodeNodeID generates a unique ID for iteration code nodes
func (g *IFlytekGenerator) generateDeterministicCodeNodeID(iterationID string) string {
	// Generate a UUID for iteration code nodes, ensuring it is different from other node IDs
	newUUID := generateRandomUUID()
	codeNodeID := "ifly-code::" + newUUID

	// Add the newly generated ID to the mapping
	if g.iterationSubNodeMapping[iterationID] == nil {
		g.iterationSubNodeMapping[iterationID] = nil
	}
	g.iterationSubNodeMapping[iterationID]["code"] = codeNodeID

	return codeNodeID
}

// updateIterationNodeParam updates IterationStartNodeId in iterationNode.nodeParam
func (g *IFlytekGenerator) updateIterationNodeParam(iflytekDSL *IFlytekDSL, iterationID string, startNodeID string) {
	actualStartNodeID := g.resolveActualStartNodeID(iflytekDSL, iterationID, startNodeID)
	g.setIterationStartNodeParam(iflytekDSL, iterationID, actualStartNodeID)
}

// resolveActualStartNodeID resolves the actual start node ID using multiple strategies
func (g *IFlytekGenerator) resolveActualStartNodeID(iflytekDSL *IFlytekDSL, iterationID string, startNodeID string) string {
	if startNodeID != "" {
		return startNodeID
	}

	// Try to find from DSL nodes
	if foundID := g.findStartNodeFromDSL(iflytekDSL, iterationID); foundID != "" {
		return foundID
	}

	// Try to get from mapping
	return g.getStartNodeFromMapping(iterationID)
}

// findStartNodeFromDSL finds iteration start node from DSL nodes
func (g *IFlytekGenerator) findStartNodeFromDSL(iflytekDSL *IFlytekDSL, iterationID string) string {
	for _, node := range iflytekDSL.FlowData.Nodes {
		if g.isIterationStartNode(node, iterationID) {
			return node.ID
		}
	}
	return ""
}

// isIterationStartNode checks if a node is iteration start node for given iteration
func (g *IFlytekGenerator) isIterationStartNode(node IFlytekNode, iterationID string) bool {
	return node.ParentID != nil &&
		*node.ParentID == iterationID &&
		node.Type == "开始节点"
}

// getStartNodeFromMapping gets start node ID from iteration sub-node mapping
func (g *IFlytekGenerator) getStartNodeFromMapping(iterationID string) string {
	if g.iterationSubNodeMapping == nil || g.iterationSubNodeMapping[iterationID] == nil {
		return ""
	}

	if mappedStartID, exists := g.iterationSubNodeMapping[iterationID]["start"]; exists {
		return mappedStartID
	}

	return ""
}

// setIterationStartNodeParam sets IterationStartNodeId in iteration node param
func (g *IFlytekGenerator) setIterationStartNodeParam(iflytekDSL *IFlytekDSL, iterationID string, startNodeID string) {
	for _, node := range iflytekDSL.FlowData.Nodes {
		if node.ID == iterationID && node.Data.NodeParam != nil {
			node.Data.NodeParam["IterationStartNodeId"] = startNodeID
			break
		}
	}
}

// getParentIterationID gets the iFlytek SparkAgent ID of the parent iteration node
func (g *IFlytekGenerator) getParentIterationID(node models.Node, iterationMap map[string]string) string {
	switch config := node.Config.(type) {
	case models.StartConfig:
		if config.IsInIteration && config.ParentID != "" {
			return iterationMap[config.ParentID]
		}
	case models.CodeConfig:
		if config.IsInIteration && config.IterationID != "" {
			return iterationMap[config.IterationID]
		}
	case models.LLMConfig:
		if config.IsInIteration && config.IterationID != "" {
			return iterationMap[config.IterationID]
		}
	}
	return ""
}

// needsRegeneration checks if a node needs to be regenerated (nodes that reference other nodes)
func (g *IFlytekGenerator) needsRegeneration(node models.Node) bool {
	// Only nodes that reference other nodes need to be regenerated
	switch node.Type {
	case models.NodeTypeEnd, models.NodeTypeLLM, models.NodeTypeCondition,
		models.NodeTypeCode, models.NodeTypeClassifier, models.NodeTypeIteration:
		// Check for input references
		for _, input := range node.Inputs {
			if input.Reference != nil && input.Reference.NodeID != "" {
				return true
			}
		}
		return false
	default:
		return false
	}
}
