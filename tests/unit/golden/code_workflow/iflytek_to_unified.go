package golden

import (
	"agentbridge/internal/models"
	"time"
)

// GetIflytekToUnified_Code_workflow returns the unified DSL for code workflow workflow parsed from iflytek
func GetIflytekToUnified_Code_workflow() *models.UnifiedDSL {
	return &models.UnifiedDSL{
		Version: "1.0.0",
		Metadata: models.Metadata{
			Name:        "自定义17520250916135920",
			Description: "",
			CreatedAt:   time.Date(2025, 9, 18, 16, 11, 2, 724978200, time.UTC),
			UpdatedAt:   time.Date(2025, 9, 18, 16, 11, 2, 724978200, time.UTC),
			UIConfig: &models.UIConfig{
				Icon:           "https://oss-beijing-m8.openstorage.cn/SparkBotProd/icon/common/emojiitem_00_10@2x.png",
				IconBackground: "#FFEAD5",
			},
		},
		PlatformMetadata: models.PlatformMetadata{
			IFlytek: &models.IFlytekMetadata{
				AvatarIcon:     "https://oss-beijing-m8.openstorage.cn/SparkBotProd/icon/common/emojiitem_00_10@2x.png",
				AvatarColor:    "#FFEAD5",
				AdvancedConfig: "{\"prologue\":{\"enabled\":true,\"inputExample\":[\"\",\"\",\"\"]},\"needGuide\":false}",
				DSLVersion:     "v1",
			},
		},
		Workflow: models.Workflow{
			Nodes: []models.Node{
				{
					ID:          "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
					Type:        models.NodeTypeStart,
					Title:       "开始",
					Description: "工作流的开启节点，用于定义流程调用所需的业务变量信息。",
					Position:    models.Position{X: -296.467493, Y: -76.938026},
					Size:        models.Size{Width: 658.000000, Height: 313.000000},
					Inputs:      []models.Input{},
					Outputs: []models.Output{
						{
							Name:     "AGENT_USER_INPUT",
							Type:     models.DataTypeString,
							Required: true,
							Default:  "用户本轮对话输入内容",
						},
						{
							Name:     "name",
							Type:     models.DataTypeString,
							Required: false,
							Default:  "学习内容",
						},
					},
					Config: models.StartConfig{
						Variables: []models.Variable{
							{
								Name:     "AGENT_USER_INPUT",
								Label:    "用户本轮对话输入内容",
								Type:     "string",
								Required: true,
								Default:  "用户本轮对话输入内容",
							},
							{
								Name:     "name",
								Label:    "学习内容",
								Type:     "string",
								Required: false,
								Default:  "学习内容",
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:          "node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					Type:        models.NodeTypeEnd,
					Title:       "结束",
					Description: "工作流的结束节点，用于输出工作流运行后的最终结果。",
					Position:    models.Position{X: 2187.545514, Y: -367.406638},
					Size:        models.Size{Width: 408.000000, Height: 656.000000},
					Inputs: []models.Input{
						{
							Name:     "result",
							Type:     models.DataTypeArrayString,
							Required: true,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e",
								OutputName: "result",
								DataType:   models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{},
					Config: models.EndConfig{
						OutputMode:   "template",
						Template:     "{{result}}\n",
						StreamOutput: true,
						Outputs: []models.EndOutput{
							{
								Variable:      "result",
								ValueSelector: []string{"ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e", "result"},
								ValueType:     models.DataTypeArrayString,
								Reference: &models.VariableReference{
									Type:       models.ReferenceTypeNodeOutput,
									NodeID:     "ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e",
									OutputName: "result",
									DataType:   models.DataTypeString,
								},
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:          "ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e",
					Type:        models.NodeTypeCode,
					Title:       "编程学习路径生成器",
					Description: "面向开发者提供代码开发能力，目前仅支持python语言，允许使用该节点已定义的变量作为参数传入，返回语句用于输出函数的结果",
					Position:    models.Position{X: 838.670198, Y: -202.398183},
					Size:        models.Size{Width: 587.000000, Height: 843.000000},
					Inputs: []models.Input{
						{
							Name:     "name",
							Type:     models.DataTypeString,
							Required: true,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
								OutputName: "name",
								DataType:   models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{
						{
							Name:     "result",
							Type:     models.DataTypeArrayString,
							Required: false,
							Default:  "",
						},
					},
					Config: models.CodeConfig{
						Language: "python3",
						Code:     "def main(name: str) -> dict:\n    # 分析编程学习内容，生成学习路径\n    if \"python\" in name.lower():\n        path = [\"基础语法\", \"数据结构\", \"函数编程\", \"项目实战\"]\n    elif \"javascript\" in name.lower():\n        path = [\"基础语法\", \"DOM操作\", \"异步编程\", \"框架学习\"]\n    else:\n        path = [\"基础概念\", \"核心语法\", \"实践练习\", \"项目应用\"]\n    \n    return{\n        \"result\": path\n    }",
					},
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Edges: []models.Edge{
				{
					ID:             "reactflow__edge-ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e-node-end::cda617af-551e-462e-b3b8-3bb9a041bf88node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					Source:         "ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e",
					Target:         "node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					TargetHandle:   "node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					Type:           models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:             "reactflow__edge-node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783-ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e",
					Source:         "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
					Target:         "ifly-code::83b0cd48-968b-4ade-a02a-75c4ed25c69e",
					TargetHandle:   "",
					Type:           models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Variables: []models.Variable{},
		},
	}
}
