
.PHONY: build build-and-push-auto

build:
	dagger call build-containers-for-current-versions --flavours="apache" --php-versions="8.1,8.2,8.3" labels value

build-and-push-auto:
	dagger call \
		with-registry-auth \
			--address="${CI_REGISTRY_ADDRESS}" \
			--username="${CI_REGISTRY_USER}" \
			--secret="env:CI_REGISTRY_TOKEN" \
		with-containers-for-current-versions \
			--flavours="apache" \
			--php-versions="8.2" \
		publish-containers \
		get-containers \
		labels value