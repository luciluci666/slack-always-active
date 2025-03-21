# Slack Always Active

A Go application that keeps your Slack status active by maintaining a WebSocket connection and periodically checking your authentication status.

## Features

- Validates Slack token and cookies
- Maintains WebSocket connection to Slack
- Automatic reconnection on connection loss
- Environment variable configuration
- Docker support
- File logging with volume support

## Prerequisites

- Go 1.21 or later
- Slack account with valid token and cookies
- Docker (optional, for containerized deployment)

## Setup

### Local Development

1. Clone the repository:
```bash
git clone https://github.com/lucy/slack-always-active.git
cd slack-always-active
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the project root with your Slack credentials:
```env
SLACK_TOKEN=your_slack_token
SLACK_COOKIE=your_slack_cookie
```

4. Run the application:
```bash
go run main.go
```

### Docker Deployment

1. Build the Docker image:
```bash
docker build -t slack-always-active .
```

2. Create a `.env` file with your Slack credentials:
```env
SLACK_TOKEN=your_slack_token
SLACK_COOKIE=your_slack_cookie
```

3. Run the container with volume for logs:
```bash
docker run -d \
  --name slack-always-active \
  --env-file .env \
  -v $(pwd)/logs:/app/logs \
  slack-always-active
```

Or run with environment variables directly:
```bash
docker run -d \
  --name slack-always-active \
  -e SLACK_TOKEN=your_slack_token \
  -e SLACK_COOKIE=your_slack_cookie \
  -v $(pwd)/logs:/app/logs \
  slack-always-active
```

## Logging

The application logs all activities to both stdout and a log file. When running in Docker:

- Logs are stored in `/app/logs/slack-always-active.log` inside the container
- The logs directory is exposed as a volume and can be mounted to the host
- Logs include timestamps and are formatted for easy reading
- Error messages are prefixed with "ERROR:"

To view logs:
```bash
# View logs from host
tail -f logs/slack-always-active.log

# View logs from container
docker logs slack-always-active
```

## Usage

Run the application:
```bash
go run main.go
```

The application will:
1. Check your Slack authentication status
2. Connect to Slack's WebSocket server
3. Maintain the connection and automatically reconnect if disconnected

## Security Note

Never commit your `.env` file or share your Slack token and cookies. These credentials should be kept private and secure.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 