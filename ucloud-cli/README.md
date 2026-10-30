# Ucloud completion

Completion for Ucloud CLI.

## Build and publish the download image

Set up Docker Buildx and registry access, then run:

```
./build-container.sh
```

This builds Linux and macOS binaries for amd64 and arm64. It publishes versioned and `latest` images to the Core registry. Set the CLI version in `version.txt` before a release.

The script uses BuildKit named contexts for the local modules in `provider-integration`. It mounts those modules during compilation.

## Install

The installer selects the binary for the current operating system and architecture. It checks the download checksum and installs to `~/.local/bin`.

```
curl -fsSL 'https://cloud.sdu.dk/cli/install.sh' | UCLOUD_CLI_BASE_URL='https://cloud.sdu.dk/cli' bash
```

### Build

run to build the `ucloud` executable

```
./build.sh
```

### Browser login

For installations where the external CLI login service is unavailable, use:

```
ucloud connect --web
```

This opens an isolated Chrome/Chromium window at the selected environment's normal
login page. Complete institutional login and MFA yourself. The CLI reads only the
UCloud refresh cookie from this temporary profile and validates it using
`POST /auth/refresh`, then saves the renewable session in
`~/.config/ucloud/config.yaml` (owner read/write only). Subsequent commands use the
existing CLI token refresh mechanism. The browser closes and its temporary profile
is removed after login, cancellation, or the ten-minute timeout. Existing browser
profiles are not accessed. Passwords and institutional cookies are not captured.

Chrome/Chromium must already be installed. To select its executable explicitly:

```
ucloud connect --web --browser /path/to/chromium
```

The server must use HTTPS, except for loopback development servers. A newly saved
session is bound to that server; connect again after switching to another server.
Only one session is stored at a time. Run the command again if the server expires
or revokes the session. This is a client-side adaptation of web login, not an
upstream-supported production CLI authentication guarantee: deployments must
accept the web session at `/auth/refresh`. Failed validation leaves existing
credentials unchanged. The original `ucloud connect` callback flow remains available.

### Usage

to make it globally available
either cp `ucloud` to `/usr/local/bin` or add to your `$PATH`

to generate completion script:

```
mkdir -p ~/.config/ucloud/completion
ucloud completion > ~/.config/ucloud/completion/_ucloud
```

Add to your `.bashrc` or `.zshrc`

```
export PATH="~/.config/ucloud/completion:$PATH"
fpath=("~/.config/ucloud/completion" $fpath)

#Ensure that completion system is initialized
autoload -Uz compinit
compinit

# Load/register ucloud completion
autoload -Uz _ucloud
compdef _ucloud ucloud

```
