#!/usr/bin/env bash
# .claude/hooks/nearest-agents-md.sh — PreToolUse hook for Edit|Write.
# Injects the nearest AGENTS.md once per directory per session.
input=$(cat)
path=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty')
[ -z "$path" ] && exit 0
dir=$(dirname "$path")
root=$(cd "$(printf '%s' "$input" | jq -r '.cwd // "."')" && pwd)
while true; do
  if [ -f "$dir/AGENTS.md" ]; then
    marker="${TMPDIR:-/tmp}/agents-md-seen-${CLAUDE_SESSION_ID:-nosession}-$(printf '%s' "$dir" | shasum | cut -c1-12)"
    [ -f "$marker" ] && exit 0
    touch "$marker"
    jq -n --rawfile c "$dir/AGENTS.md" --arg d "$dir" \
      '{hookSpecificOutput:{hookEventName:"PreToolUse",additionalContext:("Instructions covering \($d):\n\n" + $c)}}'
    exit 0
  fi
  { [ "$dir" = "$root" ] || [ "$dir" = "/" ]; } && exit 0
  dir=$(dirname "$dir")
done
