# AgentBridge CLI User Manual

Cross-platform AI agent workflow DSL converter with iFlytek Spark as the central hub, supporting bidirectional conversion between Spark, Dify, and Coze platforms.

## Conversion Architecture

**Star Architecture Design**: iFlytek Spark serves as the central hub, supporting the following conversion paths:

```
    Dify ←→ iFlytek ←→ Coze
         (Central Hub)
```

- ✅ **iFlytek ↔ Dify**: Full bidirectional conversion
- ✅ **iFlytek ↔ Coze**: Full bidirectional conversion with ZIP format support
- ❌ **Dify ↔ Coze**: Direct conversion not supported, requires iFlytek as intermediate hub

## Quick Start

```bash
# Build the CLI tool
go build -o ai-agent-converter ./cmd/

# Basic conversion: iFlytek to Coze
./ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

# Coze ZIP to iFlytek
./ai-agent-converter convert --from coze --to iflytek --input workflow.zip --output agent.yml

# Auto-detect format
./ai-agent-converter convert --to dify --input agent.yml --output dify.yml

# Validate DSL file
./ai-agent-converter validate --input agent.yml
```

## Commands

### convert

Convert between different AI agent platforms.

**Usage:**

```bash
ai-agent-converter convert [flags]
```

**Flags:**

- `--from` - Source platform (iflytek|dify|coze, auto-detect if omitted)
- `--to` - Target platform (iflytek|dify|coze) **[required]**
- `--input, -i` - Input DSL file path **[required]**
- `--output, -o` - Output DSL file path **[required]**

**Supported Conversion Paths:**

```bash
# iFlytek → Dify
ai-agent-converter convert --from iflytek --to dify --input agent.yml --output dify.yml

# iFlytek → Coze
ai-agent-converter convert --from iflytek --to coze --input agent.yml --output coze.yml

# Coze ZIP → iFlytek
ai-agent-converter convert --from coze --to iflytek --input workflow.zip --output agent.yml

# Dify → iFlytek
ai-agent-converter convert --from dify --to iflytek --input dify.yml --output agent.yml
```

**Unsupported Conversion Paths:**

```bash
# ❌ Dify → Coze (direct conversion)
# Use two-step conversion:
ai-agent-converter convert --from dify --to iflytek --input dify.yml --output temp.yml
ai-agent-converter convert --from iflytek --to coze --input temp.yml --output coze.yml
```

### validate

Validate DSL file format and structure.

**Usage:**

```bash
ai-agent-converter validate [flags]
```

**Flags:**

- `--from` - Source platform (iflytek|dify|coze, auto-detect if omitted)
- `--input, -i` - Input DSL file path **[required]**

**Examples:**

```bash
# Validate iFlytek DSL file
ai-agent-converter validate --input agent.yml --from iflytek

# Validate Dify DSL file
ai-agent-converter validate --input dify.yml --from dify

# Auto-detect and validate
ai-agent-converter validate --input agent.yml

# Note: ZIP format validation is not supported by validate command
# Use convert command to verify ZIP file compatibility
```

### info

Display tool capabilities and supported features.

**Usage:**

```bash
ai-agent-converter info [flags]
```

**Examples:**

```bash
# Show basic tool information
ai-agent-converter info

# Show supported node types
ai-agent-converter info --nodes

# Show data type mappings
ai-agent-converter info --types

# Show all information (nodes + types)
ai-agent-converter info --all
```

### platforms

List supported AI agent platforms.

**Usage:**

```bash
ai-agent-converter platforms [flags]
```

**Examples:**

```bash
# List all platforms
ai-agent-converter platforms

# Show detailed platform information
ai-agent-converter platforms --detailed
```

### batch

Batch convert multiple workflow files with concurrent processing for efficient conversion.

**Usage:**

```bash
ai-agent-converter batch [flags]
```

**Flags:**

- `--from` - Source platform (iflytek|dify|coze) **[required]**
- `--to` - Target platform (iflytek|dify|coze) **[required]**
- `--input-dir` - Input directory **[required]**
- `--output-dir` - Output directory **[required]**
- `--pattern` - File pattern (default: \*.yml)
- `--workers` - Number of concurrent workers (default: auto-detect based on CPU cores)
- `--overwrite` - Automatically overwrite existing output files without prompting

**Examples:**

```bash
# Batch convert iFlytek to Coze with default workers
ai-agent-converter batch --from iflytek --to coze --input-dir ./workflows --output-dir ./converted

# Batch convert with custom worker count
ai-agent-converter batch --from iflytek --to dify --input-dir ./workflows --output-dir ./converted --workers 8

# Batch convert with pattern matching and overwrite
ai-agent-converter batch --from iflytek --to dify --input-dir ./workflows --pattern "*.yaml" --output-dir ./converted --overwrite
```

## Global Flags

- `--verbose, -v` - Enable verbose output
- `--quiet, -q` - Quiet mode, only show errors
- `--help, -h` - Help for any command
- `--version` - Show version information

## Supported Platforms

| Platform            | Status              | Role              | Description                                                      |
| ------------------- | ------------------- | ----------------- | ---------------------------------------------------------------- |
| iFlytek Spark Agent | ✅ Production Ready | Central Hub       | iFlytek Spark AI Agent Platform, conversion hub                  |
| Dify Platform       | ✅ Production Ready | Terminal Platform | Open-source LLM application platform, bidirectional with iFlytek |
| Coze Platform       | ✅ Production Ready | Terminal Platform | ByteDance AI agent platform with ZIP format support              |

## Supported Node Types

| Unified Type | iFlytek Spark   | Dify                | Coze | Description               |
| ------------ | --------------- | ------------------- | ---- | ------------------------- |
| start        | start-node      | start               | ✅   | Workflow entry point      |
| end          | end-node        | end                 | ✅   | Workflow exit point       |
| llm          | llm-node        | llm                 | ✅   | Large Language Model node |
| code         | code-node       | code                | ✅   | Code execution node       |
| condition    | condition-node  | if-else             | ✅   | Conditional branching     |
| classifier   | classifier-node | question-classifier | ✅   | Question classification   |
| iteration    | iteration-node  | iteration           | ✅   | Loop iteration            |

## Format Support

### Input Formats

- **YAML files** (.yml, .yaml): Standard format for all platforms
- **ZIP files** (.zip): Coze official export format, auto-detected

### Output Formats

- **YAML format**: Standard output, directly compatible with target platforms
- **Auto-formatting**: Maintains platform-specific structure and fields

## Core Features

### Conversion Capabilities

- **Bidirectional Conversion**: Full bidirectional support between iFlytek and Dify/Coze
- **Data Integrity**: Lossless conversion based on unified DSL intermediate representation
- **Auto-Format Detection**: Intelligent input file format recognition
- **ZIP Format Support**: Native support for Coze official ZIP export format
- **Concurrent Processing**: Parallel batch conversion with auto-scaling worker threads

### Fault Tolerance

- **Unsupported Node Placeholder Conversion**: Converts unsupported nodes to code placeholders, maintaining workflow integrity
- **Edge Connection Repair**: Automatically handles connection relationships after node skipping
- **User-Friendly Feedback**: Provides detailed conversion statistics and manual adjustment guidance
- **Error Code Mapping**: Structured error handling with user-friendly messages and suggestions

### Validation Pipeline

- **Structural Validation**: Checks DSL basic structure integrity
- **Semantic Validation**: Verifies correctness of inter-node logical relationships
- **Platform Validation**: Ensures output conforms to target platform specifications

## Error Handling

### File Errors

```bash
Error: input file does not exist: agent.yml
# Solution: Check if the file path is correct
```

### Format Errors

```bash
Error: invalid YAML format: yaml: line 10: found character that cannot start any token
# Solution: Check YAML syntax, especially indentation and special characters
```

### Conversion Path Errors

```bash
Error: direct conversion between dify and coze is not supported. Please use iFlytek as intermediate hub:
  1. Convert dify → iflytek
  2. Convert iflytek → coze
# Solution: Use two-step conversion approach
```

### Validation Failures

```bash
❌ DSL file validation failed, found 2 issues:
   1. missing required field: flowMeta
   2. node ID is required for node node_123
# Solution: Fix DSL file according to specific error messages
```

## Usage Recommendations

### Performance Optimization

- Use `--quiet` mode to reduce output overhead for batch operations
- Use `batch` command with `--workers` parameter to optimize concurrent processing
- Use `--verbose` to monitor progress for large file conversions
- Use `--overwrite` in batch mode to avoid interactive prompts

### Best Practices

1. **Pre-conversion Validation**: Use `validate` command to ensure source file correctness
2. **Backup Original Files**: Backup important workflow files before conversion
3. **Test Conversion Results**: Test workflow correctness in target platform after conversion
4. **Review Statistics**: Pay attention to conversion statistics and placeholder node information
5. **Optimize Worker Count**: For batch processing, adjust `--workers` based on system resources and file complexity

### Troubleshooting

1. **Check File Permissions**: Ensure read/write permissions for relevant directories
2. **Validate File Format**: Use `validate` command to check input files
3. **View Detailed Logs**: Use `--verbose` parameter to get detailed error information
4. **Confirm Conversion Path**: Refer to the supported conversion paths table

## Technical Implementation

### Architecture Benefits

- **Star Architecture**: Centralized conversion hub reduces complexity compared to mesh architecture
- **Unified DSL**: Intermediate representation ensures data integrity across conversions
- **Modular Design**: Platform-specific parsers and generators enable easy extension
- **Fault Tolerance**: Graceful handling of unsupported features maintains system stability

### Conversion Process

1. **Parse**: Convert source platform DSL to unified DSL
2. **Validate**: Three-stage validation ensures data correctness
3. **Generate**: Convert unified DSL to target platform format
4. **Report**: Provide detailed conversion statistics and recommendations
