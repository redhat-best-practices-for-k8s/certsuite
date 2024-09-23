FROM registry.access.redhat.com/ubi9/python-39:1-197.1726664308@sha256:feda42148febbbadaf2ad594a342462b7d0260428e38503c0720610193fb24b6

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
