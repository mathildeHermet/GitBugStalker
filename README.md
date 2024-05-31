# GitBugStalker

Build bin to track last commited bugs on Github Opensource Software repositories and forward it via webhook URL (ex: Discord channel).

## Prerequisite

### Discord

This bot will require a Discord webhook url to send message. To do this you will need a discord server where you have admin right.

1. Go to Settings of your discord server
2. Go to Integrations
3. Go to webhook and create a webhook
4. Click on "Copy URL webhook" on discord UI

This webhook URL should be used by the bot to send message

## Build

### Go

Basic right now. Just build it.

```bash
git clone git@github.com:mathildeHermet/GitBugStalker.git
cd GitBugStalker
go build -o bug-alert main.go
```

### Using Docker

```bash
docker build --target=app -t gitbugstalker .
```

To build a multiarch Docker image:

```bash
docker buildx build  --target=app --output=bin --platform=arm64,amd64 . -t gitbugstalker
```

To push it to your dockerhub account or any registry

```bash
docker buildx build  --target=app --output=bin --platform=arm64,amd64 . -t gitbugstalker -t myaccount/gitbugstalker --push
```

## Usage

### Docker

To run the stalker with in-memory cache (deleted at container restart)

```bash
# Run 
docker run --name gitbugstalker -d gitbugstalker --github-repo https://github.com/orga/repo/issues --refresh-interval 10m --discord-hook-url "https://my-chat/webhook-to-channel"
```

To run it with persistent cache

```bash
# Run 
docker run -v gitbugstalker-cache:/var/lib/git-bug-ring/ --name gitbugstalker -d gitbugstalker --github-repo https://github.com/orga/repo/issues --cache-file /var/lib/git-bug-ring/repo.txt --refresh-interval 10m --discord-hook-url "https://my-chat/webhook-to-channel"
```

### Systemd Daemon creation example

```bash
[Unit]
Description=GitHub to Discord Notification Service
After=network.target

[Service]
ExecStart=/usr/local/bin/bug-alert --github-repo https://github.com/orga/repo/issues --refresh-interval 10m --cache-file /var/lib/git-bug-ring/repo.txt --discord-hook-url "https://my-chat/webhook-to-channel"
WorkingDirectory=/usr/local/bin # Adapt location depending on binaries installation folder
User=root
Group=root

[Install]
WantedBy=multi-user.target
```
