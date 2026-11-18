#!/bin/sh
set -eu

repo="ludanortmun/peco"
install_dir="${PECO_INSTALL_DIR:-$HOME/.local/bin}"

case "$(uname -s):$(uname -m)" in
  Darwin:arm64)
    goos="darwin"
    goarch="arm64"
    ;;
  Linux:x86_64|Linux:amd64)
    goos="linux"
    goarch="amd64"
    ;;
  *)
    printf 'Unsupported platform: %s/%s\n' "$(uname -s)" "$(uname -m)" >&2
    exit 1
    ;;
esac

for required in curl tar awk mktemp; do
  if ! command -v "$required" >/dev/null 2>&1; then
    printf 'Required command not found: %s\n' "$required" >&2
    exit 1
  fi
done
if ! command -v sha256sum >/dev/null 2>&1 &&
   ! command -v shasum >/dev/null 2>&1; then
  printf 'Required command not found: sha256sum or shasum\n' >&2
  exit 1
fi

# GitHub's latest-release redirect excludes prereleases.
latest_url=$(curl -sSL -o /dev/null -w '%{url_effective}' \
  "https://github.com/$repo/releases/latest")
tag=${latest_url##*/}
case "$tag" in
  v[0-9]*) ;;
  *)
    # Public releases are newest first; drafts are not returned anonymously.
    releases=$(curl -fsSL "https://api.github.com/repos/$repo/releases?per_page=1")
    tag=$(printf '%s\n' "$releases" | awk '
      match($0, /"tag_name"[[:space:]]*:[[:space:]]*"[^"]+"/) {
        value = substr($0, RSTART, RLENGTH)
        sub(/^"tag_name"[[:space:]]*:[[:space:]]*"/, "", value)
        sub(/"$/, "", value)
        print value
        exit
      }
    ')
    ;;
esac
case "$tag" in
  v[0-9]*) ;;
  *)
    printf 'No published release found\n' >&2
    exit 1
    ;;
esac

version=${tag#v}
archive="peco_${version}_${goos}_${goarch}.tar.gz"
download_base="https://github.com/$repo/releases/download/$tag"

workdir=$(mktemp -d "${TMPDIR:-/tmp}/peco-install.XXXXXX")
trap 'rm -rf "$workdir"' 0
trap 'exit 1' HUP INT TERM

curl -fsSL -o "$workdir/$archive" "$download_base/$archive"
curl -fsSL -o "$workdir/SHA256SUMS" "$download_base/SHA256SUMS"

(
  cd "$workdir"
  awk -v artifact="./$archive" '
    $2 == artifact { checksum = $0; count++ }
    END { if (count != 1) exit 1; print checksum }
  ' SHA256SUMS > SHA256SUMS.selected || {
    printf 'No unique checksum found for %s\n' "$archive" >&2
    exit 1
  }
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum -c SHA256SUMS.selected
  else
    shasum -a 256 -c SHA256SUMS.selected
  fi
  tar -xzf "$archive" peco peco-server
)

mkdir -p "$install_dir"
for binary in peco peco-server; do
  cp "$workdir/$binary" "$install_dir/$binary"
  chmod 0755 "$install_dir/$binary"
  printf 'Installed %s/%s\n' "$install_dir" "$binary"
done
case ":$PATH:" in
  *":$install_dir:"*) ;;
  *) printf 'Add %s to your PATH to run peco and peco-server.\n' "$install_dir" ;;
esac
