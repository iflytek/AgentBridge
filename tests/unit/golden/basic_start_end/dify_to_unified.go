package golden

import (
	"agentbridge/internal/models"
	"time"
)

// GetDifyToUnified_Basic_start_end returns the unified DSL for basic start end workflow parsed from dify
func GetDifyToUnified_Basic_start_end() *models.UnifiedDSL {
	return &models.UnifiedDSL{
		Version: "1.0",
		Metadata: models.Metadata{
			Name:        "智能学习助手",
			Description: "智能学习助手，可以根据用户输入进行问题分类、条件判断、代码处理等多种学习辅助功能",
			CreatedAt:   time.Date(2025, 9, 18, 15, 33, 52, 476952300, time.UTC),
			UpdatedAt:   time.Date(2025, 9, 18, 15, 33, 52, 476952300, time.UTC),
			UIConfig: &models.UIConfig{
				Icon:           "📚",
				IconBackground: "#E8F5E8",
			},
		},
		PlatformMetadata: models.PlatformMetadata{},
		Workflow: models.Workflow{
			Nodes: []models.Node{
				{
					ID:          "1754269219469",
					Type:        models.NodeTypeStart,
					Title:       "学习需求输入",
					Description: "用户输入学习需求和相关信息",
					Position:    models.Position{X: 208.612816, Y: 284.143547},
					Size:        models.Size{Width: 243.000000, Height: 195.000000},
					Inputs:      []models.Input{},
					Outputs: []models.Output{
						{
							Name:     "input_01",
							Type:     models.DataTypeString,
							Required: false,
						},
						{
							Name:     "input_num_01",
							Type:     models.DataTypeNumber,
							Required: false,
						},
						{
							Name:     "input_num_02",
							Type:     models.DataTypeNumber,
							Required: false,
						},
						{
							Name:     "input_text_01",
							Type:     models.DataTypeString,
							Required: false,
						},
					},
					Config: models.StartConfig{
						Variables: []models.Variable{
							{
								Name:     "input_01",
								Label:    "学习内容",
								Type:     "string",
								Required: true,
							},
							{
								Name:     "input_num_01",
								Label:    "难度级别(1-10)",
								Type:     "number",
								Required: true,
							},
							{
								Name:     "input_num_02",
								Label:    "学习时间(小时)",
								Type:     "number",
								Required: true,
							},
							{
								Name:     "input_text_01",
								Label:    "学习目标",
								Type:     "string",
								Required: true,
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:          "1754269231715",
					Type:        models.NodeTypeEnd,
					Title:       "学习方案输出",
					Description: "输出最终的学习建议和方案",
					Position:    models.Position{X: 567.879520, Y: 284.143547},
					Size:        models.Size{Width: 243.000000, Height: 195.000000},
					Inputs: []models.Input{
						{
							Name:     "result1",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "1754269219469",
								OutputName: "input_01",
								DataType:   models.DataTypeString,
							},
						},
						{
							Name:     "result2",
							Type:     models.DataTypeNumber,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "1754269219469",
								OutputName: "input_num_01",
								DataType:   models.DataTypeNumber,
							},
						},
						{
							Name:     "result3",
							Type:     models.DataTypeNumber,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "1754269219469",
								OutputName: "input_num_02",
								DataType:   models.DataTypeNumber,
							},
						},
						{
							Name:     "result4",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "1754269219469",
								OutputName: "input_text_01",
								DataType:   models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{},
					Config: models.EndConfig{
						OutputMode:   "variables",
						Template:     "",
						StreamOutput: false,
						Outputs:      []models.EndOutput{},
					},
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Edges: []models.Edge{
				{
					ID:             "1754269219469-source-1754269231715-target",
					Source:         "1754269219469",
					Target:         "1754269231715",
					TargetHandle:   "target",
					Type:           models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Variables: []models.Variable{},
		},
	}
}
