package parity

// Builder is the interface for all Plugins that build and publish
// Docker images
type Builder interface {
	Plugin
	Build() error
	Publish() error
}
