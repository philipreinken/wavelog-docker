package main

import (
	"context"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
)

type ContainerConfig struct {
	Flavour        string `json:"flavour"`
	PhpVersion     string `json:"php_version"`
	WavelogVersion string `json:"wavelog_version"`
}

// getContainerConfig determines the container configuration from the given container by parsing the container's OCI labels.
func getContainerConfig(ctx context.Context, c *Container) (*ContainerConfig, error) {
	var config ContainerConfig

	labels, err := c.Labels(ctx)
	if err != nil {
		return nil, err
	}

	for _, l := range labels {
		name, err := l.Name(ctx)
		if err != nil {
			return nil, err
		}

		value, err := l.Value(ctx)
		if err != nil {
			return nil, err
		}

		switch name {
		case oci.AnnotationBaseImageName:
			config.Flavour = value
			config.PhpVersion = value
		case oci.AnnotationVersion:
			config.WavelogVersion = value
		}
	}

	return &config, nil
}
