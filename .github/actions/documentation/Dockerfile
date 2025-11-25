FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:e0eb6facd13ae75bd85e1258795f53b767847eb5d62601457d960609db0c2a7b

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
