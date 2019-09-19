[//]: #tech "GoLang, Gorm, Oauth2, GraphQL, Facebook Dataloader"

This back-end provides persistence for user profiles, their projects, info and
project ratings.

Back-end is accessible trough GraphQL, which is processed by
[graph-gophers](https://github.com/graph-gophers/dataloader) version of [Facebook's
dataloader](https://github.com/facebook/dataloader), to prevent duplicate
queries and infinite loops in relationships.

## Config format

| Option  | Value          |
| ------- | -------------- |
| address | 127.0.0.1:8080 |

###### Discord

| Option        | Value                                |
| ------------- | ------------------------------------ |
| client_id     | Discord app id, to use for oauth2    |
| client_secret | Secret associated with the client_id |

###### Jwt

| Option | Value                            |
| ------ | -------------------------------- |
| secret | Temporay secret, for development |
| state  | Temporary state, for development |

###### postgres

| Option   | Value          |
| -------- | -------------- |
| host     | 127.0.0.1      |
| user     |                |
| password |                |
| dbname   | grip           |
| sslmode  | disable/enable |

## Example

```toml
address = "127.0.0.1:8080"

[discord]
client_id = "111111111111111111"
client_secret = ""

[jwt]
secret = "RandomSecret"
state = "RandomState"

[postgres]
host = "127.0.0.1"
user = "volskaya"
password = "changeme"
dbname = "grip"
sslmode = "disable"
```
