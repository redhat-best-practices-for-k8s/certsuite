FROM registry.access.redhat.com/ubi9/python-39:1-197.1725907694@sha256:278ae38e8f28ccba3cb7cd542f684d739a84f771e418fc8018d07a522205b05c

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
