package parsers

import (
	"fmt"
	"github.com/iflytek/agentbridge/internal/models"
)

// ValidateParserResult_BasicStartEnd validates parser result for basic start end workflow
func ValidateParserResult_BasicStartEnd(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL should not be nil")
	}

	// Validate node count equals 2
	if len(unifiedDSL.Workflow.Nodes) != 2 {
		return fmt.Errorf("expected 2 nodes, got %d", len(unifiedDSL.Workflow.Nodes))
	}

	// Validate edge count equals 1
	if len(unifiedDSL.Workflow.Edges) != 1 {
		return fmt.Errorf("expected 1 edge, got %d", len(unifiedDSL.Workflow.Edges))
	}

	// Validate start node exists with output parameters
	var hasStartNode bool
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == "start" {
			hasStartNode = true
			if len(node.Outputs) == 0 {
				return fmt.Errorf("start node should have output parameters")
			}
		}
	}
	if !hasStartNode {
		return fmt.Errorf("start node not found")
	}

	// Validate end node exists with input parameters
	var hasEndNode bool
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == "end" {
			hasEndNode = true
			if len(node.Inputs) == 0 {
				return fmt.Errorf("end node should have input parameters")
			}
		}
	}
	if !hasEndNode {
		return fmt.Errorf("end node not found")
	}

	return nil
}

// ValidateParserResult_ClassifierWorkflow validates parser result for classifier workflow
func ValidateParserResult_ClassifierWorkflow(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL should not be nil")
	}

	// Validate node count equals 5
	if len(unifiedDSL.Workflow.Nodes) != 5 {
		return fmt.Errorf("expected 5 nodes, got %d", len(unifiedDSL.Workflow.Nodes))
	}

	// Validate classifier node exists
	var hasClassifierNode bool
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == "classifier" {
			hasClassifierNode = true
			break
		}
	}
	if !hasClassifierNode {
		return fmt.Errorf("classifier node not found")
	}

	// Validate branch connection count >= 3
	if len(unifiedDSL.Workflow.Edges) < 3 {
		return fmt.Errorf("expected at least 3 edges for classifier workflow, got %d", len(unifiedDSL.Workflow.Edges))
	}

	return nil
}

// ValidateParserResult_CodeWorkflow validates parser result for code workflow
func ValidateParserResult_CodeWorkflow(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL should not be nil")
	}

	// Validate node count equals 3
	if len(unifiedDSL.Workflow.Nodes) != 3 {
		return fmt.Errorf("expected 3 nodes, got %d", len(unifiedDSL.Workflow.Nodes))
	}

	// Validate code node exists
	var codeNode *models.Node
	for i := range unifiedDSL.Workflow.Nodes {
		if unifiedDSL.Workflow.Nodes[i].Type == "code" {
			codeNode = &unifiedDSL.Workflow.Nodes[i]
			break
		}
	}
	if codeNode == nil {
		return fmt.Errorf("code node not found")
	}

	// Validate code node has input parameters
	if len(codeNode.Inputs) == 0 {
		return fmt.Errorf("code node should have input parameters")
	}

	// Validate code node has output parameters
	if len(codeNode.Outputs) == 0 {
		return fmt.Errorf("code node should have output parameters")
	}

	return nil
}

// ValidateParserResult_ConditionWorkflow validates parser result for condition workflow
func ValidateParserResult_ConditionWorkflow(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL should not be nil")
	}

	// Validate node count equals 5
	if len(unifiedDSL.Workflow.Nodes) != 5 {
		return fmt.Errorf("expected 5 nodes, got %d", len(unifiedDSL.Workflow.Nodes))
	}

	// Validate edge count equals 6
	if len(unifiedDSL.Workflow.Edges) != 6 {
		return fmt.Errorf("expected 6 edges, got %d", len(unifiedDSL.Workflow.Edges))
	}

	// Validate condition node exists
	var hasConditionNode bool
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == "condition" {
			hasConditionNode = true
			break
		}
	}
	if !hasConditionNode {
		return fmt.Errorf("condition node not found")
	}

	return nil
}

// ValidateParserResult_IterationWorkflow validates parser result for iteration workflow
func ValidateParserResult_IterationWorkflow(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL should not be nil")
	}

	// Validate iteration node exists
	var iterationNode *models.Node
	for i := range unifiedDSL.Workflow.Nodes {
		if unifiedDSL.Workflow.Nodes[i].Type == "iteration" {
			iterationNode = &unifiedDSL.Workflow.Nodes[i]
			break
		}
	}
	if iterationNode == nil {
		return fmt.Errorf("iteration node not found")
	}

	// Validate iteration configuration is not empty
	if iterationNode.Config == nil {
		return fmt.Errorf("iteration node should have configuration")
	}

	return nil
}

// ValidateParserResult_LLMWorkflow validates parser result for LLM workflow
func ValidateParserResult_LLMWorkflow(unifiedDSL *models.UnifiedDSL) error {
	if unifiedDSL == nil {
		return fmt.Errorf("unified DSL should not be nil")
	}

	// Validate node count equals 3
	if len(unifiedDSL.Workflow.Nodes) != 3 {
		return fmt.Errorf("expected 3 nodes, got %d", len(unifiedDSL.Workflow.Nodes))
	}

	// Validate edge count equals 2
	if len(unifiedDSL.Workflow.Edges) != 2 {
		return fmt.Errorf("expected 2 edges, got %d", len(unifiedDSL.Workflow.Edges))
	}

	// Validate LLM node exists
	var hasLLMNode bool
	for _, node := range unifiedDSL.Workflow.Nodes {
		if node.Type == "llm" {
			hasLLMNode = true
			break
		}
	}
	if !hasLLMNode {
		return fmt.Errorf("llm node not found")
	}

	return nil
}
