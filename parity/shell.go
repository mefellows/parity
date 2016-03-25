package parity

type Shell interface {
	Plugin
	Shell(ShellConfig) error
	Attach(ShellConfig) error
}

type ShellConfig struct {
	Command []string
	User    string
	Service string
}

var DEFAULT_INTERACTIVE_SHELL_OPTIONS = &ShellConfig{
	Service: "web",
	Command: []string{"bash"},
	User:    "",
}
