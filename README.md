devcontainer-docker
====

VSCode Remote Container で devcontainer の中で docker を使うときの volume の問題をなんとなく解決します。

devcontainer 内で docker を使うために `/var/run/docker.sock` をマウントしていると、--volume で指定したパスが devcontainer 内のパスではなくホストのパスとして解釈されます。

このツールは docker をラップし、volumes のパスを devcontainer 内のパスとして解釈させるものです。

## Usage

バイナリをコピーして、本家の docker よりも優先されるように PATH を設定します。

```
RUN set -x \
    && mkdir -p /usr/local/devcontainer-tool/bin \
    && curl -fsSL -o /usr/local/devcontainer-tool/bin/docker https://github.com/thamaji/devcontainer-docker/releases/download/v1.0.1/docker \
    && chmod +x /usr/local/devcontainer-tool/bin/docker
ENV PATH=/usr/local/devcontainer-tool/bin:${PATH}
```
