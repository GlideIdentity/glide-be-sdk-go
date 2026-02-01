package glide

import "github.com/GlideIdentity/glide-be-sdk-go/core/v2"

// Version is re-exported from core for convenience.
const Version = core.Version

// GetVersion returns the SDK version.
func GetVersion() string {
	return core.GetVersion()
}
