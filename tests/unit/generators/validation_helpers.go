package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// PlatformStructure defines expected structure requirements for each platform.
type PlatformStructure struct {
	RequiredFields []string
	NodesPath      []string // Path to nodes array: ["schema", "nodes"]
	EdgesPath      []string // Path to edges array: ["schema", "edges"]
}

// GetPlatformStructure returns platform-specific structure requirements.
func GetPlatformStructure(platform string) PlatformStructure {
	switch platform {
	case "iflytek":
		return PlatformStructure{
			RequiredFields: []string{"flowMeta", "flowData"},
			NodesPath:      []string{"flowData", "nodes"},
			EdgesPath:      []string{"flowData", "edges"},
		}
	case "coze":
		return PlatformStructure{
			RequiredFields: []string{"schema"},
			NodesPath:      []string{"schema", "nodes"},
			EdgesPath:      []string{"schema", "edges"},
		}
	case "dify":
		return PlatformStructure{
			RequiredFields: []string{"workflow"},
			NodesPath:      []string{"workflow", "graph", "nodes"},
			EdgesPath:      []string{"workflow", "graph", "edges"},
		}
	default:
		return PlatformStructure{}
	}
}

// ValidationResult represents validation check result.
type ValidationResult struct {
	Passed  bool
	Message string
	Details map[string]interface{}
}

// ValidateConversionBasics performs basic conversion validation.
func ValidateConversionBasics(t *testing.T, conversionResult []byte, targetPlatform string) (map[string]interface{}, ValidationResult) {
	var result map[string]interface{}
	err := yaml.Unmarshal(conversionResult, &result)
	require.NoError(t, err, "Should be able to parse YAML result")

	structure := GetPlatformStructure(targetPlatform)
	structurePassed := true
	for _, field := range structure.RequiredFields {
		if _, exists := result[field]; !exists {
			structurePassed = false
			break
		}
	}

	assert.True(t, structurePassed, "Should contain required %s structure", targetPlatform)
	t.Logf("✅ Platform Required Structure: Passed=%t", structurePassed)

	return result, ValidationResult{
		Passed:  structurePassed,
		Message: "Platform structure validation",
		Details: map[string]interface{}{"platform": targetPlatform},
	}
}

// ValidateNodeCount validates node count consistency.
func ValidateNodeCount(t *testing.T, result map[string]interface{}, targetPlatform string, expectedCount int) ValidationResult {
	structure := GetPlatformStructure(targetPlatform)

	current := result
	for _, key := range structure.NodesPath {
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else if key == structure.NodesPath[len(structure.NodesPath)-1] {
			if nodes, ok := current[key].([]interface{}); ok {
				actualCount := len(nodes)
				nodeCountPassed := actualCount == expectedCount
				assert.True(t, nodeCountPassed, "Should have exactly %d nodes", expectedCount)
				t.Logf("✅ Node Count Consistency: Expected=%d, Actual=%d, Passed=%t",
					expectedCount, actualCount, nodeCountPassed)

				return ValidationResult{
					Passed:  nodeCountPassed,
					Message: "Node count validation",
					Details: map[string]interface{}{
						"expected": expectedCount,
						"actual":   actualCount,
						"nodes":    nodes,
					},
				}
			}
		}
	}

	t.Logf("❌ Node Count Consistency: Nodes array not found")
	return ValidationResult{
		Passed:  false,
		Message: "Nodes array not found",
		Details: map[string]interface{}{},
	}
}

// FindNodesByType finds nodes by type field matching.
func FindNodesByType(nodes []interface{}, nodeTypes map[string]string) map[string]map[string]interface{} {
	foundNodes := make(map[string]map[string]interface{})

	for _, node := range nodes {
		if nodeMap, ok := node.(map[string]interface{}); ok {
			var nodeType string

			if directType, ok := nodeMap["type"].(string); ok {
				nodeType = directType
				if directType == "custom" {
					if data, ok := nodeMap["data"].(map[string]interface{}); ok {
						if nestedType, ok := data["type"].(string); ok {
							nodeType = nestedType
						}
					}
				}
			} else if data, ok := nodeMap["data"].(map[string]interface{}); ok {
				if nestedType, ok := data["type"].(string); ok {
					nodeType = nestedType
				}
			}

			if nodeType != "" {
				for logicalType, expectedType := range nodeTypes {
					if nodeType == expectedType {
						foundNodes[logicalType] = nodeMap
					}
				}
			}
		}
	}

	return foundNodes
}

// ValidateNodeTypeMapping validates node type mapping accuracy.
func ValidateNodeTypeMapping(t *testing.T, nodes []interface{}, expectedNodeTypes map[string]string) ValidationResult {
	foundNodes := FindNodesByType(nodes, expectedNodeTypes)

	allFound := len(foundNodes) == len(expectedNodeTypes)
	assert.True(t, allFound, "Node type mapping should be correct")
	t.Logf("✅ Node Type Mapping Accuracy: Found=%d, Expected=%d, Passed=%t",
		len(foundNodes), len(expectedNodeTypes), allFound)

	return ValidationResult{
		Passed:  allFound,
		Message: "Node type mapping validation",
		Details: map[string]interface{}{
			"foundNodes":    foundNodes,
			"expectedTypes": expectedNodeTypes,
		},
	}
}

// CheckParameters validates parameter existence in parameter list.
func CheckParameters(paramList []interface{}, expectedParams []string) bool {
	for _, expected := range expectedParams {
		found := false
		for _, param := range paramList {
			if paramMap, ok := param.(map[string]interface{}); ok {
				if paramMap["name"] == expected {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// ValidateParameterPreservation validates node parameter preservation.
func ValidateParameterPreservation(t *testing.T, foundNodes map[string]map[string]interface{}, expectedParams map[string]map[string][]string) ValidationResult {
	allPassed := true

	for nodeType, node := range foundNodes {
		if expectedForNode, exists := expectedParams[nodeType]; exists {
			nodeData := node["data"].(map[string]interface{})

			if expectedOutputs, hasOutputs := expectedForNode["outputs"]; hasOutputs {
				if outputs, ok := nodeData["outputs"].([]interface{}); ok {
					if !CheckParameters(outputs, expectedOutputs) {
						allPassed = false
					}
				}
			}

			if expectedInputs, hasInputs := expectedForNode["inputs"]; hasInputs {
				if inputs, ok := nodeData["inputs"].([]interface{}); ok {
					if !CheckParameters(inputs, expectedInputs) {
						allPassed = false
					}
				}
			}
		}
	}

	assert.True(t, allPassed, "Parameter preservation should be correct")
	t.Logf("✅ Parameter Preservation Rate: Passed=%t", allPassed)

	return ValidationResult{
		Passed:  allPassed,
		Message: "Parameter preservation validation",
		Details: map[string]interface{}{},
	}
}

// HasNonEmptyCozeReference checks for non-empty blockID references in Coze format.
func HasNonEmptyCozeReference(inputs []interface{}) bool {
	for _, input := range inputs {
		inputMap := input.(map[string]interface{})

		if inputParameters, ok := inputMap["inputParameters"].([]interface{}); ok {
			for _, param := range inputParameters {
				if paramMap, ok := param.(map[string]interface{}); ok {
					if inputField, ok := paramMap["input"].(map[string]interface{}); ok {
						if value, ok := inputField["value"].(map[string]interface{}); ok {
							if value["type"] == "ref" {
								if content, ok := value["content"].(map[string]interface{}); ok {
									if blockID, exists := content["blockID"]; exists && blockID != nil {
										return true
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

// HasNonEmptyIFlytekReference checks for non-empty nodeId references in iFlytek format.
func HasNonEmptyIFlytekReference(inputs []interface{}) bool {
	for _, input := range inputs {
		inputMap := input.(map[string]interface{})

		schema, ok := inputMap["schema"].(map[string]interface{})
		if !ok {
			continue
		}

		value, ok := schema["value"].(map[string]interface{})
		if !ok {
			continue
		}

		if value["type"] != "ref" {
			continue
		}

		content, ok := value["content"].(map[string]interface{})
		if !ok {
			continue
		}

		nodeId, exists := content["nodeId"]
		if exists && nodeId != nil {
			if nodeIdStr, ok := nodeId.(string); ok && nodeIdStr != "" {
				return true
			}
		}
	}
	return false
}

// HasNonEmptyDifyReference checks for non-empty value_selector references in Dify format.
func HasNonEmptyDifyReference(inputs []interface{}) bool {
	for _, input := range inputs {
		inputMap := input.(map[string]interface{})

		if valueSelector, exists := inputMap["value_selector"]; exists && valueSelector != nil {
			if selectorArray, ok := valueSelector.([]interface{}); ok && len(selectorArray) > 0 {
				return true
			}
		}
	}
	return false
}

// ValidateReferenceValues validates platform-specific reference values.
func ValidateReferenceValues(t *testing.T, foundNodes map[string]map[string]interface{}, targetPlatform string) ValidationResult {
	referencesPassed := false

	endNodeTypes := []string{"end", "结束节点"}
	var endNode map[string]interface{}

	for nodeType, node := range foundNodes {
		for _, endType := range endNodeTypes {
			if nodeType == endType || node["type"] == endType {
				endNode = node
				break
			}
		}
		if endNode != nil {
			break
		}
	}

	if endNode != nil {
		endData := endNode["data"].(map[string]interface{})

		switch targetPlatform {
		case "iflytek":
			if inputs, ok := endData["inputs"].([]interface{}); ok {
				referencesPassed = HasNonEmptyIFlytekReference(inputs)
			}
		case "coze":
			if inputsMap, ok := endData["inputs"].(map[string]interface{}); ok {
				referencesPassed = HasNonEmptyCozeReference([]interface{}{inputsMap})
			}
		case "dify":
			if outputs, ok := endData["outputs"].([]interface{}); ok {
				referencesPassed = HasNonEmptyDifyReference(outputs)
			}
		}
	}

	assert.True(t, referencesPassed, "Reference values should be non-empty")
	t.Logf("✅ Reference Value Non-Empty: Passed=%t", referencesPassed)

	return ValidationResult{
		Passed:  referencesPassed,
		Message: "Reference validation",
		Details: map[string]interface{}{"platform": targetPlatform},
	}
}

// RunCompleteBasicStartEndValidation runs complete validation for basic start-end workflow.
func RunCompleteBasicStartEndValidation(t *testing.T, conversionResult []byte, targetPlatform string) bool {
	// Expected parameters for basic start-end workflow
	expectedNodeCount := 2
	expectedNodeTypes := map[string]string{
		"start": getStartNodeType(targetPlatform),
		"end":   getEndNodeType(targetPlatform),
	}
	expectedParams := map[string]map[string][]string{
		"start": {
			"outputs": {"input_01", "input_num_01", "input_num_02", "input_text_01"},
		},
		"end": {
			"inputs": {"result1", "result2", "result3", "result4"},
		},
	}

	// Run validation steps
	result, structureValidation := ValidateConversionBasics(t, conversionResult, targetPlatform)
	if !structureValidation.Passed {
		return false
	}

	nodeCountValidation := ValidateNodeCount(t, result, targetPlatform, expectedNodeCount)
	if !nodeCountValidation.Passed {
		return false
	}

	nodes := nodeCountValidation.Details["nodes"].([]interface{})

	typeValidation := ValidateNodeTypeMapping(t, nodes, expectedNodeTypes)
	if !typeValidation.Passed {
		return false
	}

	foundNodes := typeValidation.Details["foundNodes"].(map[string]map[string]interface{})
	paramValidation := ValidateParameterPreservation(t, foundNodes, expectedParams)
	if !paramValidation.Passed {
		return false
	}

	refValidation := ValidateReferenceValues(t, foundNodes, targetPlatform)
	if !refValidation.Passed {
		return false
	}

	return true
}

// getStartNodeType returns platform-specific start node type.
func getStartNodeType(platform string) string {
	switch platform {
	case "iflytek":
		return "开始节点"
	case "coze":
		return "1"
	case "dify":
		return "start"
	default:
		return "start"
	}
}

func getEndNodeType(platform string) string {
	switch platform {
	case "iflytek":
		return "结束节点"
	case "coze":
		return "2"
	case "dify":
		return "end"
	default:
		return "end"
	}
}

// getCodeNodeType returns platform-specific code node type.
func getCodeNodeType(platform string) string {
	switch platform {
	case "iflytek":
		return "代码"
	case "coze":
		return "5" // 修正Coze的代码节点类型
	case "dify":
		return "code"
	default:
		return "code"
	}
}

// RunCompleteCodeWorkflowValidation runs complete validation for code workflow.
func RunCompleteCodeWorkflowValidation(t *testing.T, conversionResult []byte, targetPlatform string) bool {
	// Expected parameters for code workflow
	expectedNodeCount := 3
	expectedNodeTypes := map[string]string{
		"start": getStartNodeType(targetPlatform),
		"code":  getCodeNodeType(targetPlatform),
		"end":   getEndNodeType(targetPlatform),
	}
	expectedParams := map[string]map[string][]string{
		"code": {
			"inputs":  {"name"},
			"outputs": {"result"},
		},
	}

	// Run validation steps
	result, structureValidation := ValidateConversionBasics(t, conversionResult, targetPlatform)
	if !structureValidation.Passed {
		return false
	}

	nodeCountValidation := ValidateNodeCount(t, result, targetPlatform, expectedNodeCount)
	if !nodeCountValidation.Passed {
		return false
	}

	nodes := nodeCountValidation.Details["nodes"].([]interface{})

	typeValidation := ValidateNodeTypeMapping(t, nodes, expectedNodeTypes)
	if !typeValidation.Passed {
		return false
	}

	foundNodes := typeValidation.Details["foundNodes"].(map[string]map[string]interface{})
	paramValidation := ValidateParameterPreservation(t, foundNodes, expectedParams)
	if !paramValidation.Passed {
		return false
	}

	codeContentValidation := ValidateCodeContent(t, foundNodes, targetPlatform)
	if !codeContentValidation.Passed {
		return false
	}

	return true
}

// ValidateCodeContent validates code node contains non-empty content.
func ValidateCodeContent(t *testing.T, foundNodes map[string]map[string]interface{}, targetPlatform string) ValidationResult {
	contentPassed := false

	codeNodeTypes := []string{"code", "代码", "4", "code"}
	var codeNode map[string]interface{}

	for nodeType, node := range foundNodes {
		for _, codeType := range codeNodeTypes {
			if nodeType == "code" || node["type"] == codeType {
				codeNode = node
				break
			}
		}
		if codeNode != nil {
			break
		}
	}

	if codeNode != nil {
		contentPassed = CheckCodeContent(codeNode, targetPlatform)
	}

	assert.True(t, contentPassed, "Code node should contain non-empty code content")
	t.Logf("✅ Code Content Validation: Passed=%t", contentPassed)

	return ValidationResult{
		Passed:  contentPassed,
		Message: "Code content validation",
		Details: map[string]interface{}{"platform": targetPlatform},
	}
}

// CheckCodeContent validates code node contains non-empty content.
func CheckCodeContent(codeNode map[string]interface{}, targetPlatform string) bool {
	codeData := codeNode["data"].(map[string]interface{})

	switch targetPlatform {
	case "iflytek":
		if nodeParam, ok := codeData["nodeParam"].(map[string]interface{}); ok {
			if code, ok := nodeParam["code"].(string); ok && code != "" {
				return true
			}
		}
	case "coze":
		if inputs, ok := codeData["inputs"].(map[string]interface{}); ok {
			if code, ok := inputs["code"].(string); ok && code != "" {
				return true
			}
		}
	case "dify":
		if code, ok := codeData["code"].(string); ok && code != "" {
			return true
		}
	}
	return false
}
