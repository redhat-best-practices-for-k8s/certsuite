FROM registry.access.redhat.com/ubi9/ubi:9.4-947.1714667021@sha256:ed84f34cd929ea6b0c247b6daef54dd79602804a32480a052951021caf429494 AS build
ENV TNF_DIR=/usr/tnf
ENV \
	TNF_SRC_DIR=${TNF_DIR}/tnf-src \
	TNF_BIN_DIR=${TNF_DIR}/cnf-certification-test \
	TEMP_DIR=/tmp

# Install dependencies
# hadolint ignore=DL3041
RUN \
	mkdir ${TNF_DIR} \
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
ENV GO_BIN_URL_x86_64=${GO_DL_URL}/go1.22.3.linux-amd64.tar.gz
ENV GO_BIN_URL_aarch64=${GO_DL_URL}/go1.22.3.linux-arm64.tar.gz

# Determine the CPU architecture and download the appropriate Go binary
RUN \
	if [ "$(uname -m)" = x86_64 ]; then \
		wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_x86_64} --quiet \
		&& rm -rf /usr/local/go \
		&& tar -C /usr/local -xzf ${TEMP_DIR}/go1.22.3.linux-amd64.tar.gz; \
	elif [ "$(uname -m)" = aarch64 ]; then \
		wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_aarch64} --quiet \
		&& rm -rf /usr/local/go \
		&& tar -C /usr/local -xzf ${TEMP_DIR}/go1.22.3.linux-arm64.tar.gz; \
	else \
		echo "CPU architecture is not supported." && exit 1; \
	fi
ENV PATH=${PATH}:"/usr/local/go/bin":${GOPATH}/"bin"

# Download operator-sdk binary
ENV \
	OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.34.1 \
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
COPY . ${TNF_SRC_DIR}
WORKDIR ${TNF_SRC_DIR}
RUN make build-cnf-tests build-tnf-tool

# Extract what's needed to run at a separate location
# Quote this to prevent word splitting.
# hadolint ignore=SC2046
RUN \
	mkdir ${TNF_BIN_DIR} \
	&& cp run-cnf-suites.sh ${TNF_DIR} \
	# copy all JSON files to allow tests to run
	&& cp --parents $(find . -name '*.json*') ${TNF_DIR} \
	&& cp cnf-certification-test/cnf-certification-test ${TNF_BIN_DIR} \
	# copy the tnf command binary
	&& cp tnf ${TNF_BIN_DIR}

# Switch contexts back to the root TNF directory
WORKDIR ${TNF_DIR}

# Remove most of the build artefacts
RUN \
	dnf remove --assumeyes --disableplugin=subscription-manager gcc git wget \
	&& dnf clean all --assumeyes --disableplugin=subscription-manager \
	&& rm -rf ${TNF_SRC_DIR} \
	&& rm -rf ${TEMP_DIR} \
	&& rm -rf /root/.cache \
	&& rm -rf /root/go/pkg \
	&& rm -rf /root/go/src \
	&& rm -rf /usr/lib/golang/pkg \
	&& rm -rf /usr/lib/golang/src

# Using latest is prone to errors.
# hadolint ignore=DL3007
FROM quay.io/testnetworkfunction/oct:latest AS db

# Copy the state into a new flattened image to reduce size.
# TODO run as non-root
FROM registry.access.redhat.com/ubi9/ubi-minimal:9.4-949.1714662671@sha256:2636170dc55a0931d013014a72ae26c0c2521d4b61a28354b3e2e5369fa335a3

ENV \
	TNF_DIR=/usr/tnf \
	OSDK_BIN=/usr/local/osdk/bin

# Copy all of the necessary files over from the TNF_DIR
COPY --from=build ${TNF_DIR} ${TNF_DIR}

# Add operatorsdk binary to image
COPY --from=build ${OSDK_BIN} ${OSDK_BIN}

# Update the CNF containers, helm charts and operators DB
ENV \
	TNF_OFFLINE_DB=/usr/offline-db \
	OCT_DB_PATH=/usr/oct/cmd/tnf/fetch
COPY --from=db ${OCT_DB_PATH} ${TNF_OFFLINE_DB}

ENV TNF_BIN_DIR=${TNF_DIR}/cnf-certification-test

ENV \
	TNF_CONFIGURATION_PATH=/usr/tnf/config/tnf_config.yml \
	KUBECONFIG=/usr/tnf/kubeconfig/config \
	PFLT_DOCKERCONFIG=/usr/tnf/dockercfg/config.json \
	PATH="${OSDK_BIN}:${TNF_BIN_DIR}:${PATH}"
WORKDIR ${TNF_DIR}
ENV SHELL=/bin/bash
CMD ["./run-cnf-suites.sh", "-o", "claim", "-f", "diagnostic"]
