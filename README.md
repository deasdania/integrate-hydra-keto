## Build base components

Clone and build individual components:

### Hydra
v2.2.0 (Latest)
(on Feb 12, 2024)

```sh
git clone https://github.com/ory/hydra.git .
cd hydra
```

## Database Configuration
In the provided `quickstart-postgres.yml` file, make sure to modify the following values to match your local PostgreSQL setup:

dbuser: The PostgreSQL username (e.g., postgresuser)
password: The PostgreSQL password (e.g., secret)
dbname: The name of the database (e.g., hydra)

```
services:
  hydra-migrate:
    environment:
      - DSN=postgres://postgres:secret@host.docker.internal:5432/hydra?sslmode=disable&max_conns=20&max_idle_conns=4
  hydra:
    environment:
      - DSN=postgres://postgres:secret@host.docker.internal:5432/hydra?sslmode=disable&max_conns=20&max_idle_conns=4
```

run it with 
```sh
docker-compose -f quickstart.yml \
    -f quickstart-postgres.yml \
    up --build
```

The base-url for hydra is `http://localhost:4445/`


### Keto
v0.13.0-alpha.0 (Pre-release)
(on Feb 28, 2024)
```sh
git clone https://github.com/ory/keto.git .
cd keto
```
## Database Configuration
In the provided `docker-compose-postgres.yml` file, make sure to modify the following values to match your local PostgreSQL setup:

dbuser: The PostgreSQL username (e.g., postgresuser)
password: The PostgreSQL password (e.g., secret)
dbname: The name of the database (e.g., keto)

```
version: "3.2"
services:
  keto-migrate:
    image: oryd/keto:v0.12.0-alpha.0
    volumes:
      - type: bind
        source: ./config
        target: /home/ory
    environment:
      - LOG_LEVEL=debug
      - DSN=postgres://postgresuser:secret@host.docker.internal:5432/keto?sslmode=disable
    command: ["migrate", "up", "-y"]
    restart: on-failure
  keto:
    image: oryd/keto:v0.11.1-alpha.0
    volumes:
      - type: bind
        source: ./config
        target: /home/ory
    ports:
      - "4466:4466"
      - "4467:4467"
    depends_on:
      - keto-migrate
    environment:
      - DSN=postgres://postgresuser:secret@host.docker.internal:5432/keto?sslmode=disable
    restart: on-failure
  
networks:
  keto-network:
    driver: bridge  # Using default bridge network

```

run it with 
```sh
docker compose -f docker-compose-postgres.yml up --build
```

The base-url for keto is `http://localhost:4466`