package common

import (
	"agentbridge/internal/models"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// GraphGenerationContext holds context information for graph generation
type GraphGenerationContext struct {
	UnifiedDSL    *models.UnifiedDSL
	NodeIDMapping map[string]string
	Graph         interface{} // Generic graph type (could be DifyGraph, IflytekGraph, etc.)
}

// GenerateSimpleNodeID generates a simple random 16-digit numeric ID
func GenerateSimpleNodeID(node models.Node, index int) string {
	// Use timestamp + random number combination (Go 1.20+ auto-seeding)
	timestamp := time.Now().UnixMilli() % 10000000000 // 10-digit millisecond timestamp
	randomPart := rand.Uint64() % 10000               // 4-digit random number

	// Combine into 16-digit number: 10-digit timestamp + 2-digit index + 4-digit random number
	nodeID := fmt.Sprintf("%010d%02d%04d", timestamp, index%100, randomPart)

	return nodeID
}

// CollectIterationInternalNodeIDs collects all iteration internal node IDs for filtering
func CollectIterationInternalNodeIDs(nodes []models.Node) map[string]bool {
	iterationInternalNodeIDs := make(map[string]bool)

	for _, node := range nodes {
		if node.Type == models.NodeTypeIteration {
			if iterConfig, ok := node.Config.(*models.IterationConfig); ok {
				for _, subNode := range iterConfig.SubWorkflow.Nodes {
					iterationInternalNodeIDs[subNode.ID] = true
				}
			}
		}
	}

	return iterationInternalNodeIDs
}

// CollectAllEdgesAndNodes collects all edges and nodes from main workflow and iterations
func CollectAllEdgesAndNodes(unifiedDSL *models.UnifiedDSL, iterationInternalNodeIDs map[string]bool) ([]models.Edge, []models.Node) {
	allEdges := make([]models.Edge, 0)
	allNodes := make([]models.Node, 0)

	// Add main workflow edges (filter out iteration internal edges)
	for _, edge := range unifiedDSL.Workflow.Edges {
		// If both source and target nodes of the edge are not iteration internal nodes, add to main workflow edges
		sourceIsInternal := iterationInternalNodeIDs[edge.Source]
		targetIsInternal := iterationInternalNodeIDs[edge.Target]

		if !sourceIsInternal && !targetIsInternal {
			// This is a main workflow edge (both ends are not iteration internal nodes)
			allEdges = append(allEdges, edge)
		}
	}

	// Add main workflow nodes
	allNodes = append(allNodes, unifiedDSL.Workflow.Nodes...)

	// Add iteration sub-workflow edges and nodes
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == models.NodeTypeIteration {
			if iterConfig, ok := node.Config.(*models.IterationConfig); ok {
				// Add sub-workflow edges
				allEdges = append(allEdges, iterConfig.SubWorkflow.Edges...)
				// Add sub-workflow nodes
				allNodes = append(allNodes, iterConfig.SubWorkflow.Nodes...)
			}
		}
	}

	return allEdges, allNodes
}

// NodeGenerationPhase represents the phase of node generation
type NodeGenerationPhase int

const (
	PhaseInitialGeneration NodeGenerationPhase = iota
	PhaseRegeneration
)

// NodeGenerationContext holds context for node generation process
type NodeGenerationContext struct {
	Phase         NodeGenerationPhase
	IDMapping     map[string]string
	TitleMapping  map[string]string
	BranchMapping map[string]string
	SkipSubNodes  bool
}

// ReverseIDMapping performs reverse lookup in ID mapping to find original ID
func ReverseIDMapping(idMapping map[string]string, targetID string) string {
	if idMapping == nil {
		return ""
	}

	for originalID, mappedID := range idMapping {
		if mappedID == targetID {
			return originalID
		}
	}
	return ""
}

// CreateVariableReference creates a variable reference with given parameters
func CreateVariableReference(originalRef *models.VariableReference, nodeID, outputName string) *models.VariableReference {
	if nodeID == "" {
		return originalRef
	}

	return &models.VariableReference{
		Type:       originalRef.Type,
		NodeID:     nodeID,
		OutputName: outputName,
		DataType:   originalRef.DataType,
	}
}

// TryRemapNodeID attempts to remap a node ID using the provided mapping
func TryRemapNodeID(idMapping map[string]string, nodeID string) (string, bool) {
	if idMapping == nil {
		return "", false
	}

	if mappedID, exists := idMapping[nodeID]; exists {
		return mappedID, true
	}

	return "", false
}

// UpdateVariableSelector updates a variable selector array with mapped node ID
func UpdateVariableSelector(selector []string, nodeIDMapping map[string]string) []string {
	if len(selector) < 2 {
		return selector
	}

	oldNodeID := selector[0]
	if newNodeID, found := nodeIDMapping[oldNodeID]; found {
		return []string{newNodeID, selector[1]}
	}

	return selector
}

// ReplaceTemplateNodeReferences replaces node ID references in template strings
func ReplaceTemplateNodeReferences(text string, nodeIDMapping map[string]string) string {
	for oldID, newID := range nodeIDMapping {
		oldPattern := fmt.Sprintf("{{#%s.", oldID)
		newPattern := fmt.Sprintf("{{#%s.", newID)
		text = strings.ReplaceAll(text, oldPattern, newPattern)
	}
	return text
}
