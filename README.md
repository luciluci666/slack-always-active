# Slack Always Active

A Go application that keeps your Slack status active by maintaining a WebSocket connection and sending periodic pings. The application respects working hours and only maintains the connection during specified times.

## Features

- Maintains WebSocket connection to Slack
- Sends periodic pings to keep status active
- Configurable working hours and days
- GMT offset support for correct timezone handling
- Automatic reconnection on connection loss
- Docker support for easy deployment

## Prerequisites

- Go 1.21 or later
- Docker (optional, for containerized deployment)
- Slack account with appropriate permissions

## Configuration

Create a `.env` file in the project root with the following variables:

```env
# Required
SLACK_TOKEN=your_slack_token
SLACK_COOKIE=your_slack_cookie

# Optional - Working Hours Configuration
WORK_DAYS=Monday,Tuesday,Wednesday,Thursday,Friday
WORK_START=09:00
WORK_END=18:00
GMT_OFFSET=+2  # Your timezone offset (e.g., +2 for UTC+2)
```

### Environment Variables

- `SLACK_TOKEN`: Your Slack API token (required)
- `SLACK_COOKIE`: Your Slack session cookie (required)
- `WORK_DAYS`: Comma-separated list of working days (default: Monday-Friday)
- `WORK_START`: Start time in 24-hour format (default: 09:00)
- `WORK_END`: End time in 24-hour format (default: 18:00)
- `GMT_OFFSET`: Your timezone offset (e.g., +2 for UTC+2)

### GMT Offset Examples

- `+0`: UTC/GMT
- `+1`: Central European Time (CET)
- `+2`: Eastern European Time (EET)
- `-5`: Eastern Time (ET)
- `-8`: Pacific Time (PT)

## Building and Running

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/lucy/slack-always-active.git
   cd slack-always-active
   ```

2. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

3. Edit `.env` with your credentials and preferences

4. Build and run:
   ```bash
   go build
   ./slack-always-active
   ```

### Docker Deployment

1. Build the Docker image:
   ```bash
   docker build -t slack-always-active .
   ```

2. Make sure your `.env` file is properly configured with all required variables:
   ```env
   SLACK_TOKEN=your_slack_token
   SLACK_COOKIE=your_slack_cookie
   WORK_DAYS=Monday,Tuesday,Wednesday,Thursday,Friday
   WORK_START=09:00
   WORK_END=18:00
   GMT_OFFSET=+2
   ```

3. Run the container:
   ```bash
   docker run -d --name slack-always-active \
     --env-file .env \
     -v $(pwd)/logs:/app/logs \
     slack-always-active
   ```

   This command will:
   - Use all environment variables from your `.env` file
   - Mount the logs directory to persist logs on your host machine
   - Run the container in detached mode (-d)

   Make sure your `.env` file is in the same directory where you run the docker command.
   
## Logging

The application logs all activities to both stdout and a log file. When running in Docker, logs are stored in `/app/logs/slack-always-active.log` inside the container. The logs directory is exposed as a volume that can be mounted to the host.

### Viewing Logs

From the host:
```bash
tail -f logs/slack-always-active.log
```

From the container:
```bash
docker exec slack-always-active tail -f /app/logs/slack-always-active.log
```

## License

MIT License - see LICENSE file for details 