FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:59f3aa83a24152eeee04c27d4cc5c2b9f50519a67acc153cdb382ac914f3d503

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
