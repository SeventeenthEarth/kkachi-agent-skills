#!/bin/sh
set -eu

prompt_file=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --prompt-file)
      shift
      if [ "$#" -eq 0 ]; then
        echo "mar-zcode: --prompt-file requires a path" >&2
        exit 64
      fi
      prompt_file=$1
      ;;
    *)
      echo "mar-zcode: unsupported argument: $1" >&2
      exit 64
      ;;
  esac
  shift
done

if [ -z "$prompt_file" ] || [ ! -f "$prompt_file" ]; then
  echo "mar-zcode: prompt file missing" >&2
  exit 66
fi

# Project-local MARTL adapter: KAH must inject HOME from
# toolchain.operator.real_user_home before invoking provider CLIs.
if [ -z "${HOME:-}" ]; then
  echo "mar-zcode: HOME must be supplied by KAH provider env normalization" >&2
  exit 78
fi
prompt_text=$(cat "$prompt_file")
if [ -f /Applications/ZCode.app/Contents/Resources/glm/zcode.cjs ]; then
  cd /Applications/ZCode.app/Contents/Resources/glm
  exec node ./zcode.cjs --mode plan --prompt "$prompt_text"
fi
exec zcode --mode plan --prompt "$prompt_text"
