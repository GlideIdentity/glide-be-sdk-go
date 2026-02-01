module github.com/GlideIdentity/glide-be-sdk-go/v2

go 1.21

retract (
	v1.1.0 // Incorrectly indexed by Go proxy
	v1.0.1 // Incorrectly indexed by Go proxy
	v1.0.0 // Incorrectly indexed by Go proxy
)

require github.com/GlideIdentity/glide-be-sdk-go/core/v2 v2.0.0
 