#!/bin/bash
# scripts/empty_output.sh

echo "🧹 Deleting all .drawio.xml files in the output directory..."
rm -f output/*.drawio.xml
echo "✅ output directory cleaned."
