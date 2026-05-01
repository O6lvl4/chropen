# chropen

Open URLs in a specific Chrome profile, by email.

When you've got several Google accounts in Chrome — work, personal,
client A, client B — opening a billing portal in the *wrong* profile
is a daily papercut. `chropen` resolves an email address to the
matching Chrome profile directory and launches Chrome with
`--profile-directory` set to it.

```sh
chropen alice@example.com https://gmail.com
chropen admin@example.com https://admin.google.com
chropen --list                                  # see all detected profiles
```

You can also pass a literal directory name (`Default`, `Profile 24`),
so existing tooling that already knows the dir keeps working
unchanged.

## Install

```sh
go install github.com/O6lvl4/chropen/cmd/chropen@latest
```

Or build from source:

```sh
git clone https://github.com/O6lvl4/chropen
cd chropen
go build -o chropen ./cmd/chropen
```

## CLI

```text
chropen — open URLs in a specific Chrome profile

Usage:
  chropen <email-or-profile-dir> <url> [<url>...]
  chropen --list
  chropen --resolve <email-or-profile-dir>

Environment:
  CHROME_BIN             override Chrome binary path
  CHROME_USER_DATA_DIR   override Chrome user-data directory
```

`--list` example output:

```
  Default       alice@example.com                      Work
  Profile 1     alice.personal@example.com             Personal
  Profile 4     alice@client-a.example                 Client A
  Profile 7     archive@example.com                    Archived
```

`--resolve` is handy in scripts: it prints the directory name so you
can pipe it elsewhere.

```sh
chropen --resolve alice@example.com   # → "Profile 7"
```

## As a library

```go
import "github.com/O6lvl4/chropen/profile"

dir, err := profile.Resolve("alice@example.com")  // "Profile 7"
err = profile.Open(dir, "https://example.com")
err = profile.OpenAs("alice@example.com", "https://a", "https://b")
infos, err := profile.List()
```

## How it works

Chrome stores per-profile metadata under its user-data dir:

- macOS: `~/Library/Application Support/Google/Chrome/`
- Linux: `~/.config/google-chrome/`
- Windows: `%LOCALAPPDATA%\Google\Chrome\User Data\`

`chropen` reads `Local State` (the master JSON listing every profile's
signed-in account) and falls back to per-profile `Preferences` for
profiles missing from `Local State`. Both lookups are case-insensitive
on the email.

Once it has a directory name, it launches Chrome directly:

```text
"$CHROME_BIN" --profile-directory="<dir>" <url>...
```

so the URLs always land in the right window, even when Chrome is
already running.

## License

MIT
