package parser

import (
	"github.com/iflytek/agentbridge/internal/models"
	"fmt"
	"strings"
)

// IFlytekEdge represents iFlytek SparkAgent edge structure.
type IFlytekEdge struct {
	ID           string                 `yaml:"id"`
	Source       string                 `yaml:"source"`
	Target       string                 `yaml:"target"`
	Type         string                 `yaml:"type"`
	SourceHandle string                 `yaml:"sourceHandle,omitempty"`
	TargetHandle string                 `yaml:"targetHandle,omitempty"`
	MarkerEnd    map[string]interface{} `yaml:"markerEnd,omitempty"`
	Data         map[string]interface{} `yaml:"data,omitempty"`
	ZIndex       int                    `yaml:"zIndex,omitempty"`
}

// EdgeParser parses edges.
type EdgeParser struct {
	variableRefSystem *models.VariableReferenceSystem
}

func NewEdgeParser(variableRefSystem *models.VariableReferenceSystem) *EdgeParser {
	return &EdgeParser{
		variableRefSystem: variableRefSystem,
	}
}

// ParseEdges parses edge list.
func (p *EdgeParser) ParseEdges(iflytekEdges []interface{}) ([]models.Edge, error) {
	if len(iflytekEdges) == 0 {
		return []models.Edge{}, nil
	}

	edges := make([]models.Edge, 0, len(iflytekEdges))

	for i, edgeData := range iflytekEdges {
		edgeMap, ok := edgeData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid edge data at index %d: expected map", i)
		}

		iflytekEdge, err := p.parseIFlytekEdge(edgeMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse edge at index %d: %w", i, err)
		}

		unifiedEdge, err := p.ParseEdge(*iflytekEdge)
		if err != nil {
			return nil, fmt.Errorf("failed to convert edge at index %d: %w", i, err)
		}

		edges = append(edges, *unifiedEdge)
	}

	return edges, nil
}

// parseIFlytekEdge parses iFlytek SparkAgent edge data to structure.
func (p *EdgeParser) parseIFlytekEdge(edgeMap map[string]interface{}) (*IFlytekEdge, error) {
	edge := &IFlytekEdge{}

	if id, ok := edgeMap["id"].(string); ok {
		edge.ID = id
	}

	if source, ok := edgeMap["source"].(string); ok {
		edge.Source = source
	} else {
		return nil, fmt.Errorf("edge missing required source field")
	}

	if target, ok := edgeMap["target"].(string); ok {
		edge.Target = target
	} else {
		return nil, fmt.Errorf("edge missing required target field")
	}

	if edgeType, ok := edgeMap["type"].(string); ok {
		edge.Type = edgeType
	}

	if sourceHandle, ok := edgeMap["sourceHandle"].(string); ok {
		edge.SourceHandle = sourceHandle
	}

	if targetHandle, ok := edgeMap["targetHandle"].(string); ok {
		edge.TargetHandle = targetHandle
	}

	if markerEnd, ok := edgeMap["markerEnd"].(map[string]interface{}); ok {
		edge.MarkerEnd = markerEnd
	}

	if data, ok := edgeMap["data"].(map[string]interface{}); ok {
		edge.Data = data
	}

	if zIndex, ok := edgeMap["zIndex"].(float64); ok {
		edge.ZIndex = int(zIndex)
	}

	return edge, nil
}

// ParseEdge parses a single edge to unified format.
func (p *EdgeParser) ParseEdge(iflytekEdge IFlytekEdge) (*models.Edge, error) {
	// Create unified edge structure
	unifiedEdge := &models.Edge{
		ID:     p.generateUnifiedEdgeID(iflytekEdge),
		Source: iflytekEdge.Source,
		Target: iflytekEdge.Target,
		Type:   p.determineEdgeType(iflytekEdge),
	}

	// Parse source and target handles
	unifiedEdge.SourceHandle = iflytekEdge.SourceHandle
	unifiedEdge.TargetHandle = iflytekEdge.TargetHandle

	// Parse condition (if it's a conditional connection)
	if condition := p.extractCondition(iflytekEdge); condition != "" {
		unifiedEdge.Condition = condition
	}

	// Save platform-specific configuration
	unifiedEdge.PlatformConfig = models.PlatformConfig{
		IFlytek: p.preserveIFlytekPlatformFields(iflytekEdge),
		Dify:    make(map[string]interface{}),
	}

	return unifiedEdge, nil
}

// generateUnifiedEdgeID generates unified edge ID.
func (p *EdgeParser) generateUnifiedEdgeID(iflytekEdge IFlytekEdge) string {
	if iflytekEdge.ID != "" {
		return iflytekEdge.ID
	}

	// If no ID, generate one based on source and target
	sourceID := p.extractNodeID(iflytekEdge.Source)
	targetID := p.extractNodeID(iflytekEdge.Target)

	if iflytekEdge.SourceHandle != "" {
		return fmt.Sprintf("%s_%s_%s", sourceID, iflytekEdge.SourceHandle, targetID)
	}

	return fmt.Sprintf("%s_%s", sourceID, targetID)
}

// extractNodeID extracts short ID from complete node ID.
func (p *EdgeParser) extractNodeID(fullNodeID string) string {
	// Extract type part from "node-type::uuid" format
	parts := strings.Split(fullNodeID, "::")
	if len(parts) >= 1 {
		return strings.ReplaceAll(parts[0], "-", "_")
	}
	return fullNodeID
}

// determineEdgeType determines edge type.
func (p *EdgeParser) determineEdgeType(iflytekEdge IFlytekEdge) models.EdgeType {
	// If there's a source handle, it's usually a conditional connection
	if iflytekEdge.SourceHandle != "" {
		// Check if it's a branch condition
		if strings.Contains(iflytekEdge.SourceHandle, "branch_one_of") ||
			strings.Contains(iflytekEdge.SourceHandle, "intent-one-of") {
			return models.EdgeTypeConditional
		}
	}

	// Default to normal connection
	return models.EdgeTypeDefault
}

// extractCondition extracts condition information.
func (p *EdgeParser) extractCondition(iflytekEdge IFlytekEdge) string {
	if iflytekEdge.SourceHandle == "" {
		return ""
	}

	// Extract condition information from source handle
	if strings.Contains(iflytekEdge.SourceHandle, "branch_one_of") {
		return fmt.Sprintf("branch:%s", iflytekEdge.SourceHandle)
	}

	if strings.Contains(iflytekEdge.SourceHandle, "intent-one-of") {
		return fmt.Sprintf("intent:%s", iflytekEdge.SourceHandle)
	}

	return ""
}

// preserveIFlytekPlatformFields preserves iFlytek Platform specific fields.
func (p *EdgeParser) preserveIFlytekPlatformFields(iflytekEdge IFlytekEdge) map[string]interface{} {
	platformFields := make(map[string]interface{})

	// Preserve original ID
	if iflytekEdge.ID != "" {
		platformFields["original_id"] = iflytekEdge.ID
	}

	// Preserve original type
	if iflytekEdge.Type != "" {
		platformFields["original_type"] = iflytekEdge.Type
	}

	// Preserve marker end configuration
	if iflytekEdge.MarkerEnd != nil {
		platformFields["marker_end"] = iflytekEdge.MarkerEnd
	}

	// Preserve data configuration
	if iflytekEdge.Data != nil {
		platformFields["data"] = iflytekEdge.Data
	}

	// Preserve layer information
	if iflytekEdge.ZIndex != 0 {
		platformFields["z_index"] = iflytekEdge.ZIndex
	}

	return platformFields
}
