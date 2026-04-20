FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:ee2a11266ec90665de33302f74d913d556cf2dc67d68208cada62de15a6859b9

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
