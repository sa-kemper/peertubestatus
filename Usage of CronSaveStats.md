# Usage of ./CronSaveStats

## Command-Line Flags

### Notes

- **All flags can be used with either single dash (-) or double dash (--) syntax. For example, both `-help` and `--help` are valid.**

---

### Authentication Flags

| Flag                            | Description                                          | Default Value             |
|---------------------------------|------------------------------------------------------|---------------------------|
| `-api-client-id`               | Client ID                                           | `"exampleID"`            |
| `-api-client-secret`           | Client Secret                                       | `"exampleSecret"`        |
| `-api-host`                    | Host to authenticate with                            | `"peertube.example.com"` |
| `-api-password`                 | Password to authenticate with                        | `"examplePassword"`      |
| `-api-protocol`                | Protocol to authenticate with                        | `"https://"`             |
| `-api-username`                | Username to authenticate with                        | `"exampleUser"`          |

---

### Configuration Flags

| Flag                            | Description                                          | Default Value             |
|---------------------------------|------------------------------------------------------|---------------------------|
| `-cache-valid-seconds`         | Validity of video database cache in seconds         | `90000`                   |
| `-data-folder`                 | Folder containing video stats                        | `"./Data"`                |
| `-log-level`                   | Level of logging (0 to 4)                           | `2` (warning)             |
| `-miss-tolerance`              | Tolerance for missing statistic days                 | *Not set*                 |

---

### SMTP Configuration Flags

| Flag                            | Description                                          | Default Value             |
|---------------------------------|------------------------------------------------------|---------------------------|
| `-smtpFromAddress`             | SMTP from address                                   | `"peertubestats@example.com"` |
| `-smtpHost`                    | SMTP server host                                    | `"smtp.example.com"`      |
| `-smtpPassword`                | SMTP password                                       | *Not set*                 |
| `-smtpPort`                    | SMTP server port                                    | `25`                      |
| `-smtpToAddress`               | Recipient list for administrators                   | `"administrator@example.com"` |
| `-smtpUsername`                 | SMTP username                                       | *Not set*                 |

---

### Utility Flags

| Flag                            | Description                                          | Default Value             |
|---------------------------------|------------------------------------------------------|---------------------------|
| `-test-mail`                   | Test mail                                           | *Not set*                 |
| `-stat-io-max-threads`         | Maximum number of threads                           | `10`                      |

---

## Environment Configuration

### .env File Support

**The application supports configuration via a .env file located in the working directory of the process.**

Key features of the .env file include:
- Supports comments (lines starting with `#`)
- Supports empty lines
- Empty values are treated as unset
- Can be used without dashes (e.g., `API_CLIENT_ID=value` instead of `-api-client-id`)

---

### Important Bug Note

**CRITICAL: If a .env file or environment variables are used, command-line parameters will be IGNORED.**

This means that environment-based configuration takes complete precedence over CLI parameters, potentially leading to unexpected behavior if not carefully managed.

---

### Example .env File:

```env
# Authentication settings
API\_CLIENT\_ID=myClientId
API\_CLIENT\_SECRET=myClientSecret

# SMTP Configuration
SMTP\_HOST=smtp.example.com
SMTP\_PORT=25

# Empty line and comment are allowed above
LOG\_LEVEL=3  # This sets the log level
```