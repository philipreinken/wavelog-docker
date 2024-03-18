package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-version"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"sort"
	"strings"
	"time"
)

const wavelogRepoUrl = "https://github.com/wavelog/wavelog.git"

type WavelogDocker struct {
	RegistryAuth *RegistryAuth `json:"registryAuth"`
}

type RegistryAuth struct {
	Address  string  `json:"address"`
	Username string  `json:"username"`
	Secret   *Secret `json:"secret"`
}

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

func latestTag(tags map[string]string) (string, error) {
	versions := make([]*version.Version, 0, len(tags))

	for tag, _ := range tags {
		v, err := version.NewVersion(tag)
		if err != nil {
			return "", err
		}

		versions = append(versions, v)
	}

	sort.Sort(version.Collection(versions))

	return versions[len(versions)-1].Original(), nil
}

func (m *WavelogDocker) getContainer() *Container {
	if m.RegistryAuth != nil {
		return dag.Container().WithRegistryAuth(m.RegistryAuth.Address, m.RegistryAuth.Username, m.RegistryAuth.Secret)
	} else {
		return dag.Container()
	}
}

func (m *WavelogDocker) WithRegistryAuth(address string, username string, secret *Secret) *WavelogDocker {
	m.RegistryAuth = &RegistryAuth{Address: address, Username: username, Secret: secret}
	return m
}

func (m *WavelogDocker) GitTags(
	ctx context.Context,
	// +optional
	repository string,
) (JSON, error) {
	if repository == "" {
		repository = wavelogRepoUrl
	}

	tags, err := listTags(ctx, repository)
	if err != nil {
		return "", err
	}

	ret, err := json.Marshal(tags)
	if err != nil {
		return "", err
	}

	return JSON(ret), nil
}

func (m *WavelogDocker) LatestTag(
	ctx context.Context,
	// +optional
	repository string,
) (string, error) {
	if repository == "" {
		repository = wavelogRepoUrl
	}

	tags, err := listTags(ctx, repository)
	if err != nil {
		return "", err
	}

	return latestTag(tags)
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
		pipelinedContainer := build.Pipeline(tag).Container().
			WithLabel(oci.AnnotationVersion, tag).
			WithLabel(oci.AnnotationRevision, sha)

		containers = append(containers, m.Build(ctx, pipelinedContainer, tag, "8.3"))
	}

	return containers, nil
}

func (m *WavelogDocker) PublishPipeline(ctx context.Context) (string, error) {
	publish := dag.Pipeline("publish")
	containers, err := m.BuildPipeline(ctx)
	if err != nil {
		return "", err
	}
	tags, err := listTags(ctx, wavelogRepoUrl)
	if err != nil {
		return "", err
	}
	latestTag, err := latestTag(tags)
	if err != nil {
		return "", err
	}

	// TODO: Add a filter for the current and previous minor version

	responses := ""

	for _, container := range containers {
		id, err := container.ID(ctx)
		if err != nil {
			return "", err
		}

		tag, err := container.Label(ctx, oci.AnnotationVersion)
		if err != nil {
			return tag, err
		}

		response, err := m.Publish(ctx, publish.Pipeline(string(id)).LoadContainerFromID(id), tag)
		if err != nil {
			return response, err
		}

		responses = fmt.Sprintf("%s\n%s", responses, response)

		if tag == latestTag {
			response, err := m.Publish(ctx, publish.Pipeline("latest").LoadContainerFromID(id), "latest")
			if err != nil {
				return response, err
			}

			responses = fmt.Sprintf("%s\n%s", responses, response)

		}
	}

	return strings.TrimSpace(responses), nil
}

func (m *WavelogDocker) Publish(ctx context.Context, container *Container, tag string) (string, error) {
	if m.RegistryAuth == nil {
		return "", fmt.Errorf("RegistryAuth is not set! Define it using with-registry-auth!")
	}

	return container.
		WithRegistryAuth(m.RegistryAuth.Address, m.RegistryAuth.Username, m.RegistryAuth.Secret).
		Publish(ctx, fmt.Sprintf("%s/wavelog:%s", m.RegistryAuth.Address, tag))
}
