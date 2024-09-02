package unleash

// openapi file from {unleash-host}/docs/openapi.json
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -generate models,client,gin-server,strict-server -o unleash_gen.go -package unleash ./unleash-openapi.json
