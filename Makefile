
.PHONY: build build-and-push-auto

build:
	dagger call build --php-version=8.3 --wavelog-version=1.3.1

build-and-push-auto:
	dagger call \
		with-registry-auth \
			--address="${CI_REGISTRY_ADDRESS}" \
			--username="${CI_REGISTRY_USER}" \
			--secret="env:CI_REGISTRY_TOKEN" \
		publish-pipeline \
			--name="${CI_REGISTRY_USER}/wavelog"