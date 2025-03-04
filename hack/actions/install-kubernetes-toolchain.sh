#! /usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

readonly KUSTOMIZE_VERS="v3.8.6"
readonly KUBECTL_VERS="v1.20.4"
readonly KIND_VERS="v0.10.0"
readonly SONOBUOY_VERS="0.19.0"

readonly PROGNAME=$(basename $0)
readonly CURL=${CURL:-curl}

# Google storage is case sensitive, so we we need to lowercase the OS.
readonly OS=$(uname | tr '[:upper:]' '[:lower:]')

usage() {
  echo "Usage: $PROGNAME INSTALLDIR"
}

download() {
    local -r url="$1"
    local -r target="$2"

    echo Downloading "$target" from "$url"
    ${CURL} --progress-bar --location  --output "$target" "$url"
}

case "$#" in
  "1")
    mkdir -p "$1"
    readonly DESTDIR=$(cd "$1" && pwd)
    ;;
  *)
    usage
    exit 64
    ;;
esac

# TODO: Remove once upstream images are available (#3610).
# Install latest version of kind.
go get sigs.k8s.io/kind@master

# Move the $GOPATH/bin/kind binary to local since Github actions
# have their own version installed.
mv /home/runner/go/bin/kind ${DESTDIR}/kind

# Install ginkgo CLI
go get github.com/onsi/ginkgo/...
mv /home/runner/go/bin/ginkgo ${DESTDIR}/ginkgo

# Uncomment this once v0.11 of Kind is released.
#download \
#    "https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERS}/kind-${OS}-amd64" \
#    "${DESTDIR}/kind"
#
#chmod +x  "${DESTDIR}/kind"

download \
    "https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERS}/bin/${OS}/amd64/kubectl" \
    "${DESTDIR}/kubectl"

chmod +x "${DESTDIR}/kubectl"

download \
    "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERS}/kustomize_${KUSTOMIZE_VERS}_${OS}_amd64.tar.gz" \
    "${DESTDIR}/kustomize.tgz"

tar -C "${DESTDIR}" -xf "${DESTDIR}/kustomize.tgz" kustomize
rm "${DESTDIR}/kustomize.tgz"

readonly INTEGRATION_TESTER_PATH="$(curl -s -L https://github.com/projectcontour/integration-tester/releases/latest | grep href | grep -i "integration-tester_.*${OS}_x86_64\.tar\.gz" | cut -d '"' -f2)"
download \
    "https://github.com${INTEGRATION_TESTER_PATH}" \
    "${DESTDIR}/integration-tester.tgz"

tar -C "${DESTDIR}" -xf "${DESTDIR}/integration-tester.tgz"
rm "${DESTDIR}/integration-tester.tgz"

download \
    "https://github.com/vmware-tanzu/sonobuoy/releases/download/v${SONOBUOY_VERS}/sonobuoy_${SONOBUOY_VERS}_linux_amd64.tar.gz" \
    "${DESTDIR}/sonobuoy.tgz"

tar -C "${DESTDIR}" -xf "${DESTDIR}/sonobuoy.tgz" sonobuoy
rm "${DESTDIR}/sonobuoy.tgz"
