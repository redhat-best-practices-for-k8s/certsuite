FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:9a31f03f8b27d9065c3488bbd3650c67271c3b868eacf816ddea07ababd9fbc0

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
