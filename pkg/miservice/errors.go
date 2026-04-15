package miservice

import "errors"

// Sentinel errors for the miservice package.
var (
	ErrLogin          = errors.New("login failed")
	ErrDeviceNotFound = errors.New("device not found")
	ErrPropGet        = errors.New("failed to get property")
	ErrPropSet        = errors.New("failed to set property")
	ErrAction         = errors.New("failed to execute action")
	ErrScene          = errors.New("scene operation failed")
	ErrSceneNotFound  = errors.New("scene not found")
	ErrTokenExpired   = errors.New("token expired")
	ErrStats          = errors.New("failed to get statistics")
	ErrConsumables    = errors.New("failed to get consumables")
	ErrHomeNotFound   = errors.New("home not found")
)
