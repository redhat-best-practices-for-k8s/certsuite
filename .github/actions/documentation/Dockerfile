FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:4e9daa22c8b0bca06b15160e4e3edb87b219c750d75a64ff4ebbbe81c0fb8853

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
