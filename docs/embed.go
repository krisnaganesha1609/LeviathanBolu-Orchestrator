// Package docs embeds the OpenAPI spec so it ships inside the compiled
// binary instead of depending on a docs/ folder existing on disk next to
// wherever the binary happens to run (which the Dockerfile didn't even
// copy before this change).
package docs

import _ "embed"

//go:embed swagger.json
var SwaggerJSON []byte
