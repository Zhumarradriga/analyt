# AnalytService
Сервис для аналитики, работает на порту 8100.
Поддержка БД: ClickHouse
# Документация

## Пример использования
```
curl -i -X POST http://ip:9000/api/v1/track \
-H "Content-Type: application/json" \
-d '{
    "key": "registration",
    "value": {
        "user_id": 42,
        "email": "test@example.com",
        "source": "google"
    }
}'
```

В ответ придет статус выполнения запроса, посмотреть результаты можно в базе по адресу ip:8123/play.
Пароль для авторизации знает тот, кто знает.



