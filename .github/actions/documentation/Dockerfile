FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:0758ea0a2e5f931c33635a37592ea8672b998a58af5f032463852b296242f510

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
