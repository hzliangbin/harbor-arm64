# harbor-arm64
This repo is forked form goharbor/harbor. And it's based on v1.9.3 to Modified to run on arm64.

# what's changed
1. base image changed to Photon 3.0 to support arm64.
2. BUILDIN is set to true in Makefile
3. The ENV REGISTRY_SRC_TAG=$(REGISTRY_SRC_TAG) added to Makefile build flag
4. 
