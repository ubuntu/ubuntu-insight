package sysinfo

// WithRoot overrides default root directory of the system.
func WithRoot(root string) Options {
	return func(o *options) {
		o.root = root
	}
}

// WithCpuInfo overrides default cpu info.
func WithCpuInfo(cmd []string) Options {
	return func(o *options) {
		o.cpuInfoCmd = cmd
	}
}

// WithCpuInfo overrides default blk info.
func WithBlkInfo(cmd []string) Options {
	return func(o *options) {
		o.lsblkCmd = cmd
	}
}

func WithScreenInfo(cmd []string) Options {
	return func(o *options) {
		o.screenCmd = cmd
	}
}
