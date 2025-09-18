// Package models contains error types for the AI Agents Transformer.
package models

import (
	"fmt"
)

// ErrorSeverity represents the severity level of an error.
type ErrorSeverity string

const (
	// SeverityCritical indicates a critical error that prevents operation
	SeverityCritical ErrorSeverity = "critical"
	// SeverityError indicates a regular error
	SeverityError ErrorSeverity = "error"
	// SeverityWarning indicates a warning that doesn't prevent operation
	SeverityWarning ErrorSeverity = "warning"
	// SeverityInfo indicates informational message
	SeverityInfo ErrorSeverity = "info"
)

// ValidationError represents a validation error in DSL processing.
type ValidationError struct {
	Type           string   `json:"type"`     // node/edge/reference
	Severity       string   `json:"severity"` // error/warning
	Message        string   `json:"message"`
	AffectedItems  []string `json:"affected_items,omitempty"`
	FixSuggestions []string `json:"fix_suggestions,omitempty"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Severity, e.Message)
}

// ParseError represents a parsing error with location information.
type ParseError struct {
	Code        string    `json:"code"`
	Message     string    `json:"message"`
	Location    *Location `json:"location,omitempty"`
	Suggestions []string  `json:"suggestions,omitempty"`
}

// Location represents error location information.
type Location struct {
	NodeID string `json:"node_id,omitempty"`
	Field  string `json:"field,omitempty"`
	Line   int    `json:"line,omitempty"`
}

func (e *ParseError) Error() string {
	if e.Location != nil {
		return fmt.Sprintf("[%s] %s at %s:%s (line %d)", e.Code, e.Message, e.Location.NodeID, e.Location.Field, e.Location.Line)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// ConversionError represents a conversion error between platforms.
type ConversionError struct {
	// Core error information
	Code           string    `json:"code"`
	Message        string    `json:"message"`
	SourceLocation *Location `json:"source_location,omitempty"`
	Suggestions    []string  `json:"suggestions,omitempty"`
	Severity       ErrorSeverity `json:"severity,omitempty"`
	
	// Platform-specific information (preserved for compatibility)
	SourcePlatform   string `json:"source_platform"`
	TargetPlatform   string `json:"target_platform"`
	ErrorType        string `json:"error_type"`
	Details          string `json:"details"`
	FallbackStrategy string `json:"fallback_strategy,omitempty"`
}

func (e *ConversionError) Error() string {
	// Use new format if Code is available, otherwise fall back to old format
	if e.Code != "" {
		if e.SourceLocation != nil {
			return fmt.Sprintf("[%s] %s at %s:%s", e.Code, e.Message, e.SourceLocation.NodeID, e.SourceLocation.Field)
		}
		return fmt.Sprintf("[%s] %s", e.Code, e.Message)
	}
	// Backward compatibility: use old format
	return fmt.Sprintf("Conversion error from %s to %s: [%s] %s", e.SourcePlatform, e.TargetPlatform, e.ErrorType, e.Details)
}
