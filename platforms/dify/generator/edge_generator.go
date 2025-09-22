package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"fmt"
	"regexp"
	"strings"
)

// EdgeGenerator generates connections
type EdgeGenerator struct {
	variableSelectorConverter *VariableSelectorConverter
}

func NewEdgeGenerator() *EdgeGenerator {
	return &EdgeGenerator{
		variableSelectorConverter: NewVariableSelectorConverter(),
	}
}

// GenerateEdges generates Dify connections
func (g *EdgeGenerator) GenerateEdges(edges []models.Edge, nodes []models.Node) ([]DifyEdge, error) {
	// Create empty ID mapping, use original IDs
	nodeIDMapping := make(map[string]string)
	for _, node := range nodes {
		nodeIDMapping[node.ID] = node.ID
	}

	return g.GenerateEdgesWithIDMapping(edges, nodes, nodeIDMapping)
}

// GenerateEdgesWithIDMapping generates Dify connections using ID mapping
func (g *EdgeGenerator) GenerateEdgesWithIDMapping(edges []models.Edge, nodes []models.Node, nodeIDMapping map[string]string) ([]DifyEdge, error) {
	difyEdges := make([]DifyEdge, 0, len(edges))

	for i, edge := range edges {

		// Skip edges connecting to iteration end nodes (these are not generated in Dify)
		if g.isIterationEndNode(edge.Target, nodes) {
			// Skip edge to iteration end node
			continue
		}

		// Verify nodes actually exist in the nodes list
		sourceNodeExists := false
		targetNodeExists := false
		for _, node := range nodes {
			if node.ID == edge.Source {
				sourceNodeExists = true
			}
			if node.ID == edge.Target {
				targetNodeExists = true
			}
		}

		// Skip connections pointing to non-existent nodes
		if !sourceNodeExists || !targetNodeExists {
			// Skip edge due to missing nodes
			continue
		}

		difyEdge, err := g.generateEdgeWithIDMapping(edge, nodes, nodeIDMapping, i)
		if err != nil {
			return nil, fmt.Errorf("failed to generate edge %s: %w", edge.ID, err)
		}
		difyEdges = append(difyEdges, difyEdge)
	}

	return difyEdges, nil
}

// generateEdgeWithIDMapping generates a single connection using ID mapping
func (g *EdgeGenerator) generateEdgeWithIDMapping(edge models.Edge, nodes []models.Node, nodeIDMapping map[string]string, index int) (DifyEdge, error) {
	// Get mapped source and target node IDs
	sourceID, sourceExists := nodeIDMapping[edge.Source]
	targetID, targetExists := nodeIDMapping[edge.Target]

	// If not in mapping, use original ID
	if !sourceExists {
		sourceID = edge.Source
	}
	if !targetExists {
		targetID = edge.Target
	}

	// Generate Dify standard edge ID format: sourceID-sourceHandle-targetID-targetHandle
	sourceHandle := "source"
	targetHandle := "target"

	// Set correct handles based on source node type
	sourceType := g.getNodeTypeByID(edge.Source, nodes)
	switch sourceType {
	case "if-else":
		sourceHandle = g.mapConditionHandle(edge.SourceHandle, nodes, edge.Source)
	case "question-classifier":
		sourceHandle = g.mapClassifierHandle(edge.SourceHandle, nodes, edge.Source)
	}

	// Generate standard Dify edge ID
	edgeID := fmt.Sprintf("%s-%s-%s-%s", sourceID, sourceHandle, targetID, targetHandle)

	difyEdge := DifyEdge{
		ID:           edgeID,
		Source:       sourceID,
		Target:       targetID,
		SourceHandle: sourceHandle,
		TargetHandle: targetHandle,
		Type:         "custom", // Dify default connection type
		ZIndex:       0,
		Data: DifyEdgeData{
			IsInLoop:   false,
			SourceType: sourceType,
			TargetType: g.getNodeTypeByID(edge.Target, nodes),
		},
	}

	// Handle iteration connections
	if g.isIterationEdge(edge, nodes) {
		difyEdge.Data.IsInIteration = true
		difyEdge.ZIndex = 1002 // Iteration internal connections use a higher zIndex
		if iterationID := g.getIterationID(edge, nodes); iterationID != "" {
			// If iteration ID also needs mapping, use the mapped ID
			if mappedIterationID, exists := nodeIDMapping[iterationID]; exists {
				difyEdge.Data.IterationID = mappedIterationID
			} else {
				difyEdge.Data.IterationID = iterationID
			}
		}
	}

	// Restore Dify platform-specific fields from the platform config
	if difyConfig := edge.PlatformConfig.Dify; difyConfig != nil {
		g.restoreDifyPlatformConfig(difyConfig, &difyEdge)
	}

	return difyEdge, nil
}

// getNodeTypeByID gets the node type based on the node ID
func (g *EdgeGenerator) getNodeTypeByID(nodeID string, nodes []models.Node) string {
	for _, node := range nodes {
		if node.ID == nodeID {
			// The start node of an iteration should be displayed as iteration-start in Dify
			if node.Type == models.NodeTypeStart && g.hasIterationParent(nodeID, nodes) {
				return "iteration-start"
			}
			return mapNodeTypeToDify(node.Type)
		}
	}
	return "unknown"
}

// isIterationEdge checks if it's an iteration connection
func (g *EdgeGenerator) isIterationEdge(edge models.Edge, nodes []models.Node) bool {
	// Only consider an internal iteration connection if both source and target are within a sub-workflow of an iteration
	sourceHasIterationParent := g.hasIterationParent(edge.Source, nodes)
	targetHasIterationParent := g.hasIterationParent(edge.Target, nodes)
	return sourceHasIterationParent && targetHasIterationParent
}

// hasIterationParent checks if a node has an iteration parent
func (g *EdgeGenerator) hasIterationParent(nodeID string, nodes []models.Node) bool {
	parentID := g.getIterationParentID(nodeID, nodes)
	return parentID != ""
}

// getIterationID gets the iteration ID
func (g *EdgeGenerator) getIterationID(edge models.Edge, nodes []models.Node) string {
	// First, check if the source or target node is directly an iteration node
	for _, node := range nodes {
		if node.Type == models.NodeTypeIteration {
			if node.ID == edge.Source || node.ID == edge.Target {
				return node.ID
			}
		}
	}

	// Check the iteration parent of the source node
	if parentID := g.getIterationParentID(edge.Source, nodes); parentID != "" {
		return parentID
	}

	// Check the iteration parent of the target node
	if parentID := g.getIterationParentID(edge.Target, nodes); parentID != "" {
		return parentID
	}

	return ""
}

// getIterationParentID gets the iteration parent ID of a node
func (g *EdgeGenerator) getIterationParentID(nodeID string, nodes []models.Node) string {
	node := g.findNodeByID(nodeID, nodes)
	if node == nil {
		return ""
	}

	// Check both platform configs for parent ID
	if parentID := g.getParentIDFromPlatformConfig(node); parentID != "" {
		if g.isIterationNode(parentID, nodes) {
			return parentID
		}
	}

	return ""
}

// findNodeByID finds a node by its ID
func (g *EdgeGenerator) findNodeByID(nodeID string, nodes []models.Node) *models.Node {
	for i, node := range nodes {
		if node.ID == nodeID {
			return &nodes[i]
		}
	}
	return nil
}

// getParentIDFromPlatformConfig extracts parent ID from platform configs
func (g *EdgeGenerator) getParentIDFromPlatformConfig(node *models.Node) string {
	// Check iFlytek config first
	if parentID := g.getParentIDFromConfig(node.PlatformConfig.IFlytek); parentID != "" {
		return parentID
	}

	// Check Dify config
	return g.getParentIDFromConfig(node.PlatformConfig.Dify)
}

// getParentIDFromConfig extracts parent ID from a config map
func (g *EdgeGenerator) getParentIDFromConfig(config map[string]interface{}) string {
	if config == nil {
		return ""
	}

	if parentID, ok := config["parentId"].(string); ok && parentID != "" {
		return parentID
	}

	return ""
}

// isIterationNode checks if a node is an iteration node
func (g *EdgeGenerator) isIterationNode(nodeID string, nodes []models.Node) bool {
	node := g.findNodeByID(nodeID, nodes)
	return node != nil && node.Type == models.NodeTypeIteration
}

// restoreDifyPlatformConfig restores Dify platform-specific fields
func (g *EdgeGenerator) restoreDifyPlatformConfig(config map[string]interface{}, edge *DifyEdge) {
	// Restore basic edge properties
	g.restoreBasicEdgeProperties(config, edge)

	// Restore edge data configuration
	g.restoreEdgeDataConfig(config, edge)

	// Restore handle configurations
	g.restoreHandleConfig(config, edge)
}

// restoreBasicEdgeProperties restores basic edge properties
func (g *EdgeGenerator) restoreBasicEdgeProperties(config map[string]interface{}, edge *DifyEdge) {
	if edgeType, ok := config["type"].(string); ok {
		edge.Type = edgeType
	}

	if zIndex, ok := config["zIndex"].(int); ok {
		edge.ZIndex = zIndex
	}

	if selected, ok := config["selected"].(bool); ok {
		edge.Selected = selected
	}
}

// restoreEdgeDataConfig restores edge data configuration
func (g *EdgeGenerator) restoreEdgeDataConfig(config map[string]interface{}, edge *DifyEdge) {
	dataConfig, ok := config["data"].(map[string]interface{})
	if !ok {
		return
	}

	if isInLoop, exists := dataConfig["isInLoop"].(bool); exists {
		edge.Data.IsInLoop = isInLoop
	}
	if sourceType, exists := dataConfig["sourceType"].(string); exists {
		edge.Data.SourceType = sourceType
	}
	if targetType, exists := dataConfig["targetType"].(string); exists {
		edge.Data.TargetType = targetType
	}
	if isInIteration, exists := dataConfig["isInIteration"].(bool); exists {
		edge.Data.IsInIteration = isInIteration
	}
	if iterationID, exists := dataConfig["iteration_id"].(string); exists {
		edge.Data.IterationID = iterationID
	}
}

// restoreHandleConfig restores source and target handles
func (g *EdgeGenerator) restoreHandleConfig(config map[string]interface{}, edge *DifyEdge) {
	if sourceHandle, ok := config["sourceHandle"].(string); ok {
		edge.SourceHandle = sourceHandle
	}
	if targetHandle, ok := config["targetHandle"].(string); ok {
		edge.TargetHandle = targetHandle
	}
}

// mapConditionHandle maps condition branch handles to Dify standard format
func (g *EdgeGenerator) mapConditionHandle(sourceHandle string, nodes []models.Node, sourceNodeID string) string {
	if g.isStandardConditionHandle(sourceHandle) {
		return sourceHandle
	}

	if g.isUUIDFormat(sourceHandle) {
		return sourceHandle
	}

	if mappedHandle := g.tryPlatformConfigMapping(nodes, sourceNodeID, sourceHandle); mappedHandle != "" {
		return mappedHandle
	}

	return "true" // Default fallback
}

// isStandardConditionHandle checks if handle is already in standard format
func (g *EdgeGenerator) isStandardConditionHandle(sourceHandle string) bool {
	return sourceHandle == "true" || sourceHandle == "false"
}

// isUUIDFormat checks if handle is in UUID format
func (g *EdgeGenerator) isUUIDFormat(sourceHandle string) bool {
	return len(sourceHandle) > 10 && !strings.Contains(sourceHandle, "branch_one_of::")
}

// tryPlatformConfigMapping tries to map using platform config
func (g *EdgeGenerator) tryPlatformConfigMapping(nodes []models.Node, sourceNodeID, sourceHandle string) string {
	for _, node := range nodes {
		if node.ID == sourceNodeID && node.Type == models.NodeTypeCondition {
			return g.extractMappingFromNode(node, sourceHandle)
		}
	}
	return ""
}

// extractMappingFromNode extracts mapping from node platform config
func (g *EdgeGenerator) extractMappingFromNode(node models.Node, sourceHandle string) string {
	if node.PlatformConfig.Dify == nil {
		return ""
	}

	mapping, exists := node.PlatformConfig.Dify["case_id_mapping"].(map[string]string)
	if !exists {
		return ""
	}

	if difyCaseID, found := mapping[sourceHandle]; found {
		return difyCaseID
	}
	return ""
}

// mapClassifierHandle maps classifier handles to Dify standard format
func (g *EdgeGenerator) mapClassifierHandle(sourceHandle string, nodes []models.Node, sourceNodeID string) string {
	// Early return for simple numeric formats
	if g.isSimpleNumericFormat(sourceHandle) {
		return sourceHandle
	}

	// Find the classifier node and attempt different mapping strategies
	classifierNode := g.findClassifierNode(nodes, sourceNodeID)
	if classifierNode != nil {
		if mappedHandle := g.tryConfigBasedMapping(sourceHandle, classifierNode); mappedHandle != "" {
			return mappedHandle
		}
	}

	// Apply smart mapping rules as fallback
	return g.applySmartMappingRules(sourceHandle, sourceNodeID)
}

// isSimpleNumericFormat checks if the handle is already in simple numeric format
func (g *EdgeGenerator) isSimpleNumericFormat(sourceHandle string) bool {
	return sourceHandle == "1" || sourceHandle == "2" || sourceHandle == "3" || sourceHandle == "4" || sourceHandle == "5"
}

// findClassifierNode finds the classifier node by ID
func (g *EdgeGenerator) findClassifierNode(nodes []models.Node, sourceNodeID string) *models.Node {
	for i, node := range nodes {
		if node.ID == sourceNodeID && node.Type == models.NodeTypeClassifier {
			return &nodes[i]
		}
	}
	return nil
}

// tryConfigBasedMapping attempts to map using node configuration
func (g *EdgeGenerator) tryConfigBasedMapping(sourceHandle string, node *models.Node) string {
	// Try unified DSL config mapping
	if mappedHandle := g.tryUnifiedConfigMapping(sourceHandle, node); mappedHandle != "" {
		return mappedHandle
	}

	// Try iFlytek platform config mapping
	if mappedHandle := g.tryIFlytekConfigMapping(sourceHandle, node); mappedHandle != "" {
		return mappedHandle
	}

	// Try Dify platform config mapping
	if mappedHandle := g.tryDifyConfigMapping(sourceHandle, node); mappedHandle != "" {
		return mappedHandle
	}

	return ""
}

// tryUnifiedConfigMapping tries to map using unified DSL configuration
func (g *EdgeGenerator) tryUnifiedConfigMapping(sourceHandle string, node *models.Node) string {
	config, ok := common.AsClassifierConfig(node.Config)
	if !ok || config == nil {
		return ""
	}

	for i, class := range config.Classes {
		if sourceHandle == class.ID {
			return g.generateSemanticClassID(class, i+1)
		}
	}
	return ""
}

// tryIFlytekConfigMapping tries to map using iFlytek platform configuration
func (g *EdgeGenerator) tryIFlytekConfigMapping(sourceHandle string, node *models.Node) string {
	if node.PlatformConfig.IFlytek == nil {
		return ""
	}

	nodeParam, ok := node.PlatformConfig.IFlytek["nodeParam"].(map[string]interface{})
	if !ok {
		return ""
	}

	intentChains, ok := nodeParam["intentChains"].([]interface{})
	if !ok {
		return ""
	}

	return g.mapByIntentChains(sourceHandle, intentChains)
}

// mapByIntentChains maps handle based on intent chains
func (g *EdgeGenerator) mapByIntentChains(sourceHandle string, intentChains []interface{}) string {
	for i, intentInterface := range intentChains {
		intentMap, ok := intentInterface.(map[string]interface{})
		if !ok {
			continue
		}

		intentID, exists := intentMap["id"].(string)
		if exists && sourceHandle == intentID {
			return fmt.Sprintf("%d", i+1)
		}
	}
	return ""
}

// tryDifyConfigMapping tries to map using Dify platform configuration
func (g *EdgeGenerator) tryDifyConfigMapping(sourceHandle string, node *models.Node) string {
	if node.PlatformConfig.Dify == nil {
		return ""
	}

	configMap, ok := node.PlatformConfig.Dify["config"].(map[string]interface{})
	if !ok {
		return ""
	}

	classes, ok := configMap["classes"].([]interface{})
	if !ok {
		return ""
	}

	return g.mapByClassID(sourceHandle, classes)
}

// mapByClassID maps handle based on class ID
func (g *EdgeGenerator) mapByClassID(sourceHandle string, classes []interface{}) string {
	for i, classInterface := range classes {
		classMap, ok := classInterface.(map[string]interface{})
		if !ok {
			continue
		}

		classID, exists := classMap["id"].(string)
		if !exists {
			continue
		}

		if sourceHandle == classID || sourceHandle == "intent-one-of::"+classID {
			return fmt.Sprintf("%d", i+1)
		}
	}
	return ""
}

// applySmartMappingRules applies smart mapping rules as fallback
func (g *EdgeGenerator) applySmartMappingRules(sourceHandle, sourceNodeID string) string {
	// Handle "intent-one-of::number" format
	if mappedHandle := g.tryIntentOneOfMapping(sourceHandle); mappedHandle != "" {
		return mappedHandle
	}

	// Hash-based mapping for long IDs
	if mappedHandle := g.tryHashBasedMapping(sourceHandle); mappedHandle != "" {
		return mappedHandle
	}

	// Final fallback with logging
	return "1"
}

// tryIntentOneOfMapping handles "intent-one-of::number" format
func (g *EdgeGenerator) tryIntentOneOfMapping(sourceHandle string) string {
	if !strings.HasPrefix(sourceHandle, "intent-one-of::") {
		return ""
	}

	suffix := strings.TrimPrefix(sourceHandle, "intent-one-of::")
	if len(suffix) <= 2 && suffix >= "1" && suffix <= "9" {
		return suffix
	}
	return ""
}

// tryHashBasedMapping uses hash for consistent mapping of long IDs
func (g *EdgeGenerator) tryHashBasedMapping(sourceHandle string) string {
	if len(sourceHandle) <= 10 {
		return ""
	}

	// Use a simple hash algorithm to ensure consistency
	hash := 0
	for _, c := range sourceHandle {
		hash = hash*31 + int(c)
	}

	// Map to the range 1-9
	mappedNum := (hash % 9) + 1
	if mappedNum < 1 {
		mappedNum = 1
	}

	return fmt.Sprintf("%d", mappedNum)
}

// generateSemanticClassID generates semantic class ID consistent with ClassifierNodeGenerator
func (g *EdgeGenerator) generateSemanticClassID(class models.ClassifierClass, fallbackIndex int) string {
	if class.IsDefault {
		return "default"
	}

	// Generate semantic ID based on class name, ensuring uniqueness
	if class.Name != "" {
		// Convert to safe ID format: remove spaces, convert to lowercase
		safeID := strings.ToLower(strings.ReplaceAll(class.Name, " ", "_"))
		// Remove special characters and keep only alphanumeric and underscore
		if matched, _ := regexp.MatchString(`^[a-z0-9_]+$`, safeID); matched && safeID != "" {
			return safeID
		}
	}

	// Fallback to numeric ID if name processing fails
	return fmt.Sprintf("%d", fallbackIndex)
}

// isIterationEndNode checks if a node ID represents an iteration end node that should be skipped in Dify
func (g *EdgeGenerator) isIterationEndNode(nodeID string, nodes []models.Node) bool {
	// Check if this is an iteration-node-end pattern (these are always internal)
	if strings.Contains(nodeID, "iteration-node-end") {
		return true
	}

	// Find the actual node
	targetNode := g.findNodeByID(nodeID, nodes)
	if targetNode == nil {
		return false
	}

	// Only filter out end nodes that are INSIDE an iteration (have an iteration parent)
	if targetNode.Type == models.NodeTypeEnd {
		// Check if this end node has an iteration parent
		return g.hasIterationParent(nodeID, nodes)
	}

	return false
}
