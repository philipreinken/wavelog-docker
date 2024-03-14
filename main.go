package main

import (
	"context"
	"encoding/json"
	"fmt"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"strings"
	"time"
)

const wavelogRepoUrl = "https://github.com/wavelog/wavelog.git"

type WavelogDocker struct{}

func listTags(ctx context.Context, repository string) (map[string]string, error) {
	tagsString, err := dag.Container().
		From("bitnami/git:2").
		WithDefaultArgs([]string{"git", "ls-remote", "--tags", repository}).
		Stdout(ctx)

	if err != nil {
		return nil, err
	}

	tags := make(map[string]string)
	tagLines := strings.Split(strings.TrimSpace(tagsString), "\n")

	for _, tagLine := range tagLines {
		s := strings.Split(tagLine, "\t")

		tag := strings.TrimPrefix(s[1], "refs/tags/")
		sha := s[0]

		tags[tag] = sha
	}

	return tags, nil
}

func (m *WavelogDocker) ListWavelogGitTags(ctx context.Context) (JSON, error) {
	tags, err := listTags(ctx, wavelogRepoUrl)
	if err != nil {
		return "", err
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
func (m *WavelogDocker) Build(
	ctx context.Context,
	// +optional
	container *Container,
	wavelogVersion string,
	phpVersion string,
) *Container {
	opts := ContainerBuildOpts{Dockerfile: "Dockerfile", BuildArgs: []BuildArg{}}

	opts.BuildArgs = append(opts.BuildArgs, BuildArg{Name: "PHP_VERSION", Value: phpVersion})
	opts.BuildArgs = append(opts.BuildArgs, BuildArg{Name: "WAVELOG_VERSION", Value: wavelogVersion})

	if container == nil {
		container = dag.Container()
	}

	return m.AddLabels(container).
		Build(dag.CurrentModule().Source().Directory("docker"), opts)
}

func (m *WavelogDocker) BuildPipeline(ctx context.Context) ([]*Container, error) {
	build := dag.Pipeline("build")
	wavelogTags, err := listTags(ctx, wavelogRepoUrl)
	if err != nil {
		return nil, err
	}

	if len(wavelogTags) < 1 {
		return nil, fmt.Errorf("No tags found in repository %s!", wavelogRepoUrl)
	}

	var containers []*Container

	for tag, sha := range wavelogTags {
		pipelinedContainer := build.Pipeline("wavelog").Pipeline(tag).Container().
			WithLabel(oci.AnnotationVersion, tag).
			WithLabel(oci.AnnotationRevision, sha)

		containers = append(containers, m.Build(ctx, pipelinedContainer, tag, "8.3"))
	}

	return containers, nil
}

func (m *WavelogDocker) PublishPipeline(ctx context.Context) error {
	publish := dag.Pipeline("publish")
	containers, err := m.BuildPipeline(ctx)
	if err != nil {
		return err
	}

	for _, container := range containers {
		id, err := container.ID(ctx)
		if err != nil {
			return err
		}

		response, err := publishGhcr(ctx, publish.LoadContainerFromID(id))
		if err != nil {
			return err
		}

		fmt.Println(response)
	}

	return nil
}

func publishGhcr(ctx context.Context, container *Container) (string, error) {
	version, err := container.Label(ctx, oci.AnnotationVersion)
	if err != nil {
		return version, err
	}

	return container.Publish(ctx, fmt.Sprintf("ghcr.io/philipreinken/wavelog:%s", version))
}
