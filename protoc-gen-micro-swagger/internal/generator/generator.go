// Package generator provides an abstract interface to code generators.
package generator

import (
	"github.com/BlackCodes/micro-swagger/protoc-gen-micro-swagger/internal/descriptor"
)

// Generator is an abstraction of code generators.
type Generator interface {
	// Generate generates output files from input .proto files.
	Generate(targets []*descriptor.File) ([]*descriptor.ResponseFile, error)
}
