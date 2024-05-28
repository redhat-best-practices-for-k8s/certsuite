FROM registry.access.redhat.com/ubi9/python-39:1-182.1716478612@sha256:864e737b8cb4166001529f51396f03dbec55b634edb0b099c1b1484a0762deb9

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
