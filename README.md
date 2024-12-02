
.env
```dotenv
    BIND_ADDR=#port of your service
    DB_USERNAME=
    DB_HOST=postgres-db #db service name
    DB_PORT=
    DB_NAME=
    DB_PASSWORD=
    DB_SSLMODE=
    API_MUSIC_ADDRESS= #address of your api
    LOGGER_TYPE=#local(for text handler)/dev(for json handler)
```

```bash
make up
make migrationUp
```