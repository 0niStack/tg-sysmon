# tg-sysmon

Telegram bot that polls CPU, RAM, and disk usage for configured block devices
(e.g. `/dev/sda`, `/dev/sdb`) and alerts a chat when any exceed threshold.
`/status` returns current values on demand.

## Config (.env)

```
TELEGRAM_BOT_TOKEN=123456789:AA...
TELEGRAM_CHAT_ID=987654321
CPU_THRESHOLD=80
RAM_THRESHOLD=80
DISK_THRESHOLD=85
DISKS=sda,sdb
CHECK_INTERVAL_SECONDS=60
DISK_ROOT_PREFIX=
```

`DISKS` matches against the device path returned by the kernel (e.g. `sda`
matches `/dev/sda1`, `/dev/sda2`, ...). `TELEGRAM_CHAT_ID` is the numeric
chat/user id to send alerts to — get it by messaging your bot then hitting
`https://api.telegram.org/bot<token>/getUpdates`.

`DISK_ROOT_PREFIX` is only needed in Docker: it's prepended to each
mountpoint before checking usage, so the container reads the host's actual
disk instead of its own overlay filesystem. Leave empty when running
directly on the host.

## Run locally

```
go run .
```

## Docker

Build:

```
docker build -t tg-sysmon .
```

Run — needs the host's `/proc`, `/sys` and root filesystem visible so it
reads real host stats instead of the container's own:

```
docker run -d --name tg-sysmon \
  --restart unless-stopped \
  --env-file .env \
  --pid host \
  -v /proc:/host/proc:ro \
  -v /sys:/host/sys:ro \
  -v /:/host/root:ro \
  -e HOST_PROC=/host/proc \
  -e HOST_SYS=/host/sys \
  -e DISK_ROOT_PREFIX=/host/root \
  tg-sysmon
```

## docker-compose

```yaml
services:
  tg-sysmon:
    build: .
    restart: unless-stopped
    env_file: .env
    pid: host
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/host/root:ro
    environment:
      - HOST_PROC=/host/proc
      - HOST_SYS=/host/sys
      - DISK_ROOT_PREFIX=/host/root
```
