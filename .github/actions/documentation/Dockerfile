FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:796fe7f61b18571300928cc27595aa648c2d03f61894ff4087d9bc6ca3a50143

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
