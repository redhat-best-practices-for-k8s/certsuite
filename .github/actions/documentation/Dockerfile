FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:86d15db3a8cbe9d01a483f84b6d2a5db4eb1881710abbe0b57ec1ce496903cf1

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
