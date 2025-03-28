FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:dd3722b119e6d79cdd1a1e9d3a4414f7cc2ae9ef674cc27076e73234cc61758e

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
