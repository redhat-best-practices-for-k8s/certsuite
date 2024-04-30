FROM registry.access.redhat.com/ubi9/python-39:1-182@sha256:ccfe85751df5091531e93d8b2cd2986535367947b64438e2fcf81d072f2f8ad5

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
