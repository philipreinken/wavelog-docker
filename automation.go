package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-version"
	"golang.org/x/exp/slices"
	"sort"
	"strings"
)

func (m *WavelogDocker) ListTags(
	ctx context.Context,
	// The repository to list tags for.
	// +optional
	// +default="https://github.com/wavelog/wavelog.git"
	repository string,
) ([]string, error) {
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
		tag = strings.TrimRightFunc(tag, func(r rune) bool {
			return r != '.' && r != 'v' && (r < '0' || r > '9')
		})

		if slices.Contains(tags, tag) {
			continue
		}

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
