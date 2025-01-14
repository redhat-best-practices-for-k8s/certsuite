FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:36ae15329ade62cc7082d44db67f94494bf3836edb8b7cf733f7519168a6e9de

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
