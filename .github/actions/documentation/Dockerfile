FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:828611e40d69b213679cbd47ac1042eff6b5d138429f5ba686f3944492eb6e47

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
