package main

import (
	"context"
	"encoding/json"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"strings"
	"time"
)

type WavelogDocker struct{}

func (m *WavelogDocker) ListWavelogGitTags(ctx context.Context) (JSON, error) {
	tagsString, err := dag.Container().
		From("bitnami/git:2").
		WithDefaultArgs([]string{"git", "ls-remote", "--tags", "https://github.com/wavelog/wavelog.git"}).
		Stdout(ctx)

	if err != nil {
		return "", err
	}

	tags := make(map[string]string)
	tagLines := strings.Split(strings.TrimSpace(tagsString), "\n")

	for _, tagLine := range tagLines {
		s := strings.Split(tagLine, "\t")

		tag := strings.TrimPrefix(s[1], "refs/tags/")
		sha := s[0]

		tags[tag] = sha

		// TODO: Work out a proper way to tag the latest build as latest
		//if i >= len(tagLines)-1 {
		//	tags["latest"] = sha
		//}
	}

	ret, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return JSON(ret), nil
}

func (m *WavelogDocker) AddLabels(c *Container) *Container {
	return c.
		WithLabel(oci.AnnotationCreated, time.Now().Format(time.RFC3339)).
		WithLabel(oci.AnnotationAuthors, "Philip DO3PAR").
		WithLabel(oci.AnnotationURL, "https://github.com/philipreinken/wavelog-docker").
		WithLabel(oci.AnnotationDocumentation, "https://github.com/philipreinken/wavelog-docker/blob/main/README.md").
		WithLabel(oci.AnnotationSource, "https://github.com/philipreinken/wavelog-docker").
		WithLabel(oci.AnnotationVendor, "Philip Reinken").
		WithLabel(oci.AnnotationLicenses, "MIT").
		WithLabel(oci.AnnotationTitle, "wavelog").
		WithLabel(oci.AnnotationDescription, "Webbased Amateur Radio Logging Software")
}

// Builds the container as specified in docker/Dockerfile
func (m *WavelogDocker) Build(ctx context.Context, wavelogVersion string, phpVersion string) *Container {
	opts := ContainerBuildOpts{Dockerfile: "Dockerfile", BuildArgs: []BuildArg{}}

	opts.BuildArgs = append(opts.BuildArgs, BuildArg{Name: "PHP_VERSION", Value: phpVersion})
	opts.BuildArgs = append(opts.BuildArgs, BuildArg{Name: "WAVELOG_VERSION", Value: wavelogVersion})

	return m.AddLabels(dag.Container()).
		Build(dag.CurrentModule().Source().Directory("docker"), opts)
}
