// Package collector is the implementation of the collector component.
// The collector component is responsible for collecting data from sources, merging it into a report, and then writing the report to disk.
package collector

import (
	"log/slog"
	"time"

	"github.com/ubuntu/ubuntu-insights/internal/collector/sysinfo"
	"github.com/ubuntu/ubuntu-insights/internal/collector/sysinfo/software"
)

// Report contains all the info for a report.
type Report struct {
	App       string       `json:"appID"`
	Timestamp uint         `json:"generated"`
	Version   string       `json:"schemaVersion"`
	Common    sysinfo.Info `json:"common"`
	Specific  string       `json:"appData,string"`
}

type Collector struct {
	sysinfo      sysinfo.CollectorT[sysinfo.Info]
	timeProvider func() time.Time
	log          *slog.Logger
}

// Options is the variadic options available to the Collector.
type Options func(*options)

type options struct {
	sysinfo      sysinfo.CollectorT[sysinfo.Info]
	timeProvider func() time.Time
	log          *slog.Logger
}

// New returns a new SysInfo.
func New(source software.Source, tipe string, args ...Options) Collector {
	opts := &options{
		sysinfo:      sysinfo.New(source, tipe),
		timeProvider: func() time.Time { return time.Now() },
		log:          slog.Default(),
	}

	for _, opt := range args {
		opt(opts)
	}

	return Collector{
		sysinfo: opts.sysinfo,
		log:     opts.log,
	}
}

func (s Collector) Collect(period uint, app, appdata string) Report {
	t := uint(s.timeProvider().Unix())

	common, err := s.sysinfo.Collect()
	if err != nil {
		s.log.Warn("failed to collect common info", "error", err)
	}

	return Report{
		App:       app,
		Timestamp: t - t%period,
		Version:   "v0.0.1",
		Common:    common,
		Specific:  appdata,
	}
}
