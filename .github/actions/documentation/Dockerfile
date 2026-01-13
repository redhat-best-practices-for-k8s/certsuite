FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:54ce28c1544eac8a1779841c7590a597c42609b96454a13bf0ba20438fd17f71

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
