package hypforms

import (
	"context"
	"errors"
	"fmt"
)

type contextKey string

type Validator func(ctx context.Context, value interface{}) (stop bool, err error)

var RequiredErr = errors.New("field required")

func Required(ctx context.Context, v interface{}) (stop bool, err error) {
	if v, _ := v.(string); v != "" {
		return false, nil
	}
	name, ok := ctx.Value(contextKey("name")).(string)
	if !ok {
		return true, RequiredErr
	}
	return true, fmt.Errorf("%w: %s", RequiredErr, name)
}

// Optional

// IsRegex

// IsEmail

// IsIPAddr

// IsMACAddr

// IsUUID

// IsAnyOf

// IsNoneOf

// LengthGt, LengthGe, LengthLt, LengthLe
