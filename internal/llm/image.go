package llm

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/maximhq/bifrost/core/schemas"
)

var ErrVisionUnsupported = errors.New("model cannot process images")

type Image struct {
	Data     []byte
	MIMEType string
}

func contentBlocks(message Message) ([]schemas.ChatContentBlock, error) {
	blocks := make([]schemas.ChatContentBlock, 0, len(message.Images)+1)

	if message.Content != "" {
		blocks = append(blocks, schemas.ChatContentBlock{
			Type: schemas.ChatContentBlockTypeText,
			Text: schemas.Ptr(message.Content),
		})
	}

	for i, img := range message.Images {
		url, err := img.dataURL()
		if err != nil {
			return nil, fmt.Errorf("image %d: %w", i, err)
		}

		blocks = append(blocks, schemas.ChatContentBlock{
			Type:           schemas.ChatContentBlockTypeImage,
			ImageURLStruct: &schemas.ChatInputImage{URL: url},
		})
	}

	return blocks, nil
}

func (i Image) dataURL() (string, error) {
	if len(i.Data) == 0 {
		return "", errors.New("no image data")
	}

	mimeType := i.MIMEType
	if mimeType == "" {
		mimeType = http.DetectContentType(i.Data)
	}
	if !strings.HasPrefix(mimeType, "image/") {
		return "", fmt.Errorf("unsupported content type %q", mimeType)
	}

	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(i.Data)), nil
}
