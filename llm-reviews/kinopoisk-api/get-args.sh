#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$API_KEY" ]; then
    echo "API_KEY required (https://kinopoiskapiunofficial.tech/profile)"
    exit 1
fi

# первый аргумент - имя txt файл со списком строк вида https://www.kinopoisk.ru/film/258687/
filename="$1"

if [ -z "$filename" ]; then
    echo "filename required"
    exit 1
fi

# Каталог для результатов
outputdir="$2"

if [ -z "$outputdir" ]; then
    echo "outputdir required"
    exit 1
fi

if [ ! -d "$outputdir" ]; then
  echo "$outputdir does not exist."
  exit 1
fi

echo "outputdir = $outputdir"

filmIds=$(
    grep -oE '[0-9]+' "$filename" |
    paste -sd,
)
