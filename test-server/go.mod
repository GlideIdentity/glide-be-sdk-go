module github.com/glideidentity/glide-go-sdk/test-server

go 1.21

require (
	github.com/glideidentity/glide-go-sdk v0.1.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/mux v1.8.1
)

require (
	github.com/felixge/httpsnoop v1.0.4 // indirect
	golang.org/x/time v0.5.0 // indirect
)

replace github.com/glideidentity/glide-go-sdk => ../
