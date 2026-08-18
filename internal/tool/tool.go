package tool

import (
	"context"
	"encoding/json"

	"github.com/maximhq/bifrost/core/schemas"
)

type Tool interface {
	Info() (name, description string)
	Definition() schemas.ChatTool
	Execute(ctx context.Context, args json.RawMessage) (string, error)
}
