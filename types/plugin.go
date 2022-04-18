package types

type Plugin interface {
	ID() string
	Alias() string
	Homepage() string
	Description() string
	Maintainers() []string
	InstallScript() string
}
