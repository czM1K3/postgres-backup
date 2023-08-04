# Postgres Backup
This project create backups of PostgreSQL database and optionally uploads them to remote host with rclone.

## Usage
### Backup locally
```yml
version: "2.2"
services:
  backup:
    image: "ghcr.io/czm1k3/postgres-backup"
    environment:
      - "POSTGRESQL_URI=postgres://user:password@localhost:5432/database"
      - "CRON_INTERVAL=0 4 * * 6" # Every saturday at 4:00
      - "TZ=Europe/Prague" # Timezone
    volumes:
      - "./database-backup:/backup"
    restart: always
```

### Backup locally and remotely
```yml
version: "2.2"
services:
  backup:
    image: "ghcr.io/czm1k3/postgres-backup"
    environment:
      - "POSTGRESQL_URI=postgres://user:password@localhost:5432/database"
      - "CRON_INTERVAL=0 4 * * 6" # Every saturday at 4:00
      - "TZ=Europe/Prague" # Timezone
      - "EXTERNAL_BACKUP_PATH=/path" # On what path to put on remote
    volumes:
      - "./database-backup:/backup"
      - "./rclone:/home/backupper/.config/rclone" # It is required to create rclone.conf file inside this folder before starting container. Program expects profile called "remote" to where put files. Permissions may change so it is not recommended to use system config file. It has to be persisted, because of token refreshing.
    restart: always
```

## Note
Version of rclone depends on what version uses Debian 12. On time of making this repository it's 1.60.1.