package workout

import _ "embed"

//go:embed workouts.swagger.json
var swaggerFile []byte

// New creates a new instance of the workouts service
func GetSwaggerDescription() []byte {
	return swaggerFile
}
