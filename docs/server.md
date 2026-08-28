### Installation

From the project root, run:

```bash
./install.sh
```

The script installs the required dependencies (Go, ADB, FFmpeg, OpenCV,
Tesseract OCR), builds the server, and installs the `vdroid-scripter` binary to
`/usr/local/bin`. Supported out of the box: macOS (Homebrew), Arch,
Debian/Ubuntu, and Fedora.

The server needs **OpenCV 4** (gocv does not support OpenCV 5 yet). On macOS
that is the `opencv@4` formula. OCR runs in-process against libtesseract, so
the Tesseract headers are a build dependency too: the brew and Arch
`tesseract` packages include them, Debian/Ubuntu needs `libtesseract-dev`
and Fedora `tesseract-devel` (the script installs these). On Arch, whose official `opencv` package is
now version 5, the script installs the AUR `opencv4` package instead: through
plain `pacman` when a configured repo carries it (e.g. chaotic-aur, prebuilt),
otherwise through `paru` or `yay`; with no AUR helper available it stops and
asks you to install `opencv4` manually before re-running. Debian/Ubuntu and
Fedora still ship OpenCV 4 as their default packages.

By default the binary goes to `/usr/local/bin` (the script uses `sudo` only if
that directory is not writable). Set `PREFIX` to install into a custom directory
instead — the binary is placed in `$PREFIX/bin`. For example,
`PREFIX="$HOME/.local"` installs to `~/.local/bin` (a user-local install that
needs no `sudo`), and `PREFIX="/opt/vdroid"` installs to `/opt/vdroid/bin`.

```bash
PREFIX="$HOME/.local" ./install.sh
```

If the chosen `$PREFIX/bin` is not already on your `PATH`, the script prints the
`export PATH=...` line to add to your shell profile.

### Preparing the Android Device (Target Device)

Before connecting the Android device to the server, make sure it is properly configured for remote control:

1. Open **Settings** on the Android device
2. Enable **Developer Options**
   - (Usually by tapping *Build Number* 7 times in “About Phone”)
3. Inside Developer Options, enable:
   - **USB Debugging**
4. Connect the device to your computer using a USB cable
5. When prompted on the device, accept the **RSA fingerprint / ADB debugging authorization**
6. Verify that the device is detected:
```bash
adb devices
```

### Running the Server

Connect a device and start the server:

```bash
vdroid-scripter
```

For development, you can run from source instead:

```bash
cd server && go run cmd/main.go
```

On macOS, building or running from source needs the Homebrew `opencv@4` and
`ffmpeg` formulas on the pkg-config path first (`install.sh` does this
internally when building). Any recent FFmpeg major works — the server binds
only libavcodec's stable decode core:

```bash
export PKG_CONFIG_PATH="$(brew --prefix opencv@4)/lib/pkgconfig:$(brew --prefix ffmpeg)/lib/pkgconfig:$PKG_CONFIG_PATH"
```

### Building the MCP Server

`server/mcpserver` contains an MCP (Model Context Protocol) server that exposes
the HTTP API as tools for AI-driven flow building. It needs only Go — none of
the CV native libraries:

```bash
cd server && go build -o vdroid-mcp ./mcpserver
```

The `-o` flag is required: without it the default output name collides with the
`mcpserver` directory. Register the binary with your MCP client, e.g. for
Claude Code:

```bash
claude mcp add vdroid -- <path>/vdroid-mcp
```

The MCP server talks to a running vdroid server at the URL from the
`VDROID_URL` env var (default `http://127.0.0.1:8080`).

### API Reference

The server exposes an HTTP API for managing devices, scripts, the on-screen
keyboard, and streaming sockets. See [api](api.md) for the full list of
endpoints, request/response formats, and data models.

### Network & Security Notice

⚠️ **Important:** The communication between the server and client is **not encrypted (no TLS or encryption layer)**.

This project is intended to be used in a **trusted local network environment only**, such as:

- Personal home networks
- Development environments
- Isolated testing setups

It is **not recommended** to expose the server to the public internet or untrusted networks, as all transmitted data (including screen frames and control commands) can be intercepted.

For production or internet-facing usage, additional security layers (such as VPN, SSH tunneling, or TLS encryption) must be implemented manually.
