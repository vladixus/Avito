## Настройка запуска Проекта "TestAvito"
***
1) Склонируйте данный репозиторий на свой компьютер `git clone https://github.com/your-username/test-avito.git` 
2) В корне проекта создайте файл .env и укажите параметры подключения к базе данных PostgreSQL:
  `DB_USER=your_db_user
   DB_PASSWORD=your_db_password
   DB_HOST=db
   DB_PORT=5432
   DB_NAME=avito`
3) Запустите проект с помощью Docker Compose:
`docker compose build` и `docker compose up` из вашей директории  проекта.
4) После успешного запуска контейнеров, приложение будет доступно по адресу http://localhost:8080.
## API

+ POST /user: Создание пользователей.
> {
"name":["vladixus"]
}
+ POST /segment: Создание сегмента.
> {
"name":"AVITO_DISCOUNT_30"
}
+ DELETE /segment/{segment_name}: Удаление сегмента по названию.

+ POST /user/{user_id}/segments: Добавление и удаление сегментов у пользователя.
> {
"add":["AVITO_DISCOUNT_30"],
"remove":["AVITO_DISCOUNT_50"]
}
+ GET /user/{user_id}/active_segments: Получение активных сегментов пользователя.
