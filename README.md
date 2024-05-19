[![Publish my app to DockerHub](https://github.com/Sleeps17/linker/actions/workflows/publish.yml/badge.svg?branch=main)](https://github.com/Sleeps17/linker/actions/workflows/publish.yml)

# linker
## Пояснение
Обычно у программиста есть огромное количество задач, по которым у него открыто огромное количество вкладок в браузере. Это очень сильно захламляет рабочее пространство. Для очистки рабоче
го пространства и структуризации источников, которые вам нужно изучить был написан сервис linker. \newline
linker - сервис, который позволяет хранить все свои ссылки в соответсвующих топиках, получать их оттуда, удалять и т.д.
## Клиент
Для этого сервиса я написал простой консольных клиент для Linux: ``https://github.com/Sleeps17/linker-client``
Если вы хотите написать, что-то свое то можете использовать эти protobuf файлы: ``https://github.com/Sleeps17/linker-protos``
## Запуск
Чтобы развернуть этот сервис на своей машине вам нужно иметь установленные docker и docker-compose, а также выполнить следующие шаги:
1) Установить консольную утилиту task - ``sudo snap install task --classic``
2) Создать файл ``docker-compose.yml`` со следующим содержанием
```yaml
version: '3.9'

services:
    postgres:
        image: postgres
        container_name: linker-postgres
        restart: always
        ports:
            - "5433:5432"
        environment:
            - POSTGRES_USER=your_username
            - POSTGRES_DB=linker-db
            - POSTGRES_PASSWORD=your_password
        networks:
            - proxynet
        volumes:
            - "postgres_data:/var/lib/postgresql/data"

    linker:
        image: sleeps17/linker:v2.0.1
        container_name: linker-service
        restart: always
        ports:
            - "4404:4404"
        networks:
            - proxynet
        depends_on:
            - postgres

networks:
    proxynet:
        external:
            name: proxynet

volumes:
    postgres_data:
```
3) Создать файл ``create-network.sh`` со следующим содержанием
```yaml
#!/bin/bash

NETWORK_NAME="proxynet"

if ! docker network inspect $NETWORK_NAME &> /dev/null; then
    echo "Creating network $NETWORK_NAME"
    docker network create $NETWORK_NAME
else
    echo "Network $NETWORK_NAME already exists"
fi
```
4) Сделать его исполняемым ``sudo chmod +x create-network.sh``
5) Создать файл ``Taskfile.yml`` со следующим содержанием
```yaml
version: '3'

tasks:
  run:
    cmds:
      - ./create-network.sh
      - docker-compose -f compose-url-shortener.yml up -d
      - docker-compose -f compose-linker.yml up -d
  stop:
    cmds:
      - docker-compose -f compose-url-shortener.yml down
      - docker-compose -f compose-linker.yml down
```
6) Запустить сервис командой ``task run``
7) Если вы хотите, чтобы этот сервис работал всегда, когда работает ваш компбютер, то нужно сделать этот сервис системной службой:
- ``touch /etc/systemd/system/linker.service``
- Записать в него следующее:
```
[Unit]
Description=My Linker Service
After=docker.service
Requires=docker.service

[Service]
Type=simple
RemainAfterExit=yes
WorkingDirectory=/home/pasha/Projects/linker
ExecStart=/snap/bin/task run
ExecStop=/snap/bin/task stop

[Install]
WantedBy=multi-user.target
Alias=linker.service
```
