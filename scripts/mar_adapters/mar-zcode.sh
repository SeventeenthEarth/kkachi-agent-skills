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

# Project-local MARTL adapter: pin the 17번째 지구 Hermes user HOME and local CLI path
# so provider auth state is read consistently under Hermes. Generalize only via a
# separate toolchain-discovery task.
export HOME=/Users/draccoon
prompt_text=$(cat "$prompt_file")
if [ -f /Applications/ZCode.app/Contents/Resources/glm/zcode.cjs ]; then
  cd /Applications/ZCode.app/Contents/Resources/glm
  exec node ./zcode.cjs --mode plan --prompt "$prompt_text"
fi
exec zcode --mode plan --prompt "$prompt_text"
