FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:84c028923cd3c8554c9b5c1423a553a4cb8f3ee88c17a3d87756c9b08f5e8fe7

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
