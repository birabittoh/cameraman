# Cameraman
A simple calendar to remind yourself about yearly occurrences. It can send notifications via Telegram API.

## Instructions

First of all, you should set your Telegram parameters if you want notifications (you probably do).
```
cp .env.example .env
nano .env
```

### Docker
If you want to use the latest image, just do:
```
docker-compose up -d
```

Otherwise, you can build it yourself:
```
docker-compose up -d --build
```

### Test and debug locally
```
go test -v ./...
go run .
```
