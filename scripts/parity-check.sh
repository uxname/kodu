#!/usr/bin/env bash
# Parity check of the Go binary against the reference Node build.
# Usage: scripts/parity-check.sh [path-to-go-binary] [path-to-node-main.js]
# Requires built dist/kodu (Go) and dist/src/main.js (Node).
set -u

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
KODU="${1:-$ROOT/dist/kodu}"
NODE_MAIN="${2:-$ROOT/dist/src/main.js}"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

pass=0; fail=0
ok()  { echo "  ✅ $1"; pass=$((pass+1)); }
bad() { echo "  ❌ $1"; fail=$((fail+1)); }

node_run() { node "$NODE_MAIN" "$@"; }
go_run()   { "$KODU" "$@"; }

# Test project.
FX="$WORK/proj"; mkdir -p "$FX/src/sub"
cat > "$FX/a.ts" <<'EOF'
// line
const url = "http://x//y"; // trailing
/** jsdoc */
export const a = 1;
import './src/b';
EOF
cat > "$FX/src/b.tsx" <<'EOF'
export const B = () => <div>{/* c */}</div>; // t
EOF
cat > "$FX/src/sub/c.js" <<'EOF'
const re = /\/\*x\*\//g; // c
EOF
printf 'node_modules\n' > "$FX/.gitignore"
mkdir -p "$FX/node_modules/d"; echo x > "$FX/node_modules/d/i.js"
printf '\x00bin' > "$FX/logo.png"

cmp_cmd() { # name; rest=args
  local name="$1"; shift
  ( cd "$FX" && node_run "$@" -o "$WORK/n.out" >/dev/null 2>&1 )
  ( cd "$FX" && go_run   "$@" -o "$WORK/g.out" >/dev/null 2>&1 )
  if cmp -s "$WORK/n.out" "$WORK/g.out"; then ok "$name"; else bad "$name"; diff "$WORK/n.out" "$WORK/g.out" | head -10; fi
}

echo "== pack =="
cmp_cmd "pack -f xml"        pack -f xml
cmp_cmd "pack -f text"       pack -f text
cmp_cmd "pack --clean xml"   pack -f xml --clean
cmp_cmd "pack -p src -f xml" pack -p src -f xml

echo "== pack -l (file order) =="
if diff <(cd "$FX" && node_run pack -l 2>/dev/null | sed 's/^ℹ //') <(cd "$FX" && go_run pack -l 2>/dev/null) >/dev/null; then
  ok "pack -l list matches"; else bad "pack -l"; fi

# stdin is always treated as stdin.ts, so we compare on TS/JS input
# (TSX via stdin is already a format mismatch; see the cleaner docs).
echo "== clean --stdin =="
for f in "$FX/a.ts" "$FX/src/sub/c.js"; do
  if diff <(node_run clean --stdin < "$f" 2>/dev/null) <(go_run clean --stdin < "$f" 2>/dev/null) >/dev/null; then
    ok "clean --stdin $(basename "$f")"; else bad "clean --stdin $(basename "$f")"; fi
done

echo "== summary: $pass ok, $fail fail =="
exit $((fail > 0 ? 1 : 0))
