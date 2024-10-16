FROM registry.access.redhat.com/ubi9/python-39:1-197.1726696853@sha256:12644d1a2d214bd1be7eac3a9b1e983a987d9452b78e4c6be9c863d5038b9338

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
