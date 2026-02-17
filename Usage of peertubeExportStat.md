# PeerTube Export Stat CLI Usage

## Command-Line Flags

### Notes

- **Every flag can be used with double dashes (e.g., `--api-host`)**
- A `.env` file in the working directory is supported without dashes
- **Bug Alert: When using a `.env` file or environment variables, CLI parameters are IGNORED**
- `.env` file supports:
    - Comments
    - Empty lines
    - Empty values are treated as unset

### Available Flags

| Flag                   | Description                              | Default Value                               |
|------------------------|------------------------------------------|---------------------------------------------|
| `-api-host`            | PeerTube API host                        | `"peertube.example.com"`                    |
| `-cache-valid-seconds` | Video database cache validity in seconds | `90000` (slightly more than a day)          |
| `-data-folder`         | Folder containing video stats            | `"./Data"`                                  |
| `-end-date`            | End date                                 | *Not set*                                   |
| `-log-level`           | Logging level                            | `2` (warning)                               |
| `-miss-tolerance`      | Tolerance for missing statistic days     | *Not set*                                   |
| `-output`              | Output folder                            | `"./Reports"`                               |
| `-output-language`     | Output language (requires locale file)   | `"de"`                                      |
| `-sample-frequency`    | Sampling frequency                       | `"Daily"` (options: Daily, Monthly, Yearly) |
| `-smtpFromAddress`     | SMTP from address                        | `"peertubestats@localhost"`                 |
| `-smtpHost`            | SMTP server host                         | `"localhost"`                               |
| `-smtpPassword`        | SMTP password                            | *Not set*                                   |
| `-smtpPort`            | SMTP server port                         | `25`                                        |
| `-smtpToAddress`       | Administrator recipient list             | `"admin <root@localhost>"`                  |
| `-smtpUsername`        | SMTP username                            | *Not set*                                   |
| `-start-date`          | Start date                               | *Not set*                                   |
| `-stat-io-max-threads` | Maximum number of threads                | `10`                                        |

## Log Levels

| Level   | Numeric Value | Description                                 |
|---------|---------------|---------------------------------------------|
| Fatal   | `0`           | Critical errors that stop execution         |
| Error   | `1`           | Significant issues                          |
| Warning | `2`           | Potential problems or unexpected conditions |
| Info    | `3`           | General information about program operation |
| Debug   | `4`           | Detailed diagnostic information             |

## Example Usage

```bash
# Using single dash
./peertubeExportStat -api-host peertube.social -log-level 3

# Using double dash (equivalent)
./peertubeExportStat --api-host peertube.social --log-level 3

# Using .env file (./peertubeExportStat.env)
# Note: CLI parameters will be IGNORED when .env is present
```

## .env File Example

```ini
# PeerTube Export Stat Configuration

API_HOST=peertube.social
LOG_LEVEL=3
SMTP_HOST=smtp.example.com
# Empty lines and comments are supported
SMTP_PASSWORD=
```
