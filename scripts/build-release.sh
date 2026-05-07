#!/usr/bin/env sh
set -eu

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
dist_dir="$root_dir/dist"
tmp_dir="$root_dir/.tmp-release"

rm -rf "$dist_dir" "$tmp_dir"
mkdir -p "$dist_dir" "$tmp_dir"

cd "$root_dir"

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o "$tmp_dir/changeloger-darwin-amd64" ./cmd/changeloger
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o "$tmp_dir/changeloger-darwin-arm64" ./cmd/changeloger

if ! command -v lipo >/dev/null 2>&1; then
    printf '%s\n' "Ошибка: для changeloger-darwin-universal нужен lipo. Запусти сборку на macOS." >&2
    exit 1
fi

lipo -create \
    "$tmp_dir/changeloger-darwin-amd64" \
    "$tmp_dir/changeloger-darwin-arm64" \
    -output "$dist_dir/changeloger-darwin-universal"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$dist_dir/changeloger-linux-amd64" ./cmd/changeloger

cp "$root_dir/changeloger-install" "$dist_dir/changeloger-install"
chmod 0755 "$dist_dir/changeloger-"*

(
    cd "$dist_dir"
    shasum -a 256 changeloger-* > checksums.txt
)

rm -rf "$tmp_dir"

printf '%s\n' "Release artifacts:"
find "$dist_dir" -maxdepth 1 -type f -print | sort
