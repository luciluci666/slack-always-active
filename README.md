# Slack Always Active

A Go application that keeps your Slack status active during working hours.

## Features

- Maintains Slack active status during configured working hours
- Automatically sleeps outside of working hours
- Configurable work days and hours
- Docker support with log persistence

## Configuration

Create a `.env` file based on the provided `.env.example`:

```bash
# Slack credentials
SLACK_TOKEN=xoxc-your-slack-token
SLACK_COOKIE=your-slack-cookie

# Schedule configuration
WORK_DAYS=monday,tuesday,wednesday,thursday,friday  # Comma-separated list of work days
WORK_START=09:00  # Work start time (24-hour format, UTC)
WORK_END=17:00    # Work end time (24-hour format, UTC)
```

### Schedule Configuration

The application supports flexible scheduling:

- **Work Days**: Specify work days using their names (monday through sunday), comma-separated
- **Work Hours**: Set start and end times in 24-hour format (HH:MM) in UTC

Default values if not specified:
- Work days: Monday through Friday
- Work hours: 09:00 to 17:00 UTC

## Running Locally

1. Clone the repository
2. Copy `.env.example` to `.env` and configure your settings
3. Run the application:
   ```bash
   go run main.go
   ```

## Docker Support

### Building the Image

```bash
docker build -t slack-always-active .
```

### Running with Docker

```bash
docker run -d \
  --name slack-always-active \
  -v $(pwd)/logs:/app/logs \
  -e SLACK_TOKEN=your_slack_token \
  -e SLACK_COOKIE=your_slack_cookie \
  -e WORK_DAYS=monday,tuesday,wednesday,thursday,friday \
  -e WORK_START=09:00 \
  -e WORK_END=17:00 \
  slack-always-active
```

Or using a `.env` file:

```bash
docker run -d \
  --name slack-always-active \
  -v $(pwd)/logs:/app/logs \
  --env-file .env \
  slack-always-active
```

## Logging

The application logs its activities to both stdout and a log file:

- Log file location: `logs/slack-always-active.log`
- When running with Docker, logs are stored in `/app/logs/slack-always-active.log` inside the container
- The logs directory is exposed as a volume that can be mounted to persist logs

To view logs:

- From the host: `tail -f logs/slack-always-active.log`
- From the container: `docker logs -f slack-always-active`

## License

MIT 