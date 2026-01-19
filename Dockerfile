FROM --platform=$BUILDPLATFORM registry.access.redhat.com/ubi9/ubi:9.7@sha256:2c9bb68a869abf7d7417f6639509ab5eb8500d8429ea11ab59e677be5545162b AS build
ENV CERTSUITE_DIR=/usr/certsuite
ENV \
	CERTSUITE_SRC_DIR=${CERTSUITE_DIR}/src \
	TEMP_DIR=/tmp

# Install dependencies
# hadolint ignore=DL3041
RUN \
	mkdir ${CERTSUITE_DIR} \
	&& dnf update --assumeyes --disableplugin=subscription-manager --nobest \
	&& dnf install --assumeyes --disableplugin=subscription-manager \
		gcc \
		git \
		jq \
		cmake \
		wget \
	&& dnf clean all --assumeyes --disableplugin=subscription-manager \
	&& rm -rf /var/cache/yum

# Install Go binary and set the PATH
ENV \
	GO_DL_URL=https://golang.org/dl \
	GOPATH=/root/go
ENV GO_BIN_URL_x86_64=${GO_DL_URL}/go1.25.6.linux-amd64.tar.gz
ENV GO_BIN_URL_aarch64=${GO_DL_URL}/go1.25.6.linux-arm64.tar.gz

# Determine the CPU architecture and download the appropriate Go binary
# We only build our binaries on x86_64 and aarch64 platforms, so it is not necessary
# to support other architectures.
RUN \
	if [ "$(uname -m)" = x86_64 ]; then \
		wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_x86_64} --quiet \
		&& rm -rf /usr/local/go \
		&& tar -C /usr/local -xzf ${TEMP_DIR}/go1.25.6.linux-amd64.tar.gz; \
	elif [ "$(uname -m)" = aarch64 ]; then \
		wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_aarch64} --quiet \
		&& rm -rf /usr/local/go \
		&& tar -C /usr/local -xzf ${TEMP_DIR}/go1.25.6.linux-arm64.tar.gz; \
	else \
		echo "CPU architecture is not supported." && exit 1; \
	fi
ENV PATH=${PATH}:"/usr/local/go/bin":${GOPATH}/"bin"


# Set environment specific variables
ENV \
	OPERATOR_SDK_X86_FILENAME=operator-sdk_linux_amd64 \
	OPERATOR_SDK_ARM_FILENAME=operator-sdk_linux_arm64 \
	OPERATOR_SDK_PPC64LE_FILENAME=operator-sdk_linux_ppc64le \
	OPERATOR_SDK_S390X_FILENAME=operator-sdk_linux_s390x

# Download operator-sdk binary
ENV \
	OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.41.1 \
	OSDK_BIN=/usr/local/osdk/bin

RUN \
	mkdir -p ${OSDK_BIN}

ARG TARGETARCH
ARG TARGETOS
ARG TARGETPLATFORM

RUN \
 # echo the architecture for debugging
 echo "TARGETARCH: $TARGETARCH" \
 && echo "TARGETOS: $TARGETOS" \
 && echo "TARGETPLATFORM: $TARGETPLATFORM"

# hadolint ignore=DL4001
RUN \
	if [ "$TARGETARCH" = x86_64 ] || [ "$TARGETARCH" = amd64 ]; then \
		curl \
			--location \
			--remote-name \
			${OPERATOR_SDK_DL_URL}/${OPERATOR_SDK_X86_FILENAME} \
			&& mv ${OPERATOR_SDK_X86_FILENAME} ${OSDK_BIN}/operator-sdk \
			&& chmod +x ${OSDK_BIN}/operator-sdk; \
	elif [ "$TARGETARCH" = aarch64 ] || [ "$TARGETARCH" = arm64 ]; then \
		curl \
			--location \
			--remote-name \
			${OPERATOR_SDK_DL_URL}/${OPERATOR_SDK_ARM_FILENAME} \
			&& mv ${OPERATOR_SDK_ARM_FILENAME} ${OSDK_BIN}/operator-sdk \
			&& chmod +x ${OSDK_BIN}/operator-sdk; \
	elif [ "$TARGETARCH" = ppc64le ]; then \
		curl \
			--location \
			--remote-name \
			${OPERATOR_SDK_DL_URL}/${OPERATOR_SDK_PPC64LE_FILENAME} \
			&& mv ${OPERATOR_SDK_PPC64LE_FILENAME} ${OSDK_BIN}/operator-sdk \
			&& chmod +x ${OSDK_BIN}/operator-sdk; \
	elif [ "$TARGETARCH" = s390x ]; then \
		curl \
			--location \
			--remote-name \
			${OPERATOR_SDK_DL_URL}/${OPERATOR_SDK_S390X_FILENAME} \
			&& mv ${OPERATOR_SDK_S390X_FILENAME} ${OSDK_BIN}/operator-sdk \
			&& chmod +x ${OSDK_BIN}/operator-sdk; \
	else \
		echo "CPU architecture is not supported." && exit 1; \
	fi

# Copy all of the files into the source directory and then switch contexts
COPY . ${CERTSUITE_SRC_DIR}
WORKDIR ${CERTSUITE_SRC_DIR}

# Build the certsuite binary and clean up unnecessary files in a single step
RUN \
	export GOARCH=$TARGETARCH \
	&& export GOOS=$TARGETOS \
	&& make build-certsuite-tool \
	&& cp certsuite ${CERTSUITE_DIR} \
	&& dnf remove --assumeyes --disableplugin=subscription-manager gcc git wget \
	&& dnf clean all --assumeyes --disableplugin=subscription-manager \
	&& rm -rf ${CERTSUITE_SRC_DIR} ${TEMP_DIR} /root/.cache /root/go/pkg /root/go/src \
		/usr/lib/golang/pkg /usr/lib/golang/src /var/cache/yum /usr/local/go /usr/local/osdk/bin/*

# Switch contexts back to the root CERTSUITE directory
WORKDIR ${CERTSUITE_DIR}

# Using latest is prone to errors.
# hadolint ignore=DL3007
FROM quay.io/redhat-best-practices-for-k8s/oct:latest AS db

# Copy the state into a new flattened image to reduce size.
# TODO run as non-root
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.7@sha256:90bd85dcd061d1ad6dbda70a867c41958c04a86462d05c631f8205e8870f28f8

ENV \
	CERTSUITE_DIR=/usr/certsuite \
	OSDK_BIN=/usr/local/osdk/bin

# Install the certsuite binary
COPY --from=build ${CERTSUITE_DIR}/certsuite /usr/local/bin/certsuite

# Add operatorsdk binary to image
COPY --from=build ${OSDK_BIN} /usr/local/bin/operator-sdk

# Update the CNF containers, helm charts and operators DB
ENV \
	CERTSUITE_OFFLINE_DB=/usr/offline-db \
	OCT_DB_PATH=/usr/oct/cmd/tnf/fetch
COPY --from=db ${OCT_DB_PATH} ${CERTSUITE_OFFLINE_DB}

WORKDIR ${CERTSUITE_DIR}
ENV SHELL=/bin/bash
CMD ["certsuite", "-h"]
