package golden

import (
	"github.com/iflytek/agentbridge/internal/models"
	"time"
)

// GetCozeToUnified_Code_workflow returns the unified DSL for code workflow workflow parsed from coze
func GetCozeToUnified_Code_workflow() *models.UnifiedDSL {
	return &models.UnifiedDSL{
		Version: "1.0",
		Metadata: models.Metadata{
			Name:        "a1",
			Description: "1",
			CreatedAt:   time.Date(2025, 9, 16, 14, 7, 56, 0, time.UTC),
			UpdatedAt:   time.Date(2025, 9, 16, 14, 7, 56, 0, time.UTC),
		},
		PlatformMetadata: models.PlatformMetadata{},
		Workflow: models.Workflow{
			Nodes: []models.Node{
				{
					ID:          "100001",
					Type:        models.NodeTypeStart,
					Title:       "Start",
					Description: "The starting node of the workflow, used to set the information needed to initiate the workflow.",
					Position:    models.Position{X: 0.000000, Y: 0.000000},
					Size:        models.Size{Width: 244.000000, Height: 118.000000},
					Inputs:      []models.Input{},
					Outputs: []models.Output{
						{
							Name:     "name",
							Type:     models.DataTypeString,
							Required: false,
						},
					},
					Config: models.StartConfig{
						Variables: []models.Variable{
							{
								Name:     "name",
								Label:    "name",
								Type:     "string",
								Required: false,
								Default:  "",
							},
						},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:          "900001",
					Type:        models.NodeTypeEnd,
					Title:       "End",
					Description: "The final node of the workflow, used to return the result information after the workflow runs.",
					Position:    models.Position{X: 1000.000000, Y: 0.000000},
					Size:        models.Size{Width: 244.000000, Height: 118.000000},
					Inputs: []models.Input{
						{
							Name:     "output",
							Type:     models.DataTypeString,
							Required: false,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "185345",
								OutputName: "result",
								DataType:   models.DataTypeString,
							},
						},
					},
					Outputs: []models.Output{},
					Config: models.EndConfig{
						OutputMode:   "variables",
						Template:     "{{output}}",
						StreamOutput: true,
						Outputs:      []models.EndOutput{},
					},
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:          "185345",
					Type:        models.NodeTypeCode,
					Title:       "Code",
					Description: "Write code to process input variables to generate return values.",
					Position:    models.Position{X: 554.634870, Y: -44.933356},
					Size:        models.Size{Width: 244.000000, Height: 118.000000},
					Inputs: []models.Input{
						{
							Name:     "name",
							Type:     models.DataTypeString,
							Required: true,
							Reference: &models.VariableReference{
								Type:       models.ReferenceTypeNodeOutput,
								NodeID:     "100001",
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
						},
					},
					Config: models.CodeConfig{
						Language: "python3",
						Code:     "def main(name:str) -> dict:\n    # Here, you can use 'args' to access the input variables in the node and use 'ret' to output the result\n    # 'args'  have already been correctly injected into the environment\n    # Below is an example: First, retrieve all input parameters (params) from the node, then get the value of the parameter 'input':\n    # input = params.input;\n    # Below is an example of outputting a 'ret' object containing multiple data types:\n    # ret: Output =  { \"name\": 'Xiao Ming', \"hobbies\": [\"reading\", \"traveling\"] };\n\n    name = params.get(\"input\", \"\")  # 获取输入参数\n\n    # 分析编程学习内容，生成学习路径\n    if \"python\" in name.lower():\n        path = [\"基础语法\", \"数据结构\", \"函数编程\", \"项目实战\"]\n    elif \"javascript\" in name.lower():\n        path = [\"基础语法\", \"DOM操作\", \"异步编程\", \"框架学习\"]\n    else:\n        path = [\"基础概念\", \"核心语法\", \"实践练习\", \"项目应用\"]\n\n    # 构造返回对象\n    ret: Output = {\n        \"result\": path\n    }\n    return ret\n",
					},
					PlatformConfig: models.PlatformConfig{},
				},
			},
			Edges: []models.Edge{
				{
					ID:             "edge-100001-185345",
					Source:         "100001",
					Target:         "185345",
					TargetHandle:   "",
					Type:           models.EdgeTypeDefault,
					PlatformConfig: models.PlatformConfig{},
				},
				{
					ID:             "edge-185345-900001",
					Source:         "185345",
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
