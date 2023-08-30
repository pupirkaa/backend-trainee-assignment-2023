# Тестовое задание для стажёра Backend

## Инструкция по запуску
Docker Compose поднимет базу данных PostgreSQL и заэкспозит порт 5432 в хостовую сеть
```bash
$ docker compose up
```

## Примеры запросов

| Название | curl |
| --- | --- |
| Создать сегмент | `curl --request POST --url http://localhost:8000/api/create_segment --header 'Content-Type: application/json' --data '{"segment":"TEST_SEGMENT"}'` |
| Удалить сегмент | `curl --request POST --url http://localhost:8000/api/delete_segment --header 'Content-Type: application/json' --data '{"segment":"TEST_SEGMENT"}'` |
| Изменить сегменты пользователя | `curl --request POST --url http://localhost:8000/api/change_user_segments --header 'Content-Type: application/json' --data '{"user_id":1,"segments_to_add":["TEST_SEGMENT"], "segments_to_delete"["TEST_SEGMENT"]}'` |
| Получить сегменты пользователя | `curl --request GET --url http://localhost:8000/api/get_user_segments --header 'Content-Type: application/json' --data '{"user_id":1}'` |

