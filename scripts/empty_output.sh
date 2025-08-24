#!/bin/bash
# scripts/empty_output.sh

set -euo pipefail

OUTPUT_DIR="output"

echo "Deleting all .drawio, .drawio.xml, .figma.json, and .txt files in the '$OUTPUT_DIR' directory..."

if [ ! -d "$OUTPUT_DIR" ]; then
  echo "Directory '$OUTPUT_DIR' does not exist. Nothing to clean."
  exit 0
fi

shopt -s nullglob
FILES=("$OUTPUT_DIR"/*.drawio "$OUTPUT_DIR"/*.drawio.xml "$OUTPUT_DIR"/*.figma.json "$OUTPUT_DIR"/*.txt)
shopt -u nullglob

if [ ${#FILES[@]} -eq 0 ]; then
  echo "No matching files to delete in '$OUTPUT_DIR'."
else
  rm -f "${FILES[@]}"
  echo "Deleted ${#FILES[@]} file(s) from '$OUTPUT_DIR'."
fi