FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:eb65add9b84a3d3ac104b7bc091f49faffe9e69aef8bafb0654547b1708c4841

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
