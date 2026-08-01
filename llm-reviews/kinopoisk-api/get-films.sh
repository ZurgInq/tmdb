#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. ${SCRIPT_DIR}/get-args.sh

# Разделитель - запятая
IFS=','

for filmdId in $filmIds; do
    echo "Получение информации о фильме $filmdId..."

    outputFile="${outputdir}/films/film-${filmdId}.json"
    # Если файл уже существует - пропускаем
    if [[ -f "$outputFile" ]]; then
        echo "Файл $outputFile уже существует, пропускаю."
        continue
    fi

    # curl -sS \
    #     -X GET \
    #     "https://kinopoiskapiunofficial.tech/api/v2.2/films/${filmdId}" \
    #     -H "accept: application/json" \
    #     -H "X-API-KEY: ${API_KEY}" \
    #     -o "$outputFile"

    # echo "Сохранено: $outputFile"
    # sleep 0.1
done

echo "Готово."
