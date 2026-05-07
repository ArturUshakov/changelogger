#!/usr/bin/env sh
set -eu

root_dir="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
dist_dir="$root_dir/dist"
tmp_dir="$root_dir/.tmp-release"

rm -rf "$dist_dir" "$tmp_dir"
mkdir -p "$dist_dir" "$tmp_dir"

cd "$root_dir"

GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o "$tmp_dir/changelogger-darwin-amd64" ./cmd/changelogger
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o "$tmp_dir/changelogger-darwin-arm64" ./cmd/changelogger

if ! command -v lipo >/dev/null 2>&1; then
    printf '%s\n' "Ошибка: для changelogger-darwin-universal нужен lipo. Запусти сборку на macOS." >&2
    exit 1
fi

lipo -create \
    "$tmp_dir/changelogger-darwin-amd64" \
    "$tmp_dir/changelogger-darwin-arm64" \
    -output "$dist_dir/changelogger-darwin-universal"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$dist_dir/changelogger-linux-amd64" ./cmd/changelogger

cp "$root_dir/changelogger-install" "$dist_dir/changelogger-install"
chmod 0755 "$dist_dir/changelogger-"*

(
    cd "$dist_dir"
    shasum -a 256 changelogger-* > checksums.txt
)

rm -rf "$tmp_dir"

printf '%s\n' "Release artifacts:"
find "$dist_dir" -maxdepth 1 -type f -print | sort
