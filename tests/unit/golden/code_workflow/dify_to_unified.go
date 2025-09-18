package golden

import (
	"ai-agents-transformer/internal/models"
	"time"
)

// GetDifyToUnified_Code_workflow returns the unified DSL for code workflow workflow parsed from dify
func GetDifyToUnified_Code_workflow() *models.UnifiedDSL {
	return &models.UnifiedDSL{
		Version: "1.0",
		Metadata: models.Metadata{
			Name: "a",
			Description: "",
			CreatedAt: time.Date(2025, 9, 18, 16, 11, 39, 224315600, time.UTC),
			UpdatedAt: time.Date(2025, 9, 18, 16, 11, 39, 224315600, time.UTC),
			UIConfig: &models.UIConfig{
				Icon: "🤖",
				IconBackground: "#FFEAD5",
			},
		},
		PlatformMetadata: models.PlatformMetadata{
		},
		Workflow: models.Workflow{
			Nodes: []models.Node{
				{
					ID: "1758003239028",
					Type: models.NodeTypeStart,
					Title: "开始",
					Description: "",
					Position: models.Position{X: 80.000000, Y: 282.000000},
					Size: models.Size{Width: 244.000000, Height: 89.000000},
					Inputs: []models.Input{
					},
					Outputs: []models.Output{
						{
							Name: "name",
							Type: models.DataTypeString,
							Required: false,
						},
					},
					Config: models.StartConfig{
						Variables: []models.Variable{
							{
								Name: "name",
								Label: "name",
								Type: "string",
								Required: true,
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID: "1758003261413",
					Type: models.NodeTypeEnd,
					Title: "结束",
					Description: "",
					Position: models.Position{X: 718.000000, Y: 250.000000},
					Size: models.Size{Width: 244.000000, Height: 89.000000},
					Inputs: []models.Input{
						{
							Name: "result",
							Type: models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type: models.ReferenceTypeNodeOutput,
								NodeID: "1758003291726",
								OutputName: "result",
								DataType: models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{
					},
					Config: models.EndConfig{
						OutputMode: "variables",
						Template: "",
						StreamOutput: false,
						Outputs: []models.EndOutput{
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID: "1758003291726",
					Type: models.NodeTypeCode,
					Title: "代码执行",
					Description: "",
					Position: models.Position{X: 399.000000, Y: 258.000000},
					Size: models.Size{Width: 244.000000, Height: 53.000000},
					Inputs: []models.Input{
						{
							Name: "name",
							Type: models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type: models.ReferenceTypeNodeOutput,
								NodeID: "1758003239028",
								OutputName: "name",
								DataType: models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{
						{
							Name: "result",
							Type: models.DataTypeArrayString,
							Required: false,
						},
					},
					Config: models.CodeConfig{
						Language: "python3",
						Code: "\ndef main(name: str) -> dict:\n    # 分析编程学习内容，生成学习路径\n    if \"python\" in name.lower():\n        path = [\"基础语法\", \"数据结构\", \"函数编程\", \"项目实战\"]\n    elif \"javascript\" in name.lower():\n        path = [\"基础语法\", \"DOM操作\", \"异步编程\", \"框架学习\"]\n    else:\n        path = [\"基础概念\", \"核心语法\", \"实践练习\", \"项目应用\"]\n    \n    return{\n        \"result\": path\n    }",
					},
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Edges: []models.Edge{
				{
					ID: "1758003239028-source-1758003291726-target",
					Source: "1758003239028",
					Target: "1758003291726",
					TargetHandle: "target",
					Type: models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID: "1758003291726-source-1758003261413-target",
					Source: "1758003291726",
					Target: "1758003261413",
					TargetHandle: "target",
					Type: models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Variables: []models.Variable{},
		},
	}
}
