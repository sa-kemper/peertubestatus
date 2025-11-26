#!/usr/bin/env bash
grep -oP '(?<=translate ")[^"]*' "$1" | while read -r line; do
    printf "msgid \"$line\"\r\nmsgstr \"$line\"\r\n" >> en.pot
done
