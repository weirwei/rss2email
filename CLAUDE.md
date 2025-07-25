# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

使用中文交流

## Project Overview

This is an RSS to email service written in Go that periodically fetches RSS feeds and sends updates via email to subscribers. The service uses SQLite for storing subscription information and tracks the last processed item for each subscription to avoid duplicate emails.

## Architecture

- **Main Entry Point**: `main.go` initializes configuration and executes the root command
- **CLI Framework**: Uses Cobra for command-line interface
- **Scheduling**: Uses cron library for periodic execution of RSS checks
- **RSS Processing**: Uses gofeed library to parse RSS feeds
- **Database**: SQLite with GORM for storing user subscriptions
- **Email**: Uses gomail library for sending emails
- **Configuration**: YAML-based configuration files

## Key Components

1. **Commands** (`cmd/`): CLI commands including root command and registration
2. **Services** (`service/`): RSS feed processing logic for different sources
3. **Models** (`models/`): Database models and DAOs for user subscriptions
4. **Helpers** (`helpers/`): Database and email initialization
5. **Configuration** (`conf/`): Configuration loading and management
6. **Constants** (`constants/`): Enum-like constants for subscriptions and processing types

## Development Commands

### Building
```bash
# Build the application
go build -o rss2email main.go

# Build Docker image
make docker-image-build

# Run Docker container
make docker-run
```

### Running
```bash
# Run the application
./rss2email

# Run with Docker
docker run -v /etc/localtime:/etc/localtime:ro -v ./db:/usr/local/bin/db -d rss2email:${VERSION}
```

### Testing
```bash
# Run tests (uses test package initialization)
go test ./...
```

## Database Schema

The application uses SQLite with a single table:
- `user_subscriptions`: Tracks user email subscriptions and processing state

## Configuration Files

Configuration files are in `conf/yaml/`:
- `feedsource.yaml`: RSS feed URLs for different sources
- `email.yaml`: Email server configuration

## Adding New RSS Sources

1. Add the source to `constants/subscription.go`
2. Add the feed URL to `conf/yaml/feedsource.yaml`
3. Create a new service file in `service/` following the pattern of existing services
4. Add the service to the scheduling in `cmd/root.go`