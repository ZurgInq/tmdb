#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

. ${SCRIPT_DIR}/get-args.sh

# Разделитель - запятая
IFS=','

for filmId in $filmIds; do
    outputFile="${outputdir}/reviews/${filmId}.json"
    # Если файл уже существует - пропускаем
    if [[ -f "$outputFile" ]]; then
        echo "Файл $outputFile уже существует, пропускаю."
        continue
    fi

    echo "Получение отзывов для фильма $filmId..."

    if ! curl -fsS \
        -X GET \
        "https://kinopoiskapiunofficial.tech/api/v2.2/films/${filmId}/reviews?page=1&order=USER_POSITIVE_RATING_DESC" \
        -H "accept: application/json" \
        -H "X-API-KEY: ${API_KEY}" \
        -o "$outputFile"; then

        echo "Ошибка: запрос не выполнен или файл '$outputFile' не сохранён" >&2
        exit 1
    fi

    echo "Сохранено: $outputFile"
    sleep 0.1
done

echo "Готово."
