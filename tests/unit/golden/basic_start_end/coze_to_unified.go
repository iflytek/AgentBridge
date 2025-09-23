package golden

import (
	"agentbridge/internal/models"
	"time"
)

// GetCozeToUnified_Basic_start_end returns the unified DSL for basic start end workflow parsed from coze
func GetCozeToUnified_Basic_start_end() *models.UnifiedDSL {
	return &models.UnifiedDSL{
		Version: "1.0",
		Metadata: models.Metadata{
			Name:        "coze________",
			Description: "111",
			CreatedAt:   time.Date(2025, 9, 16, 13, 57, 52, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 9, 16, 14, 2, 59, 0, time.UTC),
		},
		PlatformMetadata: models.PlatformMetadata{},
		Workflow: models.Workflow{
			Nodes: []models.Node{
				{
					ID:          "100001",
					Type:        models.NodeTypeStart,
					Title:       "开始",
					Description: "工作流的起始节点，用于设定启动工作流需要的信息",
					Position:    models.Position{X: 0.000000, Y: 0.000000},
					Size:        models.Size{Width: 244.000000, Height: 118.000000},
					Inputs:      []models.Input{},
					Outputs: []models.Output{
						{
							Name:     "input_01",
							Type:     models.DataTypeString,
							Required: true,
						},
						{
							Name:     "input_num_01",
							Type:     models.DataTypeString,
							Required: true,
						},
						{
							Name:     "input_num_02",
							Type:     models.DataTypeString,
							Required: true,
						},
						{
							Name:     "input_text_01",
							Type:     models.DataTypeString,
							Required: true,
						},
					},
					Config: models.StartConfig{
						Variables: []models.Variable{
							{
								Name:     "input_01",
								Label:    "input_01",
								Type:     "string",
								Required: true,
							},
							{
								Name:     "input_num_01",
								Label:    "input_num_01",
								Type:     "string",
								Required: true,
							},
							{
								Name:     "input_num_02",
								Label:    "input_num_02",
								Type:     "string",
								Required: true,
							},
							{
								Name:     "input_text_01",
								Label:    "input_text_01",
								Type:     "string",
								Required: true,
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:          "900001",
					Type:        models.NodeTypeEnd,
					Title:       "结束",
					Description: "工作流的最终节点，用于返回工作流运行后的结果信息",
					Position:    models.Position{X: 545.186981, Y: -13.000000},
					Size:        models.Size{Width: 244.000000, Height: 118.000000},
					Inputs: []models.Input{
						{
							Name:     "result1",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "100001",
								OutputName: "input_01",
								DataType:   models.DataTypeString,
							},
						},
						{
							Name:     "result2",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "100001",
								OutputName: "input_num_01",
								DataType:   models.DataTypeString,
							},
						},
						{
							Name:     "result3",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "100001",
								OutputName: "input_num_02",
								DataType:   models.DataTypeString,
							},
						},
						{
							Name:     "result4",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "100001",
								OutputName: "input_text_01",
								DataType:   models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{},
					Config: models.EndConfig{
						OutputMode:   "variables",
						Template:     "{{result1}}, {{result2}}, {{result3}}, {{result4}}",
						StreamOutput: true,
						Outputs:      []models.EndOutput{},
					},
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Edges: []models.Edge{
				{
					ID:             "edge-100001-900001",
					Source:         "100001",
					Target:         "900001",
					TargetHandle:   "",
					Type:           models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Variables: []models.Variable{},
		},
	}
}
