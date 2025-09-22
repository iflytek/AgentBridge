package generator

import (
	"ai-agents-transformer/internal/models"
	"ai-agents-transformer/platforms/common"
	"strings"
)

// CodeNodeGenerator handles code node generation
type CodeNodeGenerator struct {
	*BaseNodeGenerator
	idMapping        map[string]string // Dify ID to iFlytek SparkAgent ID mapping
	nodeTitleMapping map[string]string // iFlytek SparkAgent ID to node title mapping
}

func NewCodeNodeGenerator() *CodeNodeGenerator {
	return &CodeNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeCode),
		idMapping:         make(map[string]string),
		nodeTitleMapping:  make(map[string]string),
	}
}

// SetIDMapping sets ID mapping
func (g *CodeNodeGenerator) SetIDMapping(idMapping map[string]string) {
	g.idMapping = idMapping
}

// SetNodeTitleMapping sets node title mapping
func (g *CodeNodeGenerator) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	g.nodeTitleMapping = nodeTitleMapping
}

// GenerateNode generates code node
func (g *CodeNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	// generate basic node information
	iflytekNode := g.generateBasicNodeInfo(node)
	iflytekNode.Type = "代码"

	// set node metadata
	iflytekNode.Data.NodeMeta = IFlytekNodeMeta{
		AliasName: "代码",
		NodeType:  "工具",
	}

	// set icon and description
	iflytekNode.Data.Icon = g.getNodeIcon(models.NodeTypeCode)
	iflytekNode.Data.Description = "面向开发者提供代码开发能力，目前仅支持python语言，允许使用该节点已定义的变量作为参数传入，返回语句用于输出函数的结果"

	// set input/output permissions
	iflytekNode.Data.AllowInputReference = true
	iflytekNode.Data.AllowOutputReference = true

	// generate inputs
	iflytekNode.Data.Inputs = g.generateInputsWithMapping(node.Inputs)

	// generate outputs
	iflytekNode.Data.Outputs = g.generateOutputs(node.Outputs)

	// generate nodeParam
	iflytekNode.Data.NodeParam = g.generateNodeParam(node)

	// generate references
	iflytekNode.Data.References = g.generateReferences(node.Inputs)

	// set other properties
	iflytekNode.Data.Status = ""
	iflytekNode.Data.Updatable = false

	return iflytekNode, nil
}

// generateNodeParam generates node parameters
func (g *CodeNodeGenerator) generateNodeParam(node models.Node) map[string]interface{} {
	nodeParam := map[string]interface{}{
		"uid":        "20718349453", // default uid
		"appId":      "12a0a7e2",    // default appId
		"codeErrMsg": "",
	}

	// extract code information from configuration
	if codeConfig, ok := common.AsCodeConfig(node.Config); ok && codeConfig != nil {
		nodeParam["code"] = codeConfig.Code
	} else {
		nodeParam["code"] = "def main() -> dict:\n    return {'result': ''}" // default code
	}

	return nodeParam
}

// generateInputsWithMapping generates inputs with ID mapping
func (g *CodeNodeGenerator) generateInputsWithMapping(inputs []models.Input) []IFlytekInput {
	var iflytekInputs []IFlytekInput

	for _, input := range inputs {
		iflytekInput := IFlytekInput{
			ID:         g.generateInputID(),
			Name:       input.Name,
			NameErrMsg: "",
			FileType:   "",
			Schema: IFlytekSchema{
				Type:       g.convertDataType(input.Type),
				Properties: []interface{}{},
			},
		}

		// set input value
		if input.Reference != nil {
			// get mapped node ID
			mappedNodeID := input.Reference.NodeID
			if g.idMapping != nil {
				if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
					mappedNodeID = mapped
				}
			}

			// Map output names for platform compatibility
			mappedOutputName := g.mapOutputNameForPlatform(input.Reference.OutputName, mappedNodeID)

			iflytekInput.Schema.Value = &IFlytekSchemaValue{
				Type: "ref",
				Content: &IFlytekRefContent{
					Name:   mappedOutputName,
					ID:     g.generateRefID(),
					NodeID: mappedNodeID,
				},
				ContentErrMsg: "",
			}
		}

		iflytekInputs = append(iflytekInputs, iflytekInput)
	}

	return iflytekInputs
}

// generateOutputs generates node outputs
func (g *CodeNodeGenerator) generateOutputs(outputs []models.Output) []IFlytekOutput {
	var iflytekOutputs []IFlytekOutput

	for _, output := range outputs {
		iflytekOutput := IFlytekOutput{
			ID:         g.generateOutputID(),
			Name:       output.Name,
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type:       g.convertDataType(output.Type),
				Properties: []interface{}{},
				Default:    "",
			},
		}

		iflytekOutputs = append(iflytekOutputs, iflytekOutput)
	}

	return iflytekOutputs
}

// generateReferences generates variable reference information
func (g *CodeNodeGenerator) generateReferences(inputs []models.Input) []IFlytekReference {
	nodeGroups := make(map[string][]models.Input)

	// group inputs by source node
	for _, input := range inputs {
		if input.Reference == nil || input.Reference.NodeID == "" {
			continue
		}

		// Get mapped node ID
		mappedNodeID := input.Reference.NodeID
		if g.idMapping != nil {
			if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
				mappedNodeID = mapped
			}
		}

		nodeGroups[mappedNodeID] = append(nodeGroups[mappedNodeID], input)
	}

	var references []IFlytekReference
	for nodeID, nodeInputs := range nodeGroups {
		// create parent reference for each source node
		var refDetails []IFlytekRefDetail
		for _, input := range nodeInputs {
			// Map output name for platform compatibility
			mappedOutputName := g.mapOutputNameForPlatform(input.Reference.OutputName, nodeID)

			refDetails = append(refDetails, IFlytekRefDetail{
				OriginID: nodeID,
				ID:       g.generateRefID(),
				Label:    mappedOutputName,
				Type:     g.convertDataType(input.Reference.DataType),
				Value:    mappedOutputName,
				FileType: "",
			})
		}

		reference := IFlytekReference{
			Children: []IFlytekReference{
				{
					References: refDetails,
					Label:      "",
					Value:      "",
				},
			},
			Label:      g.determineLabelByID(nodeID, g.nodeTitleMapping),
			ParentNode: true,
			Value:      nodeID,
		}

		references = append(references, reference)
	}

	return references
}

// mapOutputNameForPlatform maps output names for platform compatibility
func (g *CodeNodeGenerator) mapOutputNameForPlatform(outputName, nodeID string) string {
	// Check if it's an LLM node (spark-llm::)
	if strings.HasPrefix(nodeID, "spark-llm::") {
		// In Dify, LLM nodes output 'text', but in iFlytek they output 'output'
		if outputName == "text" {
			return "output"
		}
	}

	// For other node types, return original output name
	return outputName
}
