package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"sort"
	"strings"
)

func (m *WavelogDocker) ListTags(ctx context.Context, repository string) ([]string, error) {
	tagsString, err := dag.Container().
		From("bitnami/git:2").
		WithDefaultArgs([]string{"git", "ls-remote", "--tags", repository}).
		Stdout(ctx)

	if err != nil {
		return nil, err
	}

	var tags []string

	for _, tagLine := range strings.Split(strings.TrimSpace(tagsString), "\n") {
		s := strings.Split(tagLine, "\t")
		tag := strings.TrimPrefix(s[1], "refs/tags/")

		tags = append(tags, tag)
	}

	return tags, nil
}

func (m *WavelogDocker) GetTagsForLatestTwoMinorVersions(ctx context.Context) ([]string, error) {
	var ret []string

	tags, err := m.ListTags(ctx, wavelogRepoUrl)
	if err != nil {
		return nil, err
	}

	if len(tags) < 1 {
		return nil, fmt.Errorf("No tags found!")
	}

	versions := make([]*version.Version, len(tags))

	for i, raw := range tags {
		v, _ := version.NewVersion(raw)
		versions[i] = v
	}

	sort.Sort(version.Collection(versions))

	// Get the latest tag
	latest := versions[len(versions)-1]
	currentMajor := latest.Segments()[0]
	previousMinor := latest.Segments()[1] - 1

	if previousMinor < 0 {
		previousMinor = 0
	}

	constraint, err := version.NewConstraint(fmt.Sprintf(">= %d.%d, <= %s", currentMajor, previousMinor, latest.String()))
	if err != nil {
		return nil, err
	}

	// Filter tags that match the current and previous minor versions
	for _, v := range versions {
		if constraint.Check(v) {
			ret = append(ret, v.Original())
		}
	}

	return ret, nil
}
