FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:b28387e6c2c66bf2092748bd618036ac48b287ba042f8f353d47118480d0eaf6

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
