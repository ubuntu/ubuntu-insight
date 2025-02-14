package software

// WithOsInfo overrides default os info.
func WithOSInfo(cmd []string) Options {
	return func(o *options) {
		o.platform.osCmd = cmd
	}
}

// WithLang overrides default language command.
func WithLang(cmd []string) Options {
	return func(o *options) {
		o.platform.langCmd = cmd
	}
}
