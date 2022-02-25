# TFA coinlist

> Сервис для получения кодов аутенфикации для coinlist.co

## Содержание

1. [Подготовка](#подготовка)
2. [Deploy](#deploy)

## Подготовка

Установить `nodejs` и `redoc-cli`.

```shell
sudo apt install nodejs
sudo npm i -g redoc-cli
```

Скачать и собрать [sup](https://github.com/NovikovRoman/sup). После сборки переместить:

```shell
sudo mv bin/sup /usr/bin
```

## Deploy

В `~/.ssh/config` прописать доступы к серверам:

```shell
Host servername
    Hostname ipaddr
    User user
    Port 22
    IdentityFile ~/.ssh/key
```

Запустить:

```shell
sup production deploy
```

На сервере в директории проекта создать файл `.env` (образец `.env-sample`).
