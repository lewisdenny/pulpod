# Container Manager Notes

- Tested on both Podman version 5.3.1(latest) and 4.9.4(version shipping with RHEL 9.4).
- All conf can be overridden via `PULPOD_` env vars using `Koanf`
Note:
Testing should be inergration, to test if a container is created, check with the API rather than this package
Most fuctions will reuse the NewPodmanManager
delete test should test creating new connection, creating container, and deleting.
Can use builtin testing - nameing convention should state unit vs intergration - Like ITContainerDelete
remove os.exit - needer


## Links

https://github.com/containers/podman/tree/main/pkg/bindings
