# GitBugStalker

Build bin to track last commited bugs on Github Opensource Software repositories and forward it via webhook URL (ex: Discord channel).

## Build

Basic right now. Just build it.

```bash
git clone git@github.com:mathildeHermet/GitBugStalker.git
cd GitBugStalker
 go build -o bug-alert main.go
```

## Daemon creation example

```bash
[Unit]
Description=GitHub to Discord Notification Service
After=network.target

[Service]
ExecStart=/usr/local/bin/bug-alert --github-repo https://github.com/orga/repo --refresh-interval 10m --cache-file /var/lib/git-bug-ring/repo.txt --discord-hook-url "https://my-chat/webhook-to-channel"
WorkingDirectory=/usr/local/bin # Adapt location depending on binaries installation folder
User=root
Group=root

[Install]
WantedBy=multi-user.target
```