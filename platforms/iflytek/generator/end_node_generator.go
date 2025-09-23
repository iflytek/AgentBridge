package generator

import (
	"agentbridge/internal/models"
	"agentbridge/platforms/common"
	"fmt"
)

// EndNodeGenerator handles end node generation
type EndNodeGenerator struct {
	*BaseNodeGenerator
	idMapping        map[string]string // Dify ID to iFlytek SparkAgent ID mapping
	nodeTitleMapping map[string]string // iFlytek SparkAgent ID to node title mapping
}

func NewEndNodeGenerator() *EndNodeGenerator {
	return &EndNodeGenerator{
		BaseNodeGenerator: NewBaseNodeGenerator(models.NodeTypeEnd),
		idMapping:         make(map[string]string),
		nodeTitleMapping:  make(map[string]string),
	}
}

// SetIDMapping sets ID mapping
func (g *EndNodeGenerator) SetIDMapping(idMapping map[string]string) {
	g.idMapping = idMapping
}

// SetNodeTitleMapping sets node title mapping
func (g *EndNodeGenerator) SetNodeTitleMapping(nodeTitleMapping map[string]string) {
	g.nodeTitleMapping = nodeTitleMapping
}

// GetSupportedNodeType returns supported node type
func (g *EndNodeGenerator) GetSupportedNodeType() models.NodeType {
	return models.NodeTypeEnd
}

// GenerateNode generates iFlytek SparkAgent end node
func (g *EndNodeGenerator) GenerateNode(node models.Node) (IFlytekNode, error) {
	iflytekNode := g.generateBasicNodeInfo(node)
	iflytekNode.Type = "结束节点"
	iflytekNode.Data.Label = "结束"
	iflytekNode.Data.Description = "工作流的结束节点，用于输出工作流运行后的最终结果。"
	iflytekNode.Data.NodeMeta.AliasName = "结束节点"
	iflytekNode.Data.NodeMeta.NodeType = "基础节点"

	// end node specific configuration
	iflytekNode.Data.AllowInputReference = true
	iflytekNode.Data.AllowOutputReference = false
	iflytekNode.Data.Icon = g.getNodeIcon(models.NodeTypeEnd)

	// parse end node configuration
	if endConfig, ok := common.AsEndConfig(node.Config); ok && endConfig != nil {
		// generate nodeParam configuration
		nodeParam := map[string]interface{}{
			"templateErrMsg": "",
			"outputMode":     g.convertOutputMode(endConfig.OutputMode),
			"streamOutput":   endConfig.StreamOutput,
		}

		// set template content if available
		if endConfig.Template != "" {
			nodeParam["template"] = endConfig.Template
		} else {
			// generate default template based on input variables
			template := g.generateDefaultTemplate(node.Inputs)
			nodeParam["template"] = template
		}

		iflytekNode.Data.NodeParam = nodeParam
	}

	// generate inputs (end node receives data through inputs)
	iflytekNode.Data.Inputs = g.generateInputsWithMapping(node.Inputs)

	// generate variable reference information
	iflytekNode.Data.References = g.generateReferences(node.Inputs)

	// end node has no outputs
	iflytekNode.Data.Outputs = []IFlytekOutput{}

	// restore iFlytek SparkAgent specific fields from platform config
	if iflytekConfig := node.PlatformConfig.IFlytek; iflytekConfig != nil {
		g.restoreIFlytekPlatformConfig(iflytekConfig, &iflytekNode)
	}

	return iflytekNode, nil
}

// ValidateNode validates unified DSL end node
func (g *EndNodeGenerator) ValidateNode(node models.Node) error {
	if node.Type != models.NodeTypeEnd {
		return fmt.Errorf("node type must be 'end', got '%s'", node.Type)
	}

	// end node should have inputs
	if len(node.Inputs) == 0 {
		return fmt.Errorf("end node should have at least one input")
	}

	return nil
}

// convertOutputMode converts output mode
func (g *EndNodeGenerator) convertOutputMode(mode string) int {
	switch mode {
	case "template":
		return 1 // template mode
	case "variables":
		return 0 // variable mode
	default:
		return 0 // default variable mode
	}
}

// generateDefaultTemplate generates default template
func (g *EndNodeGenerator) generateDefaultTemplate(inputs []models.Input) string {
	if len(inputs) == 0 {
		return ""
	}

	template := ""
	for i, input := range inputs {
		if i > 0 {
			template += "\n"
		}
		template += "{{" + input.Name + "}}"
	}

	return template
}

// generateReferences generates variable reference information
func (g *EndNodeGenerator) generateReferences(inputs []models.Input) []IFlytekReference {
	// group inputs by node ID
	nodeGroups := make(map[string][]models.Input)

	for _, input := range inputs {
		if input.Reference == nil || input.Reference.NodeID == "" {
			continue
		}

		// get mapped node ID
		mappedNodeID := input.Reference.NodeID
		if g.idMapping != nil {
			if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
				mappedNodeID = mapped
			}
		}

		nodeGroups[mappedNodeID] = append(nodeGroups[mappedNodeID], input)
	}

	// create parent reference for each node
	references := make([]IFlytekReference, 0, len(nodeGroups))

	for nodeID, nodeInputs := range nodeGroups {
		// create all output references for this node
		refDetails := make([]IFlytekRefDetail, 0, len(nodeInputs))

		for _, input := range nodeInputs {
			refDetail := IFlytekRefDetail{
				OriginID: nodeID,
				ID:       g.generateRefID(),
				Label:    input.Reference.OutputName,
				Type:     g.convertDataType(input.Type),
				Value:    input.Reference.OutputName,
				FileType: "",
				Children: []IFlytekReference{},
			}
			refDetails = append(refDetails, refDetail)
		}

		// create parent reference with flexible label retrieval
		ref := IFlytekReference{
			Label:      g.determineLabelByID(nodeID, g.nodeTitleMapping),
			ParentNode: true,
			Value:      nodeID,
			Children: []IFlytekReference{
				{
					Label:      "",
					Value:      "",
					References: refDetails,
				},
			},
		}

		references = append(references, ref)
	}

	return references
}

// getNodeLabelByID retrieves node label by ID
// Uses BaseNodeGenerator.determineLabelByID for unified processing

// generateInputsWithMapping generates inputs with ID mapping
func (g *EndNodeGenerator) generateInputsWithMapping(inputs []models.Input) []IFlytekInput {
	iflytekInputs := make([]IFlytekInput, 0, len(inputs))

	for _, input := range inputs {
		iflytekInput := IFlytekInput{
			ID:         g.generateRefID(),
			Name:       input.Name,
			NameErrMsg: "",
			Schema: IFlytekSchema{
				Type:    g.convertDataType(input.Type),
				Default: input.Default,
			},
			FileType: "",
		}

		// handle variable references with mapped ID
		if input.Reference != nil {
			// get mapped node ID
			mappedNodeID := input.Reference.NodeID
			if g.idMapping != nil {
				if mapped, exists := g.idMapping[input.Reference.NodeID]; exists {
					mappedNodeID = mapped
				}
			}

			iflytekInput.Schema.Value = &IFlytekSchemaValue{
				Type: "ref",
				Content: &IFlytekRefContent{
					Name:   input.Reference.OutputName,
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

// restoreIFlytekPlatformConfig restores iFlytek SparkAgent platform specific configuration
func (g *EndNodeGenerator) restoreIFlytekPlatformConfig(config map[string]interface{}, node *IFlytekNode) {
	// restore node label
	if label, ok := config["label"].(string); ok {
		node.Data.Label = label
	}

	// restore node description
	if description, ok := config["description"].(string); ok {
		node.Data.Description = description
	}

	// restore other node specific configuration
	if icon, ok := config["icon"].(string); ok {
		node.Data.Icon = icon
	}
}
