## harbor-arm64

This repo is forked from goharbor/harbor. It's based on v1.9.3 to modify.Run and tested on arm64.

Tested on Huawei Thaishan V2280H Centos 7.6.

## What's Changed

* base image changed to Photon 3.0 to support arm64.

* BUILDBIN is set to true in Makefile

* The ENV REGISTRY_SRC_TAG=$(REGISTRY_SRC_TAG) added to Makefile build flag

* change ../binary/registry/ to ../binary/bin/registry in make/photon/Makefile

* change ../binary/registry/ to ../binary/bin/registry in make/photon/registry/Dockerfile

* change ../binary/registry/ to ../binary/bin/registry in make/photon/registryctl/Dockerfile

* change v0.4.1/dep-linux-amd64 to v0.5.4/dep-linux-arm64 in make/photon/notary/binary.Dockerfile

* replace dumb_init with arm64 one

* portal images use multi-stage building but no arm surport for node.js.I simply complete the stage 1 job in x86 machine and the output dir is build_dir.

* rebuild redis docker-library redis to avoid "Unsupported system page size" error, tested on Huawei Kunpeng920 machine.

## How To Run 

Clone this repo and run the command to build your own package:

`make package_offline -e  VERSIONTAG=v1.9.3 PKGVERSIONTAG=v1.9.3 UIVERSIONTAG=v1.9.3  DEVFLAG=false  CLAIRFLAG=true`

or download the release package [here](https://github.com/hzliangbin/harbor-arm64/releases/tag/v1.9.3).

Detailed intall instructions are [here](https://github.com/hzliangbin/harbor-arm64/blob/master/docs/installation_guide.md). 
## FAQ
1、sudo: Account or password is expired, reset your password and try again

A:因为当前版本是1.9.3，这是harbor的一个[BUG](https://github.com/goharbor/harbor/issues/11633),解决参考文档：https://xie.infoq.cn/article/07b3a4de1062852e23395b19e

2、no logs are available with the 'syslog' log driver

修改yaml中registry、registryctl的日志驱动格式为json-file

3、registry       | sudo: unable to change expired password: Authentication token manipulation error

A:问题跟第一个类似，直接添加以下内容即可
```
      - type: bind
        source: /opt/harbor-log-etc/shadow
        target: /etc/shadow
```

