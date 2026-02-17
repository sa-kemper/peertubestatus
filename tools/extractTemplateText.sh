#!/usr/bin/env bash
# Extract every translatable string from the gohtml template files provided as arguments.
grep -oP '(?<=translate ")[^"]*' "$1" | while read -r line; do
    printf "msgid \"$line\"\r\nmsgstr \"$line\"\r\n" >> en.pot
done
