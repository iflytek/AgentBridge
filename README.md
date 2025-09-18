# AgentBridge CLI 用户手册

跨平台 AI 智能体工作流 DSL 转换器，以科大讯飞星辰为中枢，支持星辰与 Dify、Coze 平台双向转换。

## 转换架构

**星型架构设计**：以 iFlytek 星辰为中央枢纽，支持以下转换路径：

```
    Dify ←→ iFlytek ←→ Coze
           (中央枢纽)
```

- ✅ **iFlytek ↔ Dify**：完整双向转换
- ✅ **iFlytek ↔ Coze**：完整双向转换，支持 ZIP 格式
- ❌ **Dify ↔ Coze**：不支持直接转换，需通过 iFlytek 中转

## 快速开始

```bash
# 构建CLI工具
go build -o ai-agent-converter ./cmd/

# 基础转换：星辰转Coze
./ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

# Coze ZIP转星辰
./ai-agent-converter convert --from coze --to iflytek --input workflow.zip --output agent.yml

# 自动检测格式
./ai-agent-converter convert --to dify --input agent.yml --output dify.yml

# 验证DSL文件
./ai-agent-converter validate --input agent.yml
```

## 命令详解

### convert

在不同 AI 智能体平台间进行转换。

**用法：**

```bash
ai-agent-converter convert [flags]
```

**参数：**

- `--from` - 源平台 (iflytek|dify|coze，如省略则自动检测)
- `--to` - 目标平台 (iflytek|dify|coze) **[必需]**
- `--input, -i` - 输入 DSL 文件路径 **[必需]**
- `--output, -o` - 输出 DSL 文件路径 **[必需]**

**支持的转换路径：**

```bash
# 星辰 → Dify
ai-agent-converter convert --from iflytek --to dify --input agent.yml --output dify.yml

# 星辰 → Coze
ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

# Coze ZIP → 星辰
ai-agent-converter convert --from coze --to iflytek --input workflow.zip --output agent.yml

# Dify → 星辰
ai-agent-converter convert --from dify --to iflytek --input dify.yml --output agent.yml
```

**不支持的转换路径：**

```bash
# ❌ Dify → Coze (直接转换)
# 使用两步转换：
ai-agent-converter convert --from dify --to iflytek --input dify.yml --output temp.yml
ai-agent-converter convert --from iflytek --to coze --input temp.yml --output coze.yml
```

### validate

验证 DSL 文件格式和结构。

**用法：**

```bash
ai-agent-converter validate [flags]
```

**参数：**

- `--from` - 源平台 (iflytek|dify|coze，如省略则自动检测)
- `--input, -i` - 输入 DSL 文件路径 **[必需]**

**示例：**

```bash
# 验证星辰DSL文件
ai-agent-converter validate --input agent.yml --from iflytek

# 验证Dify DSL文件
ai-agent-converter validate --input dify.yml --from dify

# 自动检测并验证
ai-agent-converter validate --input agent.yml

# 注意：validate命令不支持ZIP格式验证
# 请使用convert命令来验证ZIP文件兼容性
```

### info

显示工具功能和支持的特性。

**用法：**

```bash
ai-agent-converter info [flags]
```

**示例：**

```bash
# 显示基本工具信息
ai-agent-converter info

# 显示支持的节点类型
ai-agent-converter info --nodes

# 显示数据类型映射
ai-agent-converter info --types

# 显示所有信息（节点类型 + 数据类型）
ai-agent-converter info --all
```

### platforms

列出支持的 AI 智能体平台。

**用法：**

```bash
ai-agent-converter platforms [flags]
```

**示例：**

```bash
# 列出所有平台
ai-agent-converter platforms

# 显示详细平台信息
ai-agent-converter platforms --detailed
```

### batch

批量转换多个工作流文件，支持并发处理以提高转换效率。

**用法：**

```bash
ai-agent-converter batch [flags]
```

**参数：**

- `--from` - 源平台 (iflytek|dify|coze) **[必需]**
- `--to` - 目标平台 (iflytek|dify|coze) **[必需]**
- `--input-dir` - 输入目录 **[必需]**
- `--output-dir` - 输出目录 **[必需]**
- `--pattern` - 文件模式 (默认: \*.yml)
- `--workers` - 并发工作线程数 (默认：根据 CPU 核心数自动检测)
- `--overwrite` - 自动覆盖已存在的输出文件，无需提示

**示例：**

```bash
# 使用默认工作线程数批量转换星辰到Coze
ai-agent-converter batch --from iflytek --to coze --input-dir ./workflows --output-dir ./converted

# 使用自定义工作线程数批量转换
ai-agent-converter batch --from iflytek --to dify --input-dir ./workflows --output-dir ./converted --workers 8

# 使用文件模式匹配和自动覆盖
ai-agent-converter batch --from iflytek --to dify --input-dir ./workflows --pattern "*.yaml" --output-dir ./converted --overwrite
```

## 全局参数

- `--verbose, -v` - 启用详细输出
- `--quiet, -q` - 安静模式，仅显示错误
- `--help, -h` - 任何命令的帮助信息
- `--version` - 显示版本信息

## 支持的平台

| 平台         | 状态        | 角色     | 描述                                       |
| ------------ | ----------- | -------- | ------------------------------------------ |
| 科大讯飞星辰 | ✅ 生产就绪 | 中央枢纽 | 科大讯飞星火认知大模型智能体平台，转换中枢 |
| Dify 平台    | ✅ 生产就绪 | 终端平台 | 开源 LLM 应用开发平台，支持双向转换        |
| Coze 平台    | ✅ 生产就绪 | 终端平台 | 字节跳动 AI 智能体开发平台，支持 ZIP 格式  |

## 支持的节点类型

| 统一类型   | 星辰     | Dify                | Coze | 描述           |
| ---------- | -------- | ------------------- | ---- | -------------- |
| start      | 开始节点 | start               | ✅   | 工作流入口点   |
| end        | 结束节点 | end                 | ✅   | 工作流出口点   |
| llm        | 大模型   | llm                 | ✅   | 大语言模型节点 |
| code       | 代码     | code                | ✅   | 代码执行节点   |
| condition  | 分支器   | if-else             | ✅   | 条件分支       |
| classifier | 决策器   | question-classifier | ✅   | 问题分类       |
| iteration  | 迭代器   | iteration           | ✅   | 循环迭代       |

## 格式支持

### 输入格式

- **YAML 文件** (.yml, .yaml)：所有平台标准格式
- **ZIP 文件** (.zip)：Coze 官方导出格式，自动识别

### 输出格式

- **YAML 格式**：标准输出，直接兼容目标平台
- **自动格式化**：保持平台特定的结构和字段

## 核心特性

### 转换能力

- **双向转换**：星辰与 Dify、Coze 间完整双向支持
- **数据完整性**：基于统一 DSL 中间表示，保证无损转换
- **格式自动检测**：智能识别输入文件格式
- **ZIP 格式支持**：原生支持 Coze 官方导出的 ZIP 格式
- **并发处理**：支持并行批量转换，自动调整工作线程数

### 容错机制

- **不支持节点占位符转换**：将不支持的节点转换为代码占位符，保持工作流完整性
- **边连接修复**：自动处理节点跳过后的连接关系
- **友好提示**：提供详细的转换统计和手动调整建议
- **错误代码映射**：结构化错误处理，提供用户友好的错误信息和建议

### 验证流水线

- **结构验证**：检查 DSL 基础结构完整性
- **语义验证**：验证节点间逻辑关系正确性
- **平台验证**：确保输出符合目标平台规范

## 错误处理

### 文件错误

```bash
Error: input file does not exist: agent.yml
# 解决：检查文件路径是否正确
```

### 格式错误

```bash
Error: invalid YAML format: yaml: line 10: found character that cannot start any token
# 解决：检查YAML语法，特别是缩进和特殊字符
```

### 转换路径错误

```bash
Error: direct conversion between dify and coze is not supported. Please use iFlytek as intermediate hub:
  1. Convert dify → iflytek
  2. Convert iflytek → coze
# 解决：使用两步转换方式
```

### 验证失败

```bash
❌ DSL file validation failed, found 2 issues:
   1. missing required field: flowMeta
   2. node ID is required for node node_123
# 解决：根据具体错误信息修复DSL文件
```

## 使用建议

### 性能优化

- 使用`--quiet`模式减少批量操作的输出开销
- 使用`batch`命令的`--workers`参数优化并发处理
- 大文件转换时使用`--verbose`监控进度
- 批量模式下使用`--overwrite`避免交互式提示

### 最佳实践

1. **转换前验证**：使用`validate`命令确保源文件正确
2. **备份原文件**：转换前备份重要工作流文件
3. **测试转换结果**：转换后在目标平台测试工作流正确性
4. **查看统计信息**：关注转换统计信息和占位符节点信息
5. **优化工作线程数**：批量处理时，根据系统资源和文件复杂度调整`--workers`参数

### 故障排除

1. **检查文件权限**：确保有读写相应目录的权限
2. **验证文件格式**：使用`validate`命令检查输入文件
3. **查看详细日志**：使用`--verbose`参数获取详细错误信息
4. **确认转换路径**：参考支持的转换路径表
