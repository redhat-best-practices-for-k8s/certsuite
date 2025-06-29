FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:a49364adae4cd27b1b4c469174f9bd4988ea4191c4dcb7da5a2d44513e590ba7

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
