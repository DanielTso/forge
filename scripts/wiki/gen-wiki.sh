#!/usr/bin/env bash
# Regenerate the GitHub wiki pages from real forge output.
#
#   scripts/wiki/gen-wiki.sh <output-dir>
#
# Needs: forge on PATH (make install), python3, tree.
# Publish: git clone git@github.com:DanielTso/forge.wiki.git, copy <output-dir>/*.md
# into it, review the diff, commit, push.
set -euo pipefail
out=${1:?usage: gen-wiki.sh <output-dir>}
here=$(cd "$(dirname "$0")" && pwd)
command -v forge >/dev/null || { echo "forge not on PATH (run make install)" >&2; exit 1; }
command -v tree >/dev/null || { echo "tree not installed" >&2; exit 1; }

gen=$(mktemp -d)
trap 'rm -rf "$gen"' EXIT
for t in go-cli go-service go-lib python node typescript deno bun java rust gleam zig generic; do
  (cd "$gen" && forge new "my-$t" -t "$t" --no-git -y -m "github.com/you/my-$t" >/dev/null && tree -a --dirsfirst "my-$t" > "$t.tree")
done
mkdir -p "$out"
python3 "$here/mkwiki.py" "$out" "$gen"
echo "Wiki pages written to $out — remember to update version numbers in mkwiki.py (Command-Reference header, examples) after a release."
