FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:08d36504594394fc17d515cf89bf5d4acc08d5fd42afea2d7b6aa478f6a95f54

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
