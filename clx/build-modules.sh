#!/bin/bash

for go_file in $(find mods -maxdepth 2 -name "*.go"); do
  dir=$(dirname "$go_file")
  filename=$(basename "$go_file" .go)

  output_path="$dir/$filename.so"

  echo "Building plugin: $go_file -> $output_path"
  go build -buildmode=plugin -o "$output_path" "$go_file"

  if [ $? -eq 0 ]; then
    echo "Successfully built $output_path"
  else
    echo "Failed to build $go_file"
    exit 1
  fi
done
