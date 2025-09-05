FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:cc46d2979862d93e9da4d021e9100c7feb542dac3157acb2715b8baf54730c2b

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
