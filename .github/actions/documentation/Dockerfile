FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:c5180b2ecc2f662d614406498893b43e6914bf419e5605c0fe973cc610ea7369

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
