# PeerTube Stats CLI Usage

## Important Notes

- <b>Every flag can be used with double dashes</b> (e.g., `--help` is valid)
- <b>A .env file in the working directory of the process is supported without dashes</b>
- <b>BUG ALERT: When using a .env file or environment variables, CLI parameters are IGNORED</b>
- The .env file supports:
    - Comments
    - Empty lines
    - Empty values are treated as unset

## Flags Reference

| Flag                                                                           | Description                              | Default Value                      |
|--------------------------------------------------------------------------------|------------------------------------------|------------------------------------|
| `-api-client-id` / `--api-client-id`                                           | Client ID                                | `"exampleID"`                      |
| `-api-client-secret` / `--api-client-secret`                                   | Client Secret                            | `"exampleSecret"`                  |
| `-api-host` / `--api-host`                                                     | Host to authenticate with                | `"peertube.example.com"`           |
| `-api-password` / `--api-password`                                             | Password to authenticate with            | `"examplePassword"`                |
| `-api-protocol` / `--api-protocol`                                             | Protocol to authenticate with            | `"https"`                          |
| `-api-username` / `--api-username`                                             | Username to authenticate with            | `"exampleUser"`                    |
| `-bind-address` / `--bind-address`                                             | Bind address                             | `"127.0.0.1"`                      |
| `-cache-valid-seconds` / `--cache-valid-seconds`                               | Video database cache validity (seconds)  | `90000` (slightly more than a day) |
| `-data-folder` / `--data-folder`                                               | Folder containing video stats            | `"./Data"`                         |
| `-http-port` / `--http-port`                                                   | HTTP port                                | `8080`                             |
| `-log-level` / `--log-level`                                                   | Logging level                            | `2` (warning)                      |
| Log Level Values                                                               |                                          |                                    |
| `0`                                                                            | Fatal                                    |                                    |
| `1`                                                                            | Error                                    |                                    |
| `2`                                                                            | Warning                                  |                                    |
| `3`                                                                            | Info                                     |                                    |
| `4`                                                                            | Debug                                    |                                    |
| `-max-concurrent-request-connections` / `--max-concurrent-request-connections` | Max concurrent request connections       | `10`                               |
| `-max-request-size` / `--max-request-size`                                     | Max request size                         | `1048576`                          |
| `-miss-tolerance` / `--miss-tolerance`                                         | Tolerance of days for missing statistics | *not specified*                    |
| `-request-timeout` / `--request-timeout`                                       | Request timeout in seconds               | `-1`                               |
| `-stat-io-max-threads` / `--stat-io-max-threads`                               | Max number of threads to use             | `10`                               |

### .env File Example

```
# This is a comment
API_CLIENT_ID=myClientId
API_CLIENT_SECRET=myClientSecret
API_HOST=peertube.example.com
# Empty lines are allowed
API_PASSWORD=
```