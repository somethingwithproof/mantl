#!/usr/bin/env bash
# Exercise native package installation, reinstallation and removal in disposable containers.
set -euo pipefail

usage() {
  echo "Usage: $0 ARTIFACT_DIRECTORY [amd64|arm64]" >&2
}

if [[ ${1:-} == --help ]]; then
  usage
  exit 0
fi
if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 2
fi
artifact_dir=$(cd -- "$1" && pwd -P)
architecture=${2:-$(docker info --format '{{.Architecture}}')}
case "$architecture" in
  amd64|x86_64) architecture=amd64 ;;
  arm64|aarch64) architecture=arm64 ;;
  *) echo "Unsupported package architecture: $architecture" >&2; exit 2 ;;
esac

shopt -s nullglob
archives=("$artifact_dir"/mantl_*_linux_"$architecture".tar.gz)
debs=("$artifact_dir"/mantl_*_linux_"$architecture".deb)
rpms=("$artifact_dir"/mantl_*_linux_"$architecture".rpm)
if [[ ${#archives[@]} != 1 || ${#debs[@]} != 1 || ${#rpms[@]} != 1 ]]; then
  echo "Require exactly one CLI archive, DEB and RPM for $architecture" >&2
  exit 2
fi
for file in "${archives[0]}" "${debs[0]}" "${rpms[0]}"; do
  if [[ ! -f "$file" || -L "$file" ]]; then
    echo "Package artifacts must be regular files: $file" >&2
    exit 2
  fi
done

for kind in deb rpm; do
  capabilities=(--cap-drop ALL)
  case "$kind" in
    deb)
      case "$architecture" in
        amd64) image=ubuntu:24.04@sha256:f610ab94648195aa356059f5b41d6085c9d4d903c072430cdd1af7bdb646106b ;;
        arm64) image=ubuntu:24.04@sha256:08571ca13e00ca07a2a84eab83a959b4242e22cceb16486a11bef1428c9e93a7 ;;
      esac
      package=${debs[0]}
      ;;
    rpm)
      case "$architecture" in
        amd64) image=rockylinux:9@sha256:d644d203142cd5b54ad2a83a203e1dee68af2229f8fe32f52a30c6e1d3c3a9e0 ;;
        arm64) image=rockylinux:9@sha256:370b6bd1851d5023c5c673535c85cdc5c1d8a59416ad83380a8db7ce3691bd45 ;;
      esac
      package=${rpms[0]}
      # Rocky's system directories are mode 0555; RPM needs this filesystem
      # capability to install there. Keep all other capabilities dropped.
      capabilities+=(--cap-add DAC_OVERRIDE)
      ;;
  esac
  docker pull --platform "linux/$architecture" "$image"
  docker run --rm -i --platform "linux/$architecture" --network none \
    "${capabilities[@]}" --security-opt no-new-privileges --pids-limit 128 --memory 1g \
    --mount "type=bind,source=$artifact_dir,target=/artifacts,readonly" \
    --env "PACKAGE_KIND=$kind" --env "PACKAGE_FILE=$(basename "$package")" \
    --env "ARCHIVE_FILE=$(basename "${archives[0]}")" --env "PACKAGE_ARCH=$architecture" \
    "$image" bash -euo pipefail -s <<'CONTAINER'
if command -v mantl; then
  echo "Validation image already contains Mantl" >&2
  exit 1
fi
mkdir /tmp/mantl-archive
tar --no-same-owner -xzf "/artifacts/$ARCHIVE_FILE" -C /tmp/mantl-archive mantl LICENSE docs/releases.md
expected_version=$(/tmp/mantl-archive/mantl --version)
case "$expected_version" in
  'mantl version '*) ;;
  *) echo "Unexpected CLI version output" >&2; exit 1 ;;
esac
if [[ "$PACKAGE_KIND" == deb ]]; then
  test "$(dpkg-deb --field "/artifacts/$PACKAGE_FILE" Package)" = mantl
  test "$(dpkg-deb --field "/artifacts/$PACKAGE_FILE" Architecture)" = "$PACKAGE_ARCH"
else
  rpm_arch=x86_64
  if [[ "$PACKAGE_ARCH" == arm64 ]]; then rpm_arch=aarch64; fi
  test "$(rpm -qp --queryformat '%{NAME}' "/artifacts/$PACKAGE_FILE")" = mantl
  test "$(rpm -qp --queryformat '%{ARCH}' "/artifacts/$PACKAGE_FILE")" = "$rpm_arch"
fi
install_package() {
  if [[ "$PACKAGE_KIND" == deb ]]; then
    # Minimal Ubuntu images exclude documentation by default. Include only this
    # package's documentation so its installed contents can be verified.
    dpkg --path-include=/usr/share/doc/mantl --path-include='/usr/share/doc/mantl/*' \
      --install "/artifacts/$PACKAGE_FILE"
  else
    rpm --upgrade --replacepkgs "/artifacts/$PACKAGE_FILE"
  fi
}
file_hash() {
  sha256sum "$1" | cut -d ' ' -f1
}
verify_install() {
  test "$(command -v mantl)" = /usr/bin/mantl
  test "$(mantl --version)" = "$expected_version"
  test "$(file_hash /tmp/mantl-archive/mantl)" = "$(file_hash /usr/bin/mantl)"
  test "$(file_hash /tmp/mantl-archive/LICENSE)" = "$(file_hash /usr/share/licenses/mantl/LICENSE)"
  test "$(file_hash /tmp/mantl-archive/docs/releases.md)" = "$(file_hash /usr/share/doc/mantl/releases.md)"
  mantl plan --help >/dev/null
  mantl compliance --help >/dev/null
}
install_package
verify_install
install_package
verify_install
if [[ "$PACKAGE_KIND" == deb ]]; then
  dpkg --purge mantl
  if dpkg --status mantl >/dev/null 2>&1; then exit 1; fi
else
  rpm --erase mantl
  if rpm --query mantl >/dev/null 2>&1; then exit 1; fi
fi
hash -r
if command -v mantl; then exit 1; fi
test ! -e /usr/bin/mantl
test ! -e /usr/share/licenses/mantl/LICENSE
test ! -e /usr/share/doc/mantl/releases.md
printf 'Package lifecycle passed: %s %s (%s)\n' "$PACKAGE_KIND" "$PACKAGE_ARCH" "$expected_version"
CONTAINER
done
