FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:c871688f22d2ac783a37703b1944fd94d8d2707aac8a3363fd1a039141b365f6

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
