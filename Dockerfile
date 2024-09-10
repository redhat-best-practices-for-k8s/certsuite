FROM registry.access.redhat.com/ubi9/ubi:9.4-1214.1725849297@sha256:7575b6e3cc492f856daf8c43f30692d8f5fcd5b7077806dba4bac436ad0a84e8 AS build
ENV CERTSUITE_DIR=/usr/certsuite
ENV \
	CERTSUITE_SRC_DIR=${CERTSUITE_DIR}/src \
	TEMP_DIR=/tmp

# Install dependencies
# hadolint ignore=DL3041
RUN \
	mkdir ${CERTSUITE_DIR} \
	&& dnf update --assumeyes --disableplugin=subscription-manager \
	&& dnf install --assumeyes --disableplugin=subscription-manager \
		gcc \
		git \
		jq \
		cmake \
		wget \
	&& dnf clean all --assumeyes --disableplugin=subscription-manager \
	&& rm -rf /var/cache/yum

# Set environment specific variables
ENV \
	OPERATOR_SDK_X86_FILENAME=operator-sdk_linux_amd64 \
	OPERATOR_SDK_ARM_FILENAME=operator-sdk_linux_arm64

# Install Go binary and set the PATH
ENV \
	GO_DL_URL=https://golang.org/dl \
	GOPATH=/root/go
ENV GO_BIN_URL_x86_64=${GO_DL_URL}/go1.23.1.linux-amd64.tar.gz
ENV GO_BIN_URL_aarch64=${GO_DL_URL}/go1.23.1.linux-arm64.tar.gz

# Determine the CPU architecture and download the appropriate Go binary
RUN \
	if [ "$(uname -m)" = x86_64 ]; then \
		wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_x86_64} --quiet \
		&& rm -rf /usr/local/go \
		&& tar -C /usr/local -xzf ${TEMP_DIR}/go1.23.1.linux-amd64.tar.gz; \
	elif [ "$(uname -m)" = aarch64 ]; then \
		wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_aarch64} --quiet \
		&& rm -rf /usr/local/go \
		&& tar -C /usr/local -xzf ${TEMP_DIR}/go1.23.1.linux-arm64.tar.gz; \
	else \
		echo "CPU architecture is not supported." && exit 1; \
	fi
ENV PATH=${PATH}:"/usr/local/go/bin":${GOPATH}/"bin"

# Download operator-sdk binary
ENV \
	OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.36.1 \
	OSDK_BIN=/usr/local/osdk/bin

RUN \
	mkdir -p ${OSDK_BIN}

# hadolint ignore=DL4001
RUN \
	if [ "$(uname -m)" = x86_64 ]; then \
		curl \
			--location \
			--remote-name \
			${OPERATOR_SDK_DL_URL}/${OPERATOR_SDK_X86_FILENAME} \
			&& mv ${OPERATOR_SDK_X86_FILENAME} ${OSDK_BIN}/operator-sdk \
			&& chmod +x ${OSDK_BIN}/operator-sdk; \
	elif [ "$(uname -m)" = aarch64 ]; then \
		curl \
			--location \
			--remote-name \
			${OPERATOR_SDK_DL_URL}/${OPERATOR_SDK_ARM_FILENAME} \
			&& mv ${OPERATOR_SDK_ARM_FILENAME} ${OSDK_BIN}/operator-sdk \
			&& chmod +x ${OSDK_BIN}/operator-sdk; \
	else \
		echo "CPU architecture is not supported." && exit 1; \
	fi

# Copy all of the files into the source directory and then switch contexts
COPY . ${CERTSUITE_SRC_DIR}
WORKDIR ${CERTSUITE_SRC_DIR}

# Build the certsuite binary
RUN make build-certsuite-tool \
	&& cp certsuite ${CERTSUITE_DIR}

# Switch contexts back to the root CERTSUITE directory
WORKDIR ${CERTSUITE_DIR}

# Remove most of the build artefacts
RUN \
	dnf remove --assumeyes --disableplugin=subscription-manager gcc git wget \
	&& dnf clean all --assumeyes --disableplugin=subscription-manager \
	&& rm -rf ${CERTSUITE_SRC_DIR} \
	&& rm -rf ${TEMP_DIR} \
	&& rm -rf /root/.cache \
	&& rm -rf /root/go/pkg \
	&& rm -rf /root/go/src \
	&& rm -rf /usr/lib/golang/pkg \
	&& rm -rf /usr/lib/golang/src

# Using latest is prone to errors.
# hadolint ignore=DL3007
FROM quay.io/redhat-best-practices-for-k8s/oct:latest AS db

# Copy the state into a new flattened image to reduce size.
# TODO run as non-root
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.4-1227@sha256:f182b500ff167918ca1010595311cf162464f3aa1cab755383d38be61b4d30aa

ENV \
	CERTSUITE_DIR=/usr/certsuite \
	OSDK_BIN=/usr/local/osdk/bin

# Install the certsuite binary
COPY --from=build ${CERTSUITE_DIR} ${CERTSUITE_DIR}
RUN cp ${CERTSUITE_DIR}/certsuite /usr/local/bin

# Add operatorsdk binary to image
COPY --from=build ${OSDK_BIN} ${OSDK_BIN}

# Update the CNF containers, helm charts and operators DB
ENV \
	CERTSUITE_OFFLINE_DB=/usr/offline-db \
	OCT_DB_PATH=/usr/oct/cmd/tnf/fetch
COPY --from=db ${OCT_DB_PATH} ${CERTSUITE_OFFLINE_DB}


ENV PATH="${OSDK_BIN}:${PATH}"
WORKDIR ${CERTSUITE_DIR}
ENV SHELL=/bin/bash
CMD ["certsuite", "-h"]
