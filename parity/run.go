package parity

type Run interface {
	Plugin
	Run() error
}
