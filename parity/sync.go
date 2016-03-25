package parity

type Sync interface {
	Plugin
	Sync() error
}
