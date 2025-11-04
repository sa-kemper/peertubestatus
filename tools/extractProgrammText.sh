#!/usr/bin/env bash
# Extract every translatable string from the go files provided as arguments.
grep -oP '(?<=Translate\(")[^"]*' "$1" | while read -r line; do
    echo "msgid \"$line\"\r\nmsgstr \"$line\"\r\n" >> en.pot
done
