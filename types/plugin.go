package types

type Definition interface {
	ID() string
	Alias() string
	Homepage() string
	Description() string
	Maintainers() []string
}
