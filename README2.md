# headers - traefik_headers

Плагин "headers" для сервиса traefik.

## 1.1. Хранение конфигурации

конфигурация "headers" загружается из keeper для оперативного управления "headers", или из параметров middleware **headersData** при инициализации 
плагина, в случае недоступности keeper.
Данная конфигурация предназначена для подстановки в ответы сервера ключей и значений http header в случае если такого ключа еще не устанвлено.
конфигурация представляет из себя **hashmap** и состоит из:

- **Ключ header**
  - *Тип:* Строка
  - *Обязательность:* Да
  - *Примечание:* Содержит ключ заголовка http ответа. Значение будет приведено к каноническому виду . Например значение ключа ```Content-security-Policy``` будет приведено к 
    ```Content-Security-Policy```. Ключи с пустым списком значений или со списком значений не содержащих значащие символы будут проигнорированы.
    Если присутствуют полностью идентичные ключи приведенные к каноническому виду, то будет использоваться рандомный (такой ситуации надо избегать).

- **Значения header**
  - *Тип:* Массив строк
  - *Обязательность:* Да
  - *Примечание:* Содержит значения заголовка http ответа. Пустые значения и значения не содержашие значащих символов будут проигнорированы. При загрузке в плагин пробелы справа и слева будут убраны из значения.

-  примеры заголовков:
   - ``` 
     { 
       "Access-Control-Allow-Origin": [
         "*"
       ],
       "Access-Control-Allow-Methods": [
         "GET,POST,OPTIONS"
       ],
       "Access-Control-Allow-Credentials": [
         "true"
       ],
       "Access-Control-Expose-Headers": [
         "Date,Content-Length,Content-Range,X-Request-Id,Strict-Transport-Security,X-Content-Type-Options,X-Frame-Options,X-XSS-Protection"
       ],
       "Access-Control-Allow-Headers": [
         "Accept,DNT,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Range,Host,X-Request-Id,Access-Control-Allow-Origin,Access-Control-Allow-Methods,Access-Control-Expose-Headers"
       ],
       "Strict-Transport-Security": [
         "max-age=31536000; includeSubDomains; preload"
       ],
       "X-Content-Type-Options": [
         "nosniff"
       ],
       "X-Frame-Options": [
         "SAMEORIGIN",
         "ALLOW-FROM wildberries.ru",
         "ALLOW-FROM *.wildberries.ru",
         "ALLOW-FROM wildberries.am",
         "ALLOW-FROM *.wildberries.am",
         "ALLOW-FROM wildberries.kg",
         "ALLOW-FROM *.wildberries.kg",
         "ALLOW-FROM wildberries.by",
         "ALLOW-FROM *.wildberries.by",
         "ALLOW-FROM wildberries.kz",
         "ALLOW-FROM *.wildberries.kz",
         "ALLOW-FROM wildberries.ua",
         "ALLOW-FROM *.wildberries.ua",
         "ALLOW-FROM wildberries.eu",
         "ALLOW-FROM *.wildberries.eu",
         "ALLOW-FROM wildberries.ge",
         "ALLOW-FROM *.wildberries.ge"
       ],
       "Content-Security-Policy": [
         "connect-src *; frame-ancestors wildberries.ru *.wildberries.ru wildberries.am *.wildberries.am wildberries.kg *.wildberries.kg wildberries.by *.wildberries.by wildberries.kz *.wildberries.kz wildberries.ua *.wildberries.ua wildberries.eu *.wildberries.eu wildberries.ge *.wildberries.ge"
       ],
       "X-XSS-Protection": [
         "1; mode=block"
       ],
       "Cache-Control": [
         "no-cache, no-store, must-revalidate"
       ]

     }
     ```
    
    заголовки http ответа

Конфигурация хранится в локальном кэше headers

## 1.2. Обновление конфигурации

Конфигурация обновляется периодически, 1 раз в 30 сек из keeper

## 1.2. Параметры плагина

```
traefikMiddleware: 
    traefik-headers:
    spec:
      keeperHeadersKey: headers
      keeperURL: https://keeper-ext-feature-wg-8238.k8s.dev.paywb.lan
      keeperReqTimeout: 100s
      keeperUsername: username
      keeperPassword: Pas$w0rd
      keeperReloadInterval: 45s
      headersData: |
        { 
          "Content-Security-Policy": [
            "connect-src *; frame-ancestors wildberries.ru *.wildberries.ru wildberries.am *.wildberries.am wildberries.kg *.wildberries.kg wildberries.by *.wildberries.by wildberries.kz *.wildberries.kz wildberries.ua *.wildberries.ua wildberries.eu *.wildberries.eu wildberries.ge *.wildberries.ge"
          ]
        }
```
- *keeperHeadersKey* - ключ в keeper, под которым хранится json конфигурация
- *keeperURL* - url keeper, в котором хранится json кофиграция
- *keeperReqTimeout* - таймаут ожидания ответа при запросе к keeper. По умолчанию 300s
- *keeperUsername* - пользователь keeper
- *keeperPassword* - пароль keeper
- *keeperReloadInterval* - интервал опроса keeper для получения обновлений конфигурации. По умолчанию 30s
- *headersData* - json конфигурации плагина, который будет использоваться в случае недоступности keeper при инициализации плагина

## Логика работы "headers"

Плагин подставляет в заголовок http ответа те ключи, которые отсутствуют в изначальном ответе серевера.
