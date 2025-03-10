FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:4b1db7c433eda5143a7ca9c1e609cf77c7762d92d2baed08de3c91972890699e

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
