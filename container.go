package main

import (
	"context"
	"fmt"
	oci "github.com/opencontainers/image-spec/specs-go/v1"
	"golang.org/x/sync/errgroup"
	"time"
)

const phpExtensionInstallerUrl = "https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions"
const wavelogRepoUrl = "https://github.com/wavelog/wavelog.git"

const port = 8080

const apachePortsConfig = `
Listen %d
`

const apacheSiteConfig = `
<VirtualHost *:%d>
	DocumentRoot /var/www/html
	ErrorLog /dev/stderr
	CustomLog /dev/stdout combined
</VirtualHost>
`

func base(
	// +optional
	c *Container,
	phpVersion,
	flavour string,
) *Container {
	if c == nil {
		c = dag.Container()
	}

	return c.From(fmt.Sprintf("php:%s-%s", phpVersion, flavour))
}

func withPhpExtensionInstaller(c *Container) *Container {
	return c.WithFile("/usr/local/bin/install-php-extensions", dag.HTTP(phpExtensionInstallerUrl), ContainerWithFileOpts{Permissions: 0555})
}

func withPhpExtensions(c *Container) *Container {
	return c.WithExec([]string{"install-php-extensions", "curl", "mbstring", "mysqli", "pdo_mysql", "xml", "zip"})
}

func withConfig(c *Container) *Container {
	return c.
		WithNewFile("/etc/apache2/ports.conf", ContainerWithNewFileOpts{
			Contents:    fmt.Sprintf(apachePortsConfig, port),
			Permissions: 0644,
		}).
		WithNewFile("/etc/apache2/sites-enabled/000-default.conf", ContainerWithNewFileOpts{
			Contents:    fmt.Sprintf(apacheSiteConfig, port),
			Permissions: 0644,
		})
}

func withModRewrite(c *Container) *Container {
	return c.WithExec([]string{"a2enmod", "rewrite"})
}

func withWavelog(c *Container, tag string) *Container {
	wavelogCode := dag.
		Git(wavelogRepoUrl).
		Tag(tag).
		Tree()

	return c.
		WithDirectory("/var/www/html", wavelogCode, ContainerWithDirectoryOpts{Owner: "www-data"}).
		WithFile("/var/www/html/.htaccess", wavelogCode.File(".htaccess.sample"), ContainerWithFileOpts{Owner: "www-data"}).
		WithLabel(oci.AnnotationTitle, "wavelog").
		WithLabel(oci.AnnotationDescription, "Container for wavelog - Webbased Amateur Radio Logging Software").
		WithLabel(oci.AnnotationVersion, tag)
}

func withLabels(c *Container) *Container {
	return c.
		WithLabel(oci.AnnotationCreated, time.Now().Format(time.RFC3339)).
		WithLabel(oci.AnnotationAuthors, "Philip DO3PAR").
		WithLabel(oci.AnnotationURL, "https://github.com/philipreinken/wavelog-docker").
		WithLabel(oci.AnnotationDocumentation, "https://github.com/philipreinken/wavelog-docker/blob/main/README.md").
		WithLabel(oci.AnnotationSource, "https://github.com/philipreinken/wavelog-docker").
		WithLabel(oci.AnnotationVendor, "Philip Reinken").
		WithLabel(oci.AnnotationLicenses, "MIT")

}

// BuildContainer builds a container for the given PHP version, flavour and wavelog version.
func (m *WavelogDocker) BuildContainer(
	ctx context.Context,
	// The PHP image flavour to use.
	// +default="apache"
	flavour,
	// The PHP version to use.
	// +default="8.2"
	phpVersion,
	// The version of wavelog to use.
	wavelogVersion string,
) *Container {
	if phpVersion == "" {
		phpVersion = "8.2"
	}

	if flavour == "" {
		flavour = "apache"
	}

	c := base(nil, phpVersion, flavour)

	c = withPhpExtensionInstaller(c)
	c = withPhpExtensions(c)
	c = withConfig(c)
	c = withModRewrite(c)
	c = withWavelog(c, wavelogVersion)
	c = withLabels(c)

	return c
}

// WithContainer builds a container for the given PHP version, flavour and wavelog version and attaches it to the module instance.
func (m *WavelogDocker) WithContainer(
	ctx context.Context,
	// The PHP image flavour to use.
	// +default="apache"
	flavour,
	// The PHP version to use.
	// +default="8.2"
	phpVersion,
	// The version of wavelog to use.
	wavelogVersion string,
) *WavelogDocker {
	m.Containers = append(m.Containers, m.BuildContainer(ctx, flavour, phpVersion, wavelogVersion))

	return m
}

// BuildContainers builds containers for all combinations of the given flavours, PHP versions and wavelog versions.
func (m *WavelogDocker) BuildContainers(
	ctx context.Context,
	// The PHP image flavours to use.
	flavours,
	// The PHP versions to use.
	phpVersions,
	// The versions of wavelog to use.
	wavelogVersions []string,
) ([]*Container, error) {
	var containers []*Container

	eg, gctx := errgroup.WithContext(ctx)
	eg.SetLimit(len(flavours) * len(phpVersions) * len(wavelogVersions))

	for _, flavour := range flavours {
		for _, phpVersion := range phpVersions {
			for _, wavelogVersion := range wavelogVersions {
				eg.Go(m.syncBuilder(gctx, &containers, flavour, phpVersion, wavelogVersion))
			}
		}
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return containers, nil
}

// WithContainers builds containers for all combinations of the given flavours, PHP versions and wavelog versions and attaches them to the module instance.
func (m *WavelogDocker) WithContainers(
	ctx context.Context,
	// The PHP image flavours to use.
	flavours,
	// The PHP versions to use.
	phpVersions,
	// The versions of wavelog to use.
	wavelogVersions []string,
) (*WavelogDocker, error) {
	c, err := m.BuildContainers(ctx, flavours, phpVersions, wavelogVersions)
	if err != nil {
		return m, err
	}

	m.Containers = append(m.Containers, c...)

	return m, err
}

// BuildContainersForCurrentVersions builds containers, automatically selecting the current versions of wavelog.
func (m *WavelogDocker) BuildContainersForCurrentVersions(
	ctx context.Context,
	// The PHP image flavours to use.
	flavours,
	// The PHP versions to use.
	phpVersions []string,
) ([]*Container, error) {
	wavelogVersions, err := m.GetTagsForLatestTwoMinorVersions(ctx)
	if err != nil {
		return nil, err
	}

	return m.BuildContainers(ctx, flavours, phpVersions, wavelogVersions)
}

// WithContainersForCurrentVersions builds containers, automatically selecting the current versions of wavelog and attaches them to the module instance.
func (m *WavelogDocker) WithContainersForCurrentVersions(
	ctx context.Context,
	// The PHP image flavours to use.
	flavours,
	// The PHP versions to use.
	phpVersions []string,
) (*WavelogDocker, error) {
	c, err := m.BuildContainersForCurrentVersions(ctx, flavours, phpVersions)
	if err != nil {
		return m, err
	}

	m.Containers = append(m.Containers, c...)

	return m, err
}

// PublishContainer Publishes a single container.
func (m *WavelogDocker) PublishContainer(ctx context.Context, c *Container) (string, error) {
	if m.RegistryAuth == nil {
		return "", fmt.Errorf("RegistryAuth is not set! Define it using with-registry-auth!")
	}

	name, err := c.Label(ctx, oci.AnnotationTitle)
	if err != nil {
		return "", err
	}

	tag, err := c.Label(ctx, oci.AnnotationVersion)
	if err != nil {
		return "", err
	}

	return c.
		WithRegistryAuth(m.RegistryAuth.Address, m.RegistryAuth.Username, m.RegistryAuth.Secret).
		Publish(ctx, fmt.Sprintf("%s/%s/%s:%s", m.RegistryAuth.Address, m.RegistryAuth.Username, name, tag))
}

// PublishContainers Publishes containers prepared with WithContainer, WithContainers or WithContainersForCurrentVersions.
func (m *WavelogDocker) PublishContainers(ctx context.Context) error {
	if len(m.Containers) < 1 {
		return fmt.Errorf("no Containers to publish")
	}

	eg, gctx := errgroup.WithContext(ctx)
	eg.SetLimit(len(m.Containers))

	for _, c := range m.Containers {
		eg.Go(m.syncPublisher(gctx, c))
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

func (m *WavelogDocker) syncBuilder(ctx context.Context, containers *[]*Container, flavour, phpVersion, wavelogVersion string) func() error {
	return func() error {
		container, err := m.BuildContainer(ctx, flavour, phpVersion, wavelogVersion).Sync(ctx)

		if err == nil {
			// TODO: Use channels here
			*containers = append(*containers, container)
		}

		return err
	}
}

func (m *WavelogDocker) syncPublisher(ctx context.Context, c *Container) func() error {
	return func() error {
		_, err := m.PublishContainer(ctx, c)

		return err
	}
}
