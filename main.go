package main

// TODO: Generate README.md

type WavelogDocker struct {
	RegistryAuth *RegistryAuth `json:"registryAuth"`
	Containers   []*Container  `json:"containers"`
}

type RegistryAuth struct {
	Address  string  `json:"address"`
	Username string  `json:"username"`
	Secret   *Secret `json:"secret"`
}

func (m *WavelogDocker) WithRegistryAuth(
	// Address is the address of the registry.
	// +optional
	// +default="docker.io"
	address,
	username string,
	secret *Secret,
) *WavelogDocker {
	m.RegistryAuth = &RegistryAuth{
		Address:  address,
		Username: username,
		Secret:   secret,
	}

	return m
}
