FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:9bbc4cfeac896544ab3eafa088c3d6995e82592362d374606f00d221f2986fe0

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
