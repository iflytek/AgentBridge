package generator

import (
	"agentbridge/core/interfaces"
	"agentbridge/internal/models"
	"agentbridge/platforms/common"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Interface compliance check at compile time
var _ interfaces.DSLGenerator = (*CozeGenerator)(nil)

// CozeGenerator implements DSL generation for ByteDance Coze workflow platform
type CozeGenerator struct {
	*common.BaseGenerator
	nodeGeneratorFactory *NodeGeneratorFactory
	edgeGenerator        *EdgeGenerator
	idGenerator          *CozeIDGenerator
}

// NewCozeGenerator creates a Coze DSL generator
func NewCozeGenerator() *CozeGenerator {
	idGenerator := NewCozeIDGenerator()
	edgeGenerator := NewEdgeGenerator()
	edgeGenerator.idGenerator = idGenerator // Share the same ID generator instance

	return &CozeGenerator{
		BaseGenerator:        common.NewBaseGenerator(models.PlatformCoze),
		nodeGeneratorFactory: NewNodeGeneratorFactory(),
		edgeGenerator:        edgeGenerator,
		idGenerator:          idGenerator,
	}
}

// Generate generates Coze DSL from unified DSL
func (g *CozeGenerator) Generate(unifiedDSL *models.UnifiedDSL) ([]byte, error) {
	// Validate input
	if err := g.Validate(unifiedDSL); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set unified DSL reference for edge generator context
	g.edgeGenerator.SetUnifiedDSL(unifiedDSL)

	// Build Coze DSL structure
	cozeDSL := &CozeRootStructure{}

	// Generate workflow metadata
	if err := g.generateWorkflowMetadata(unifiedDSL, cozeDSL); err != nil {
		return nil, fmt.Errorf("failed to generate workflow metadata: %w", err)
	}

	// Generate schema section
	if err := g.generateSchema(unifiedDSL, cozeDSL); err != nil {
		return nil, fmt.Errorf("failed to generate schema: %w", err)
	}

	// Generate nodes
	if err := g.generateNodes(unifiedDSL, cozeDSL); err != nil {
		return nil, fmt.Errorf("failed to generate nodes: %w", err)
	}

	// Generate edges
	if err := g.generateEdges(unifiedDSL, cozeDSL); err != nil {
		return nil, fmt.Errorf("failed to generate edges: %w", err)
	}

	// Generate metadata and dependencies
	g.generateMetadataAndDependencies(unifiedDSL, cozeDSL)

	// Convert to YAML
	return yaml.Marshal(cozeDSL)
}

// generateWorkflowMetadata generates workflow-level metadata
func (g *CozeGenerator) generateWorkflowMetadata(unifiedDSL *models.UnifiedDSL, cozeDSL *CozeRootStructure) error {
	// Generate workflow ID (use timestamp-based approach similar to example)
	cozeDSL.WorkflowID = g.idGenerator.GenerateWorkflowID()

	// Set basic metadata
	cozeDSL.Name = unifiedDSL.Metadata.Name
	if cozeDSL.Name == "" {
		cozeDSL.Name = "Generated Coze Workflow"
	}

	cozeDSL.Description = unifiedDSL.Metadata.Description
	cozeDSL.Version = ""

	// Set timestamps
	now := time.Now().Unix()
	cozeDSL.CreateTime = now
	cozeDSL.UpdateTime = now

	// Set export format
	cozeDSL.ExportFormat = "yml"
	cozeDSL.SerializedData = ""

	return nil
}

// generateSchema generates the schema section
func (g *CozeGenerator) generateSchema(unifiedDSL *models.UnifiedDSL, cozeDSL *CozeRootStructure) error {
	schema := &CozeSchema{
		Edges: make([]CozeSchemaEdge, 0),
		Nodes: make([]CozeSchemaNode, 0),
		Versions: &CozeVersions{
			Loop: "v2",
		},
	}

	// Generate schema edges
	for _, edge := range unifiedDSL.Workflow.Edges {
		cozeEdge := g.edgeGenerator.GenerateSchemaEdge(&edge)

		// Handle iteration node outputs with required sourcePortID
		sourceNode := g.findNodeByID(edge.Source, unifiedDSL)
		targetNode := g.findNodeByID(edge.Target, unifiedDSL)

		if sourceNode != nil && sourceNode.Type == models.NodeTypeIteration {
			// Iteration to other nodes must use loop-output port
			cozeEdge.SourcePortID = "loop-output"
			// Clear targetPortID as it should not be used
			cozeEdge.TargetPortID = ""
		}

		// Ensure connections to end nodes have empty target ports
		if targetNode != nil && targetNode.Type == models.NodeTypeEnd {
			cozeEdge.TargetPortID = ""
		}

		schema.Edges = append(schema.Edges, *cozeEdge)
	}

	// Generate schema nodes (simplified version)
	for _, node := range unifiedDSL.Workflow.Nodes {
		generator, err := g.nodeGeneratorFactory.GetNodeGenerator(node.Type)
		if err != nil {
			// Skip unsupported node types, continue processing other nodes
			continue
		}

		// Set the shared ID generator
		generator.SetIDGenerator(g.idGenerator)

		// Set dependencies for iteration node generator in schema generation
		if node.Type == models.NodeTypeIteration {
			if iterationGen, ok := generator.(*IterationNodeGenerator); ok {
				iterationGen.SetNodeFactory(g.nodeGeneratorFactory)
				iterationGen.SetEdgeGenerator(g.edgeGenerator)
			}
		}

		schemaNode, err := generator.GenerateSchemaNode(&node)
		if err != nil {
			return fmt.Errorf("failed to generate schema node %s (type: %s): %w", node.ID, node.Type, err)
		}

		schema.Nodes = append(schema.Nodes, *schemaNode)
	}

	cozeDSL.Schema = schema
	return nil
}

// generateNodes generates workflow nodes
func (g *CozeGenerator) generateNodes(unifiedDSL *models.UnifiedDSL, cozeDSL *CozeRootStructure) error {
	nodes := make([]CozeNode, 0)

	for _, node := range unifiedDSL.Workflow.Nodes {
		generator, err := g.nodeGeneratorFactory.GetNodeGenerator(node.Type)
		if err != nil {
			// Skip unsupported node types, continue processing other nodes
			continue
		}

		// Set the shared ID generator
		generator.SetIDGenerator(g.idGenerator)

		// Set additional dependencies for iteration node generator
		if node.Type == models.NodeTypeIteration {
			if iterationGen, ok := generator.(*IterationNodeGenerator); ok {
				iterationGen.SetNodeFactory(g.nodeGeneratorFactory)
				iterationGen.SetEdgeGenerator(g.edgeGenerator)
			}
		}

		cozeNode, err := generator.GenerateNode(&node)
		if err != nil {
			return fmt.Errorf("failed to generate node %s (type: %s): %w", node.ID, node.Type, err)
		}

		nodes = append(nodes, *cozeNode)
	}

	cozeDSL.Nodes = nodes
	return nil
}

// generateEdges generates workflow edges
func (g *CozeGenerator) generateEdges(unifiedDSL *models.UnifiedDSL, cozeDSL *CozeRootStructure) error {
	edges := make([]CozeEdge, 0)

	for _, edge := range unifiedDSL.Workflow.Edges {
		cozeEdge := g.edgeGenerator.GenerateEdge(&edge)
		g.addIterationPortsIfNeeded(cozeEdge, &edge, unifiedDSL)
		edges = append(edges, *cozeEdge)
	}

	cozeDSL.Edges = edges
	return nil
}

// addIterationPortsIfNeeded adds special loop ports for iteration nodes and fixes end node connections
func (g *CozeGenerator) addIterationPortsIfNeeded(cozeEdge *CozeEdge, originalEdge *models.Edge, unifiedDSL *models.UnifiedDSL) {
	sourceNode := g.findNodeByID(originalEdge.Source, unifiedDSL)
	targetNode := g.findNodeByID(originalEdge.Target, unifiedDSL)

	// Handle iteration node outputs
	if sourceNode != nil && sourceNode.Type == models.NodeTypeIteration {
		cozeEdge.FromPort = "loop-output"
	}

	// CRITICAL: All connections to end nodes must have empty target ports in Coze format
	if targetNode != nil && targetNode.Type == models.NodeTypeEnd {
		cozeEdge.ToPort = ""
	}
}

// findNodeByID finds a node by its ID
func (g *CozeGenerator) findNodeByID(nodeID string, unifiedDSL *models.UnifiedDSL) *models.Node {
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.ID == nodeID {
			return &node
		}
	}
	return nil
}

// generateMetadataAndDependencies generates metadata and dependencies sections
func (g *CozeGenerator) generateMetadataAndDependencies(unifiedDSL *models.UnifiedDSL, cozeDSL *CozeRootStructure) {
	// Generate metadata
	cozeDSL.Metadata = &CozeMetadata{
		ContentType: "0",
		CreatorID:   "generated_creator_id",
		Mode:        "0",
		SpaceID:     "generated_space_id",
	}

	// Use platform metadata if available
	if unifiedDSL.PlatformMetadata.Coze != nil {
		if unifiedDSL.PlatformMetadata.Coze.ContentType != "" {
			cozeDSL.Metadata.ContentType = unifiedDSL.PlatformMetadata.Coze.ContentType
		}
		if unifiedDSL.PlatformMetadata.Coze.CreatorID != "" {
			cozeDSL.Metadata.CreatorID = unifiedDSL.PlatformMetadata.Coze.CreatorID
		}
		if unifiedDSL.PlatformMetadata.Coze.Mode != "" {
			cozeDSL.Metadata.Mode = unifiedDSL.PlatformMetadata.Coze.Mode
		}
		if unifiedDSL.PlatformMetadata.Coze.SpaceID != "" {
			cozeDSL.Metadata.SpaceID = unifiedDSL.PlatformMetadata.Coze.SpaceID
		}
	}

	// Generate dependencies for each node
	dependencies := make([]CozeDependency, 0)
	for _, node := range unifiedDSL.Workflow.Nodes {
		dependency := CozeDependency{
			Metadata: &CozeDependencyMeta{
				NodeType: "workflow_node",
			},
			ResourceID:   "node_" + g.idGenerator.MapToCozeNodeID(node.ID),
			ResourceName: node.Title,
			ResourceType: "node",
		}
		dependencies = append(dependencies, dependency)
	}

	cozeDSL.Dependencies = dependencies
}

// Validate validates the unified DSL before generation
func (g *CozeGenerator) Validate(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL is nil")
	}

	if len(unifiedDSL.Workflow.Nodes) == 0 {
		return fmt.Errorf("no nodes found in workflow")
	}

	// Check for required start and end nodes (Phase 1 requirement)
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
		return fmt.Errorf("workflow must have a start node")
	}

	if !hasEnd {
		return fmt.Errorf("workflow must have an end node")
	}

	return nil
}

// GetPlatformType returns the platform type
func (g *CozeGenerator) GetPlatformType() models.PlatformType {
	return models.PlatformCoze
}

// CozeIDGenerator handles ID generation and mapping for Coze platform
type CozeIDGenerator struct {
	nodeIDCounter      int
	nodeIDMapping      map[string]string // unified ID -> coze ID
	currentIterationID string            // Current iteration node ID being processed
}

// NewCozeIDGenerator creates a Coze ID generator
func NewCozeIDGenerator() *CozeIDGenerator {
	return &CozeIDGenerator{
		nodeIDCounter: 197161, // Start from 197161 like in example (LLM node ID)
		nodeIDMapping: make(map[string]string),
	}
}

// SetCurrentIterationNodeID sets the current iteration node ID for edge generation
func (g *CozeIDGenerator) SetCurrentIterationNodeID(nodeID string) {
	g.currentIterationID = nodeID
}

// GetCurrentIterationNodeID gets the current iteration node ID for edge generation
func (g *CozeIDGenerator) GetCurrentIterationNodeID() string {
	return g.currentIterationID
}

// GenerateWorkflowID generates a workflow ID
func (g *CozeIDGenerator) GenerateWorkflowID() string {
	// Generate timestamp-based ID similar to example
	return strconv.FormatInt(time.Now().UnixNano()/1000000, 10)
}

// MapToCozeNodeID dynamically maps unified node ID to Coze node ID based on node type
func (g *CozeIDGenerator) MapToCozeNodeID(unifiedID string) string {
	if cozeID, exists := g.nodeIDMapping[unifiedID]; exists {
		return cozeID
	}

	// CRITICAL: Handle iteration internal node reference remapping
	if strings.Contains(unifiedID, "iteration-node-start::") && g.currentIterationID != "" {
		fmt.Printf("✅ Mapped internal start node %s -> %s\n", unifiedID, g.currentIterationID)
		g.nodeIDMapping[unifiedID] = g.currentIterationID
		return g.currentIterationID
	}

	// If this is an iteration internal end node, also map to current iteration node (for output collection)
	if strings.Contains(unifiedID, "iteration-node-end::") && g.currentIterationID != "" {
		fmt.Printf("✅ Mapped internal end node %s -> %s\n", unifiedID, g.currentIterationID)
		g.nodeIDMapping[unifiedID] = g.currentIterationID
		return g.currentIterationID
	}

	// Extract node type information from unified ID (if possible)
	nodeType := g.extractNodeTypeFromID(unifiedID)

	// Generate corresponding Coze ID based on node type
	var cozeID string
	switch nodeType {
	case "start":
		cozeID = "100001" // Start node uses fixed ID
	case "end":
		cozeID = "900001" // End node uses fixed ID
	case "llm":
		// LLM nodes should get incremental IDs, starting from 197161 (based on example file)
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	default:
		// Other node types use incremental ID
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	}

	g.nodeIDMapping[unifiedID] = cozeID
	return cozeID
}

// extractNodeTypeFromID extracts node type from unified node ID
func (g *CozeIDGenerator) extractNodeTypeFromID(unifiedID string) string {
	// iFlytek format: node-start::uuid, node-end::uuid etc.
	if strings.HasPrefix(unifiedID, "node-start::") {
		return "start"
	}
	if strings.HasPrefix(unifiedID, "node-end::") {
		return "end"
	}
	if strings.HasPrefix(unifiedID, "spark-llm::") {
		return "llm"
	}
	if strings.HasPrefix(unifiedID, "code-node::") {
		return "code"
	}
	if strings.HasPrefix(unifiedID, "condition-node::") {
		return "condition"
	}
	if strings.HasPrefix(unifiedID, "classifier-node::") {
		return "classifier"
	}
	if strings.HasPrefix(unifiedID, "iteration-node::") {
		return "iteration"
	}

	// Return unknown if cannot be identified
	return "unknown"
}

// MapToCozeNodeIDByType generates Coze ID directly based on node type (fallback method)
func (g *CozeIDGenerator) MapToCozeNodeIDByType(unifiedID string, nodeType models.NodeType) string {
	if cozeID, exists := g.nodeIDMapping[unifiedID]; exists {
		return cozeID
	}

	var cozeID string
	switch nodeType {
	case models.NodeTypeStart:
		cozeID = "100001"
	case models.NodeTypeEnd:
		cozeID = "900001"
	case models.NodeTypeLLM:
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	case models.NodeTypeCode:
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	case models.NodeTypeCondition:
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	case models.NodeTypeClassifier:
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	case models.NodeTypeIteration:
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	default:
		cozeID = strconv.Itoa(g.nodeIDCounter)
		g.nodeIDCounter++
	}

	g.nodeIDMapping[unifiedID] = cozeID
	return cozeID
}
