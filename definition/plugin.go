package definition

type Plugin interface {
	// ID is the canonical representation of the plugin
	ID() string
	// Alias is a human-readable name for the plugin which is unique.
	Alias() string
	// Homepage for the plugin
	Homepage() string
	// Description for the plugin
	Description() string
	// Maintainers for the plugin
	Maintainers() []string

	// Installation hooks
	BeforeInstall() error
	Install() error
	AfterInstall() error
}
