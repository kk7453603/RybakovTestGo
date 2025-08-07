#!/bin/bash
echo "Исправление Swagger файла..."

SWAGGER_FILE="docs/service.swagger.json"

if [ -f "$SWAGGER_FILE" ]; then
    # Создаем исправленный файл с правильной структурой
    jq '{
        swagger: .swagger,
        info: {
            title: "Crypto Currency API",
            version: "1.0.0",
            description: "API для управления криптовалютами"
        },
        host: "localhost:8080",
        schemes: ["http"],
        tags: .tags,
        consumes: .consumes,
        produces: .produces,
        paths: .paths,
        definitions: .definitions
    }' "$SWAGGER_FILE" > "${SWAGGER_FILE}.tmp" && mv "${SWAGGER_FILE}.tmp" "$SWAGGER_FILE"
    rm "${SWAGGER_FILE}.tmp"
    echo "✅ Swagger файл исправлен!"
else
    echo "❌ Swagger файл не найден: $SWAGGER_FILE"
    exit 1
fi
