#!/bin/sh
set -eu

prompt_file=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --prompt-file)
      shift
      if [ "$#" -eq 0 ]; then
        echo "mar-agy: --prompt-file requires a path" >&2
        exit 64
      fi
      prompt_file=$1
      ;;
    *)
      echo "mar-agy: unsupported argument: $1" >&2
      exit 64
      ;;
  esac
  shift
done

if [ -z "$prompt_file" ] || [ ! -f "$prompt_file" ]; then
  echo "mar-agy: prompt file missing" >&2
  exit 66
fi

# Project-local MARTL adapter: KAH must inject HOME from
# toolchain.operator.real_user_home before invoking provider CLIs.
if [ -z "${HOME:-}" ]; then
  echo "mar-agy: HOME must be supplied by KAH provider env normalization" >&2
  exit 78
fi
exec /Users/draccoon/.local/bin/agy \
  --print \
  --model "Gemini 3.5 Flash (High)" \
  --print-timeout 120s \
  --sandbox < "$prompt_file"
