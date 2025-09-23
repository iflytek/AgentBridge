package common

import (
	"agentbridge/core/interfaces"
	"agentbridge/internal/models"
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// UnifiedIDMapper provides unified node ID mapping with configurable strategies.
type UnifiedIDMapper struct {
	mutex    sync.RWMutex
	mapping  map[string]string
	usedIDs  map[string]bool
	strategy IDGenerationStrategy
}

// IDGenerationStrategy defines ID generation strategies.
type IDGenerationStrategy int

const (
	StrategyTimestampBased IDGenerationStrategy = iota // Timestamp-based ID generation for general use
	StrategyCryptoSecure                               // Cryptographically secure ID generation
	StrategySimpleNumeric                              // Simple numeric ID generation for platforms requiring simple IDs
)

// NewUnifiedIDMapper creates a unified ID mapper with configurable strategies.
func NewUnifiedIDMapper(strategy IDGenerationStrategy) interfaces.IDMapper {
	return &UnifiedIDMapper{
		mapping:  make(map[string]string),
		usedIDs:  make(map[string]bool),
		strategy: strategy,
	}
}

// NewSecureNodeIDMapper creates a cryptographically secure ID mapper.
func NewSecureNodeIDMapper() interfaces.IDMapper {
	return NewUnifiedIDMapper(StrategyCryptoSecure)
}

// NewNodeIDMapper creates a timestamp-based ID mapper.
func NewNodeIDMapper() interfaces.IDMapper {
	return NewUnifiedIDMapper(StrategyTimestampBased)
}

// MapNodeID maps original node ID to generated ID.
func (m *UnifiedIDMapper) MapNodeID(originalID string, nodeType models.NodeType) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Return existing mapping if present
	if mappedID, exists := m.mapping[originalID]; exists {
		return mappedID
	}

	// Generate ID using configured strategy
	newID := m.generateIDWithStrategy(nodeType)
	m.mapping[originalID] = newID
	m.usedIDs[newID] = true

	return newID
}

// GetMapping returns the complete mapping table.
func (m *UnifiedIDMapper) GetMapping() map[string]string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Return copy to prevent concurrent modification
	result := make(map[string]string, len(m.mapping))
	for k, v := range m.mapping {
		result[k] = v
	}
	return result
}

// SetMapping sets the mapping table.
func (m *UnifiedIDMapper) SetMapping(mapping map[string]string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mapping = make(map[string]string, len(mapping))
	m.usedIDs = make(map[string]bool, len(mapping))

	for k, v := range mapping {
		m.mapping[k] = v
		m.usedIDs[v] = true
	}
}

// Clear clears all mappings.
func (m *UnifiedIDMapper) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.mapping = make(map[string]string)
	m.usedIDs = make(map[string]bool)
}

// HasMapping checks if a mapping exists.
func (m *UnifiedIDMapper) HasMapping(originalID string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	_, exists := m.mapping[originalID]
	return exists
}

// generateIDWithStrategy generates ID based on configured strategy.
func (m *UnifiedIDMapper) generateIDWithStrategy(nodeType models.NodeType) string {
	maxAttempts := 1000

	for attempts := 0; attempts < maxAttempts; attempts++ {
		var newID string

		switch m.strategy {
		case StrategyCryptoSecure:
			newID = m.generateSecureID()
		case StrategySimpleNumeric:
			newID = m.generateSimpleNumericID(attempts)
		case StrategyTimestampBased:
		default:
			newID = m.generateTimestampBasedID(attempts)
		}

		// Check if ID is already used
		if !m.usedIDs[newID] {
			return newID
		}
	}

	// Fallback if all attempts fail
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), maxAttempts)
}

// generateSecureID generates cryptographically secure ID.
func (m *UnifiedIDMapper) generateSecureID() string {
	// Generate 8 bytes of cryptographically secure random data
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based generation if crypto fails
		return m.generateTimestampBasedID(0)
	}

	// Convert to 16-digit numeric ID
	var result uint64
	for _, b := range bytes {
		result = (result << 8) + uint64(b)
	}

	return fmt.Sprintf("%016d", result)
}

// generateSimpleNumericID generates simple numeric ID for platforms requiring simple IDs.
func (m *UnifiedIDMapper) generateSimpleNumericID(attempt int) string {
	// Use timestamp + attempt for uniqueness
	timestamp := time.Now().UnixMilli() % 10000000000 // 10-digit millisecond timestamp

	// Add attempt to ensure uniqueness
	return fmt.Sprintf("%010d%06d", timestamp, attempt%1000000)
}

// generateTimestampBasedID generates timestamp-based ID.
func (m *UnifiedIDMapper) generateTimestampBasedID(attempt int) string {
	timestamp := time.Now().UnixMilli() % 10000000000 // 10-digit millisecond timestamp

	// Use last 6 digits of nanosecond timestamp as random component
	nanoRandom := (time.Now().UnixNano() % 1000000) + int64(attempt)

	return fmt.Sprintf("%010d%06d", timestamp, nanoRandom%1000000)
}

// SetStrategy sets the ID generation strategy.
func (m *UnifiedIDMapper) SetStrategy(strategy IDGenerationStrategy) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.strategy = strategy
}

// GetStrategy returns the current ID generation strategy.
func (m *UnifiedIDMapper) GetStrategy() IDGenerationStrategy {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.strategy
}

// Statistics returns mapper statistics.
func (m *UnifiedIDMapper) Statistics() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return map[string]interface{}{
		"total_mappings": len(m.mapping),
		"strategy":       m.strategy,
		"used_ids":       len(m.usedIDs),
	}
}
