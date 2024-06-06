## birthday_notify

English version of this README can be found below.

birthday_notify - сервис для отслеживания дней рождений пользователей.

#### Для запуска необходимо следующее:
- установленный docker
- ```.env``` файл следующего формата:
```POSTGRES_USER=birthday_user
POSTGRES_PASSWORD=1234
POSTGRES_DB=birthday
DB_HOST=localhost
DB_PORT=5432
JWT_SECRET_KEY=secret-key
ADMIN_USER_FIRST_NAME="admin"
ADMIN_USER_LAST_NAME="admin"
ADMIN_USER_EMAIL="admin@gmail.com"
ADMIN_USER_BIRTHDAY="2002-05-16T00:00:00Z"
ADMIN_USER_PASSWORD="qwerty123"
```

####  Сервис запускается с помощью ```docker compose up```

#### В сервисе доступны следующие эндпоинты:
- GET /api/users *Получить список всех пользователей*
- POST /api/users *Создать пользователя (доступно по токену)*
- PUT, PATCH /api/users/{id:[0-9]+} *Частично или полностью обновить пользователя (доступно по токену)*
- GET /api/users/{id:[0-9]+} *Получить пользователя по его id*
- POST /api/users/{id:[0-9]+}/subscribe *Подписаться на день рождения пользователя (доступно по токену)*
- POST /api/users/{id:[0-9]+}/unsubscribe *Отписаться от дня рождения пользователя (доступно по токену)*
- GET /api/birthdays *Получить список пользователей, на которых подписан текущий пользователь, и у кого из них сегодня день рождения (доступно по токену)*
- GET /api/subscriptions *Получить список пользователей, на которых подписан текущий пользователь (доступно по токену)*
- POST /api/auth/token *Получить токен для пользователя*
- GET /api/liveness *liveness-check сервиса*

При запуске сервиса в первый раз в базе данных создается пользователь-админ, с помощью указанных в ```.env``` файле данных. Пользователь-админ может получить токен и создать дополнительных пользователей.

birthday_notify - a service for tracking users' birthdays.

#### Requirements for running:
- Installed Docker
- A `.env` file with the following format:
```plaintext
POSTGRES_USER=birthday_user
POSTGRES_PASSWORD=1234
POSTGRES_DB=birthday
DB_HOST=localhost
DB_PORT=5432
JWT_SECRET_KEY=secret-key
ADMIN_USER_FIRST_NAME="admin"
ADMIN_USER_LAST_NAME="admin"
ADMIN_USER_EMAIL="admin@gmail.com"
ADMIN_USER_BIRTHDAY="2002-05-16T00:00:00Z"
ADMIN_USER_PASSWORD="qwerty123"
```

#### To start the service, use: ```docker compose up```

Available endpoints in the service:
- GET /api/users *Retrieve a list of all users*
- POST /api/users *Create a user (token required)*
- PUT, PATCH /api/users/{id:[0-9]+} *Partially or fully update a user (token required)*
- GET /api/users/{id:[0-9]+} *Retrieve a user by their ID*
- POST /api/users/{id:[0-9]+}/subscribe *Subscribe to a user's birthday (token required)*
- POST /api/users/{id:[0-9]+}/unsubscribe *Unsubscribe from a user's birthday (token required)*
- GET /api/birthdays *Get a list of users the current user is subscribed to and whose birthday is today (token required)*
- GET /api/subscriptions *Get a list of users the current user is subscribed to (token required)*
- POST /api/auth/token *Get a token for a user*
- GET /api/liveness *Service liveness check*

When the service is launched for the first time, an admin user is created in the database using the data specified in the ```.env``` file. The admin user can obtain a token and create additional users.