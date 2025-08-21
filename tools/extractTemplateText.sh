#!/usr/bin/env bash
grep -oP '(?<=translate ")[^"]*' $1 | while read -r line; do
    echo "msgid \"$line\"" >> en.pot
    echo "msgstr \"$line\"" >> en.pot
    echo "" >> en.pot
done
