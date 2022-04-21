package images

import (
	"fmt"
	"github.com/containers/image/v5/docker"
	"github.com/containers/image/v5/types"
	"strings"
)

func ParseReference(imageStr string) (types.ImageReference, error) {
	if !strings.HasPrefix(imageStr, "//") {
		imageStr = fmt.Sprintf("//%s", imageStr)
	}
	return docker.ParseReference(imageStr)
}
