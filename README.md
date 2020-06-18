## harbor-arm64
This repo is forked form goharbor/harbor. It's based on v1.9.3 to modify.Run and tested on arm64.

## what's changed
* base image changed to Photon 3.0 to support arm64.
* BUILDIN is set to true in Makefile
* The ENV REGISTRY_SRC_TAG=$(REGISTRY_SRC_TAG) added to Makefile build flag
* change ../binary/registry/ to ../binary/bin/registry in make/photon/Makefile
* change ../binary/registry/ to ../binary/bin/registry in make/photon/registry/Dockerfile
* replace dump_init with arm64 one
