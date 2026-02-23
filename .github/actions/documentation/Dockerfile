FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:367b7bfa98e2692c79369620eda1b33a821d625cbfecc7e6dc5262f03f53b9a2

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
