package golden

import (
	"ai-agents-transformer/internal/models"
	"time"
)

// GetIFlytekToUnified_BasicStartEnd returns the unified DSL for basic start-end workflow parsed from iFlytek
func GetIFlytekToUnified_BasicStartEnd() *models.UnifiedDSL {
	return &models.UnifiedDSL{
		Version: "1.0.0",
		Metadata: models.Metadata{
			Name: "自定义17520250916140208",
			Description: "",
			CreatedAt: time.Date(2025, 9, 18, 15, 3, 52, 196205700, time.UTC),
			UpdatedAt: time.Date(2025, 9, 18, 15, 3, 52, 196205700, time.UTC),
			UIConfig: &models.UIConfig{
				Icon: "https://oss-beijing-m8.openstorage.cn/SparkBotProd/icon/common/emojiitem_00_10@2x.png",
				IconBackground: "#FFEAD5",
			},
		},
		PlatformMetadata: models.PlatformMetadata{
			IFlytek: &models.IFlytekMetadata{
				AvatarIcon: "https://oss-beijing-m8.openstorage.cn/SparkBotProd/icon/common/emojiitem_00_10@2x.png",
				AvatarColor: "#FFEAD5",
				AdvancedConfig: "{\"prologue\":{\"enabled\":true,\"inputExample\":[\"\",\"\",\"\"]},\"needGuide\":false}",
				DSLVersion: "v1",
			},
		},
		Workflow: models.Workflow{
			Nodes: []models.Node{
				{
					ID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
					Type: models.NodeTypeStart,
					Title: "开始",
					Description: "工作流的开启节点，用于定义流程调用所需的业务变量信息。",
					Position: models.Position{X: -208.315916, Y: 501.794874},
					Size: models.Size{Width: 658.000000, Height: 416.000000},
					Inputs: []models.Input{
					},
					Outputs: []models.Output{
						{
							Name: "AGENT_USER_INPUT",
							Type: models.DataTypeString,
							Required: true,
							Default: "用户本轮对话输入内容",
						},
						{
							Name: "input_01",
							Type: models.DataTypeString,
							Required: true,
							Default: "学习内容",
						},
						{
							Name: "input_num_01",
							Type: models.DataTypeString,
							Required: true,
							Default: "难度级别(1-10)",
						},
						{
							Name: "input_num_02",
							Type: models.DataTypeString,
							Required: true,
							Default: "学习时间(小时)",
						},
						{
							Name: "input_text_01",
							Type: models.DataTypeString,
							Required: true,
							Default: "学习目标",
						},
					},
					Config: models.StartConfig{
						Variables: []models.Variable{
							{
								Name: "AGENT_USER_INPUT",
								Label: "用户本轮对话输入内容",
								Type: "string",
								Required: true,
								Default: "用户本轮对话输入内容",
							},
							{
								Name: "input_01",
								Label: "学习内容",
								Type: "string",
								Required: true,
								Default: "学习内容",
							},
							{
								Name: "input_num_01",
								Label: "难度级别(1-10)",
								Type: "string",
								Required: true,
								Default: "难度级别(1-10)",
							},
							{
								Name: "input_num_02",
								Label: "学习时间(小时)",
								Type: "string",
								Required: true,
								Default: "学习时间(小时)",
							},
							{
								Name: "input_text_01",
								Label: "学习目标",
								Type: "string",
								Required: true,
								Default: "学习目标",
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID: "node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					Type: models.NodeTypeEnd,
					Title: "结束",
					Description: "工作流的结束节点，用于输出工作流运行后的最终结果。",
					Position: models.Position{X: 627.982515, Y: 328.298326},
					Size: models.Size{Width: 408.000000, Height: 760.000000},
					Inputs: []models.Input{
						{
							Name: "result1",
							Type: models.DataTypeString,
							Required: true,
							Reference: &models.VariableReference{
								Type: models.ReferenceTypeNodeOutput,
								NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
								OutputName: "input_01",
								DataType: models.DataTypeString,
							},
						},
						{
							Name: "result2",
							Type: models.DataTypeString,
							Required: true,
							Reference: &models.VariableReference{
								Type: models.ReferenceTypeNodeOutput,
								NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
								OutputName: "input_num_01",
								DataType: models.DataTypeString,
							},
						},
						{
							Name: "result3",
							Type: models.DataTypeString,
							Required: true,
							Reference: &models.VariableReference{
								Type: models.ReferenceTypeNodeOutput,
								NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
								OutputName: "input_num_02",
								DataType: models.DataTypeString,
							},
						},
						{
							Name: "result4",
							Type: models.DataTypeString,
							Required: true,
							Reference: &models.VariableReference{
								Type: models.ReferenceTypeNodeOutput,
								NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
								OutputName: "input_text_01",
								DataType: models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{
					},
					Config: models.EndConfig{
						OutputMode: "template",
						Template: "{{result1}}\n{{result2}}\n{{result3}}\n{{result4}}",
						StreamOutput: true,
						Outputs: []models.EndOutput{
							{
								Variable: "result1",
								ValueSelector: []string{"node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783", "input_01"},
								ValueType: models.DataTypeString,
								Reference: &models.VariableReference{
									Type: models.ReferenceTypeNodeOutput,
									NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
									OutputName: "input_01",
									DataType: models.DataTypeString,
								},
							},
							{
								Variable: "result2",
								ValueSelector: []string{"node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783", "input_num_01"},
								ValueType: models.DataTypeString,
								Reference: &models.VariableReference{
									Type: models.ReferenceTypeNodeOutput,
									NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
									OutputName: "input_num_01",
									DataType: models.DataTypeString,
								},
							},
							{
								Variable: "result3",
								ValueSelector: []string{"node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783", "input_num_02"},
								ValueType: models.DataTypeString,
								Reference: &models.VariableReference{
									Type: models.ReferenceTypeNodeOutput,
									NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
									OutputName: "input_num_02",
									DataType: models.DataTypeString,
								},
							},
							{
								Variable: "result4",
								ValueSelector: []string{"node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783", "input_text_01"},
								ValueType: models.DataTypeString,
								Reference: &models.VariableReference{
									Type: models.ReferenceTypeNodeOutput,
									NodeID: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
									OutputName: "input_text_01",
									DataType: models.DataTypeString,
								},
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Edges: []models.Edge{
				{
					ID: "reactflow__edge-node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783-node-end::cda617af-551e-462e-b3b8-3bb9a041bf88node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					Source: "node-start::d61b0f71-87ee-475e-93ba-f1607f0ce783",
					Target: "node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					TargetHandle: "node-end::cda617af-551e-462e-b3b8-3bb9a041bf88",
					Type: models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Variables: []models.Variable{},
		},
	}
}
