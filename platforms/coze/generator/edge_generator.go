package generator

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
	"strings"
)

// EdgeGenerator handles Coze workflow edge generation and port mapping between platforms.
type EdgeGenerator struct {
	idGenerator *CozeIDGenerator
	unifiedDSL  *models.UnifiedDSL // Unified DSL reference for context-aware mapping
}

// NewEdgeGenerator creates an edge generator with platform-specific ID mapping.
func NewEdgeGenerator() *EdgeGenerator {
	return &EdgeGenerator{
		idGenerator: NewCozeIDGenerator(),
	}
}

// SetUnifiedDSL configures the unified DSL reference for context-aware port mapping.
func (g *EdgeGenerator) SetUnifiedDSL(unifiedDSL *models.UnifiedDSL) {
	g.unifiedDSL = unifiedDSL
}

// GenerateEdge converts unified edge definitions to Coze edge format.
func (g *EdgeGenerator) GenerateEdge(unifiedEdge *models.Edge) *CozeEdge {
	fromPort := g.mapToCozePort(unifiedEdge.SourceHandle)
	// Coze format does not use target port for normal edges
	toPort := ""

	return &CozeEdge{
		FromNode: g.idGenerator.MapToCozeNodeID(unifiedEdge.Source),
		FromPort: fromPort,
		ToNode:   g.idGenerator.MapToCozeNodeID(unifiedEdge.Target),
		ToPort:   toPort,
	}
}

// GenerateSchemaEdge converts unified edge definitions to Coze schema edge format.
func (g *EdgeGenerator) GenerateSchemaEdge(unifiedEdge *models.Edge) *CozeSchemaEdge {
	fromPort := g.mapToCozePort(unifiedEdge.SourceHandle)
	// Coze schema edges typically omit targetPortID
	toPort := ""

	edge := &CozeSchemaEdge{
		SourceNodeID: g.idGenerator.MapToCozeNodeID(unifiedEdge.Source),
		TargetNodeID: g.idGenerator.MapToCozeNodeID(unifiedEdge.Target),
	}
	if fromPort != "" {
		edge.SourcePortID = fromPort
	}
	if toPort != "" {
		edge.TargetPortID = toPort
	}
	return edge
}

// mapToCozePort transforms unified port handles to Coze-specific port identifiers.
func (g *EdgeGenerator) mapToCozePort(handle string) string {
	if handle == "" {
		return ""
	}
	// Ignore generic Dify handles for Coze format
	if handle == "source" || handle == "target" {
		return ""
	}

	// Handle iFlytek branch format: branch_one_of::xxx (for condition nodes)
	if strings.HasPrefix(handle, "branch_one_of::") {
		return g.mapBranchToCozePort(handle)
	}

	// Handle iFlytek classifier intent format: intent-one-of::xxx (for classifier nodes)
	if strings.HasPrefix(handle, "intent-one-of::") {
		return g.mapIntentToCozePort(handle)
	}

	// Do not pass through unknown handles
	return ""
}

// mapBranchToCozePort converts iFlytek branch handles to Coze condition port format.
func (g *EdgeGenerator) mapBranchToCozePort(branchHandle string) string {
	// Extract UUID from branch_one_of::uuid format
	parts := strings.Split(branchHandle, "::")
	if len(parts) != 2 {
		return ""
	}

	branchID := parts[1]

	// Look up the branch level in unified DSL if available
	if g.unifiedDSL != nil {
		for _, node := range g.unifiedDSL.Workflow.Nodes {
			if node.Type == models.NodeTypeCondition {
				conditionConfig, ok := node.Config.(*models.ConditionConfig)
				if !ok {
					continue
				}

				// Dynamic branch mapping based on Coze calcPortId rules
				// Collect all non-default branches and sort by level
				type BranchInfo struct {
					CaseID string
					Level  int
				}

				var nonDefaultBranches []BranchInfo
				var isDefaultBranch bool

				for _, caseItem := range conditionConfig.Cases {
					// Check if current branchID is a default branch
					if strings.Contains(caseItem.CaseID, branchID) && caseItem.Level == 999 {
						isDefaultBranch = true
						break
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

				// Sort non-default branches by level (ascending)
				for i := 0; i < len(nonDefaultBranches)-1; i++ {
					for j := i + 1; j < len(nonDefaultBranches); j++ {
						if nonDefaultBranches[i].Level > nonDefaultBranches[j].Level {
							nonDefaultBranches[i], nonDefaultBranches[j] = nonDefaultBranches[j], nonDefaultBranches[i]
						}
					}
				}

				// Generate port ID based on Coze calcPortId rules
				// index=0 -> "true", index=1 -> "true_1", index=N -> "true_N"
				for index, branch := range nonDefaultBranches {
					if strings.Contains(branch.CaseID, branchID) {
						if index == 0 {
							return "true"
						} else {
							return fmt.Sprintf("true_%d", index)
						}
					}
				}
			}
		}
	}

	// Fallback mapping - return empty for unrecognized branches
	return ""
}

// mapIntentToCozePort converts iFlytek intent handles to Coze classifier port format.
func (g *EdgeGenerator) mapIntentToCozePort(intentHandle string) string {
	// intentHandle is already in format: intent-one-of::uuid
	if intentHandle == "" {
		return ""
	}

	// Look up the intent in unified DSL for dynamic mapping
	if g.unifiedDSL != nil {
		for _, node := range g.unifiedDSL.Workflow.Nodes {
			if node.Type == models.NodeTypeClassifier {
				classifierConfig, ok := node.Config.(*models.ClassifierConfig)
				if !ok {
					continue
				}

				// Find the intent by matching intentHandle with class IDs
				// class.ID contains the full "intent-one-of::uuid" format
				for i, class := range classifierConfig.Classes {
					if class.ID == intentHandle {
						// Handle default intent specially
						if class.IsDefault {
							return "default"
						}
						// Map non-default intents to branch_0, branch_1, etc.
						return fmt.Sprintf("branch_%d", i)
					}
				}
			}
		}
	}

	// Fallback: if no match found in DSL, return empty
	return ""
}
