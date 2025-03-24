# GitLab to Apple Reminders Integration

This Go application integrates GitLab issues with Apple Reminders by creating reminders for issues assigned to a specific GitLab user.

## Features

- Fetches open GitLab issues assigned to a specified user
- Creates corresponding reminders in a specified Apple Reminders list
- Avoids creating duplicate reminders
- Runs periodically to keep reminders in sync with GitLab issues
- Includes issue details and URL in the reminder notes

## Requirements

- macOS (for Apple Reminders integration)
- Go 1.16 or later
- GitLab personal access token with API access

## Installation

1. Clone this repository or download the source code
2. Install the required dependencies:

```bash
go get github.com/everdev/mack
```

## Configuration

Create a `config.json` file with the following structure:

```json
{
  "gitlab_token": "your_gitlab_personal_access_token",
  "gitlab_url": "https://gitlab.com",
  "gitlab_username": "your_gitlab_username",
  "reminder_list": "GitLab Issues",
  "poll_interval_minutes": 60
}
```

- `gitlab_token`: Your GitLab personal access token (with `read_api` scope)
- `gitlab_url`: Your GitLab instance URL (default: https://gitlab.com)
- `gitlab_username`: The GitLab username whose assigned issues you want to track
- `reminder_list`: The name of the Apple Reminders list where to create reminders
- `poll_interval_minutes`: How often (in minutes) to sync issues

## Usage

1. Build the application:

```bash
go build -o gitlab-reminders
```

2. Run the application:

```bash
./gitlab-reminders -config=/path/to/config.json
```

If no config path is specified, the application will look for `config.json` in the current directory.

## How It Works

The application:

1. Fetches all open issues assigned to the specified GitLab user
2. For each issue, checks if a corresponding reminder already exists
3. If no matching reminder exists, creates a new reminder in the specified list
4. Repeats this process at the configured interval

## Limitations

- This application only works on macOS, as it uses AppleScript to interact with Apple Reminders
- The application only syncs one way (GitLab → Reminders)
- Only open issues are synced
- Changes to existing issues (other than state) do not update existing reminders

## Troubleshooting

- Ensure your GitLab token has appropriate permissions
- Check that the specified Reminders list exists in your Apple Reminders app
- Look at the application logs for error messages

## License

This project is licensed under the MIT License - see the LICENSE file for details.