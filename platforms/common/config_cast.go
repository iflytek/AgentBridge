package common

import "ai-agents-transformer/internal/models"

// AsStartConfig returns a pointer to StartConfig regardless of value or pointer storage.
func AsStartConfig(cfg interface{}) (*models.StartConfig, bool) {
    switch c := cfg.(type) {
    case *models.StartConfig:
        return c, true
    case models.StartConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

// AsEndConfig returns a pointer to EndConfig regardless of value or pointer storage.
func AsEndConfig(cfg interface{}) (*models.EndConfig, bool) {
    switch c := cfg.(type) {
    case *models.EndConfig:
        return c, true
    case models.EndConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

// AsCodeConfig returns a pointer to CodeConfig regardless of value or pointer storage.
func AsCodeConfig(cfg interface{}) (*models.CodeConfig, bool) {
    switch c := cfg.(type) {
    case *models.CodeConfig:
        return c, true
    case models.CodeConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

// AsLLMConfig returns a pointer to LLMConfig regardless of value or pointer storage.
func AsLLMConfig(cfg interface{}) (*models.LLMConfig, bool) {
    switch c := cfg.(type) {
    case *models.LLMConfig:
        return c, true
    case models.LLMConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

// AsConditionConfig returns a pointer to ConditionConfig regardless of value or pointer storage.
func AsConditionConfig(cfg interface{}) (*models.ConditionConfig, bool) {
    switch c := cfg.(type) {
    case *models.ConditionConfig:
        return c, true
    case models.ConditionConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

// AsClassifierConfig returns a pointer to ClassifierConfig regardless of value or pointer storage.
func AsClassifierConfig(cfg interface{}) (*models.ClassifierConfig, bool) {
    switch c := cfg.(type) {
    case *models.ClassifierConfig:
        return c, true
    case models.ClassifierConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

// AsIterationConfig returns a pointer to IterationConfig regardless of value or pointer storage.
func AsIterationConfig(cfg interface{}) (*models.IterationConfig, bool) {
    switch c := cfg.(type) {
    case *models.IterationConfig:
        return c, true
    case models.IterationConfig:
        cc := c
        return &cc, true
    default:
        return nil, false
    }
}

