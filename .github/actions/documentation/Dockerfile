FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:8392799f609b0de3f9a4640400d460f5e2563b2b6f09e6b5fe89a67adda75c6a

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
