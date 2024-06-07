## birthday_notify

English version of this README can be found below.

birthday_notify - сервис для отслеживания дней рождений пользователей.

Подробную документацию можно найти на /api/docs

#### Для запуска необходимо следующее:
- установленный docker
- ```.env``` файл следующего формата:
```
POSTGRES_USER=birthday_user
POSTGRES_PASSWORD=1234
POSTGRES_DB=birthday
DB_HOST=localhost
DB_PORT=5432
JWT_SECRET_KEY=secret-key
```

####  Сервис запускается с помощью ```docker compose up```

#### В сервисе доступны следующие эндпоинты:
- GET /api/users *Получить список всех пользователей (доступна пагинация через page и page_zize параметры запроса)*
- POST /api/users *Создать пользователя (доступно по токену)*
- PUT, PATCH /api/users/{id:[0-9]+} *Частично или полностью обновить пользователя (доступно по токену)*
- GET /api/users/{id:[0-9]+} *Получить пользователя по его id*
- POST /api/users/{id:[0-9]+}/subscribe *Подписаться на день рождения пользователя (доступно по токену)*
- POST /api/users/{id:[0-9]+}/unsubscribe *Отписаться от дня рождения пользователя (доступно по токену)*
- GET /api/birthdays *Получить список пользователей, на которых подписан текущий пользователь, и у кого из них сегодня день рождения (доступно по токену)*
- GET /api/subscriptions *Получить список пользователей, на которых подписан текущий пользователь (доступно по токену)*
- POST /api/auth/token *Получить токен для пользователя*
- GET /api/liveness *liveness-check сервиса*

Доступны следующие поля к теле запроса:
```
    "firstName": string,
    "lastName": string,
    "email": string в формате email,
    "birthday": string в формате "2002-05-16T00:00:00Z",
    "password": string
```

birthday_notify - a service for tracking users' birthdays.

Detailed documentation can be found on /api/docs

#### Requirements for running:
- Installed Docker
- A `.env` file with the following format:
```
POSTGRES_USER=birthday_user
POSTGRES_PASSWORD=1234
POSTGRES_DB=birthday
DB_HOST=localhost
DB_PORT=5432
JWT_SECRET_KEY=secret-key
```

#### To start the service, use: ```docker compose up```

Available endpoints in the service:
- GET /api/users *Retrieve a list of all users (pagination is possible with page and page_size query parameters)*
- POST /api/users *Create a user (token required)*
- PUT, PATCH /api/users/{id:[0-9]+} *Partially or fully update a user (token required)*
- GET /api/users/{id:[0-9]+} *Retrieve a user by their ID*
- POST /api/users/{id:[0-9]+}/subscribe *Subscribe to a user's birthday (token required)*
- POST /api/users/{id:[0-9]+}/unsubscribe *Unsubscribe from a user's birthday (token required)*
- GET /api/birthdays *Get a list of users the current user is subscribed to and whose birthday is today (token required)*
- GET /api/subscriptions *Get a list of users the current user is subscribed to (token required)*
- POST /api/auth/token *Get a token for a user*
- GET /api/liveness *Service liveness check*

The following fields in the request body are available:
```
    "firstName": string,
    "lastName": string,
    "email": string in email format,
    "birthday": string in "2002-05-16T00:00:00Z" format,
    "password": string
```