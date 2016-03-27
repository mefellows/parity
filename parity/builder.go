package parity

// Builder is the interface for all Plugins that build and publish
// Docker images
type Builder interface {
	Plugin
	Build(BuilderConfig) error
	Publish(BuilderConfig) error
}

// BuilderConfig contains the mapping from yaml files to configure the Builders
type BuilderConfig struct {
	ImageName string ` mapstructure:"image_name"`
}
