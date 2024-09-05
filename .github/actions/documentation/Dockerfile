FROM registry.access.redhat.com/ubi9/python-39:1-197@sha256:525ab063375303f8f80aeb76ed7f4b6e9226677afbb806964dff4ffcb6ba38d2

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
