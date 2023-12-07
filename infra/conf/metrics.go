package conf

import (
	"github.com/4nd3r5on/Xray-core/app/metrics"
)

type MetricsConfig struct {
	Tag string `json:"tag"`
}

func (c *MetricsConfig) Build() (*metrics.Config, error) {
	if c.Tag == "" {
		return nil, newError("metrics tag can't be empty.")
	}

	return &metrics.Config{
		Tag: c.Tag,
	}, nil
}
