# Server HTTP API

This document describes every HTTP endpoint exposed by the server, defined in
[`server/internal/server`](../server/internal/server).

## Base URL & conventions

- **Base URL:** `http://<host>:8080`
  The port comes from the `SERVER_PORT` env var and defaults to `:8080`
  (`server/config/config.go`).
- **Path parameters** use Go 1.22+ `net/http` patterns, e.g. `{serial}`, `{location}`,
  `{name}`.
- **Routing is path-exact.** Each route only accepts the method(s) listed below;
  any other method returns `405 Method Not Allowed`.
- **CORS / headers** (set on most JSON responses):
  - `Access-Control-Allow-Origin: *`
  - `Access-Control-Allow-Methods: GET, POST, OPTIONS`
  - `Access-Control-Allow-Headers: Content-Type`
  - `Content-Type: application/json`
- **Status response** — endpoints that don't return data reply with:
  ```json
  { "status": "ok" }
  ```
- **Errors** are returned as plain-text bodies (via `http.Error`) with an
  appropriate status code: `400` (bad request / missing params), `405` (wrong
  method), `500` (internal failure).
- ⚠️ **No authentication or TLS.** Intended for a trusted local network only.

## Endpoint overview

| Method | Path | Description |
| ------ | ---- | ----------- |
| GET | `/ping` | Health check |
| GET | `/devices` | List connected ADB devices |
| GET | `/locations` | List location (screen) names |
| GET | `/locations/{location}` | List script names saved under a location |
| DELETE | `/locations/{location}` | Delete a location and all its scripts |
| GET | `/locations/{location}/{name}` | Get a single script |
| DELETE | `/locations/{location}/{name}` | Delete a script |
| POST | `/run_scripts` | Queue one or more scripts to run in order on a device |
| POST | `/save_rectangle` | Crop and save a template image from the device screen |
| POST | `/save_script` | Create or replace a script |
| POST | `/edit_script` | Update a script, moving its files if the location changed |
| GET | `/devices/{serial}/find_text` | OCR: locate text on the current screen |
| GET | `/devices/{serial}/keyboard` | Detect on-screen keyboard keys |
| POST | `/devices/{serial}/edit_keyboard` | Override a keyboard key's rectangle |
| GET | `/devices/{serial}/reset_keyboard` | Reset keyboard keys to defaults |
| GET | `/devices/{serial}/delete_button` | Delete a keyboard key override |
| POST | `/devices/{serial}/session` | Open a session: start scrcpy and the streaming sockets |
| GET | `/devices/{serial}/session` | Get the active session's ports |
| DELETE | `/devices/{serial}/session` | Close the session |

## Devices

### `GET /ping`

Health check.

- **Response `200`:** `{ "status": "ok" }`

### `GET /devices`

Returns all ADB devices currently visible to the server.

- **Response `200`:**
  ```json
  {
    "devices": [
      {
        "serial": "ABCD1234",
        "brand": "google",
        "device": "raven",
        "locale": "en-US",
        "model": "Pixel 6 Pro",
        "os_version": "14",
        "manufacturer": "Google",
        "marketing_name": "Pixel 6 Pro",
        "scrcpy_running": false
      }
    ]
  }
  ```
  `devices` is `[]` when none are connected.

## Locations & Scripts

Scripts are organised into **locations** — a location is a screen, and the scripts under
it are the interactions recorded on that screen. A script is identified by its
`location` + `name` pair. On disk it lives at `locations/<location>/<name>/run.json`, with
template images stored next to it as `<param.value>.png`.

Locations and scripts are stored server-wide: listing, fetching, and deleting are
device-independent. Only running a script targets a specific device.

### `GET /locations`

Lists the saved location (screen) names.

- **Response `200`:** array of strings.
  ```json
  ["main_screen", "profile"]
  ```
- **Errors:** `500` on read failure.

### `GET /locations/{location}`

Lists the names of the scripts saved under a location.

- **Path params:** `location` (required).
- **Response `200`:** array of strings.
  ```json
  ["open_profile", "open_settings"]
  ```
- **Errors:** `500` on read failure.

### `DELETE /locations/{location}`

Deletes a whole location — its directory and every script inside it.

- **Path params:** `location` (required).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` if `location` is empty.

### `GET /locations/{location}/{name}`

Returns a single script.

- **Path params:** `location`, `name` (both required).
- **Response `200`:** a [`Script`](#script) object.
- **Errors:** `400` if `location` is empty, `500` on read failure.

### `DELETE /locations/{location}/{name}`

Deletes a single script (its directory, including template images).

- **Path params:** `location`, `name` (both required).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` if `location` is empty.

### `POST /run_scripts`

Queues one or more scripts for execution on a device, in the given order. If
the device has no active session, one is opened first. Every entry is validated
up front (the script must exist and be non-empty) — one bad entry rejects the
whole batch and nothing is queued. Entries are then appended to the session's
queue as `location/name`; a per-session worker pops them one by one and
executes them. Progress is observable via `GET /devices/{serial}/session`:
`running <name> on <location>` while executing, `idle` when the queue drains,
or an error text (e.g. `unable to find parameter <type> with value <value> in
script <location>/<name>`) — a failure also clears the remaining queue.

During execution the script's `params` are located on screen **in order** —
each must be found before the script's `timeout` expires — progressively
identifying the element to act on. The script's `events` are then replayed over
the scrcpy control socket, offset to the **last** parameter's matched region. A
`type_text` parameter instead types its `value` on the on-screen keyboard. If
`params` is empty and `events` is not, the events are replayed exactly as
recorded.

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "scripts": [
      { "location": "main_screen", "name": "open_profile" },
      { "location": "profile", "name": "logout" }
    ]
  }
  ```
  `serial` and a non-empty `scripts` array are required; every entry needs both
  `location` and `name`.
- **Response `200`:** `{ "status": "ok" }` — the scripts were queued, not yet run.
- **Errors:** `400` on invalid JSON, empty `serial`, or an empty `scripts`
  array; `500` if an entry is invalid, its script is empty/missing, or the
  session could not be started.

### `POST /save_rectangle`

Takes a screenshot of the device, crops the given rectangle out of it, and
saves it as `<value>.png` in the script's directory
(`locations/<location>/<name>/`, created if missing). This is the template image
that a `template` parameter with the same `value` is matched against.

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "location": "main_screen",
    "name": "open_profile",
    "value": "login_button",
    "rectangle": { "left_x": 100, "right_x": 300, "top_y": 200, "bottom_y": 260 }
  }
  ```
  All fields are required; `rectangle` must be non-empty.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or a missing field, `500` if the screenshot
  or crop fails.

### `POST /save_script`

Creates a new script or replaces an existing one with the same `location` + `name`.
The body is a whole [`Script`](#script) object.

- **Request body:**
  ```json
  {
    "name": "open_profile",
    "location": "main_screen",
    "next_location": ["profile"],
    "params": [
      { "type": "text", "value": "Profile" },
      { "type": "template", "value": "" }
    ],
    "events": [ /* Event, see models */ ],
    "timeout": 15
  }
  ```
  `name` and `location` are required (both are trimmed). The body is stored as-is
  to `locations/<location>/<name>/run.json`. `timeout` is optional (seconds to keep
  locating a parameter's target before failing; values omitted or `<= 0` fall
  back to `15` at run time). Template images referenced by `template`
  parameters are saved separately via
  [`POST /save_rectangle`](#post-save_rectangle).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or empty `location`/`name`, `500` if not saved.

### `POST /edit_script`

Updates an existing script. Unlike [`POST /save_script`](#post-save_script),
it also handles a location change: when `prev_location` differs from the
script's `location`, the whole script directory — `run.json` plus its template
PNGs — is moved from `locations/<prev_location>/<name>/` to
`locations/<location>/<name>/` before the updated script is written.

- **Request body:**
  ```json
  {
    "prev_location": "main_screen",
    "script": { /* Script, see models */ }
  }
  ```
  `script.name` and `script.location` are required. `prev_location` is the
  location the script is currently stored under; when it is empty or equal to
  `script.location`, the endpoint behaves like `save_script`.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or empty `script.location`/`script.name`,
  `500` if the move or the save fails.

### `GET /devices/{serial}/find_text`

Takes a screenshot and runs OCR to locate matching text on the current screen.

- **Path params:** `serial` (required).
- **Query params:**
  - `text` — text to search for.
  - `locale` — OCR language/locale.
- **Response `200`:** array of [`OCRResult`](#ocrresult).
  ```json
  [{ "text": "Login", "rectangle": { "left_x": 40, "right_x": 220, "top_y": 900, "bottom_y": 980 } }]
  ```
- **Errors:** `400` if `serial` is empty.

## Keyboard

### `GET /devices/{serial}/keyboard`

Detects on-screen keyboard keys via template matching / OCR.

- **Path params:** `serial` (required).
- **Query params:** `locale`.
- **Response `200`:**
  ```json
  { "buttons": [ { "text": "a", "rectangle": { "left_x": 0, "right_x": 50, "top_y": 1500, "bottom_y": 1560 } } ] }
  ```
  `buttons` is `[]` when nothing is detected.
- **Errors:** `400` if `serial` is empty.

### `POST /devices/{serial}/edit_keyboard`

Overrides the rectangle for a single keyboard key.

> Note: `serial` is taken from the **request body**, not the path, for this
> endpoint (the path `serial` is ignored by the handler).

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "locale": "en",
    "name": "a",
    "rectangle": { "left_x": 0, "right_x": 50, "top_y": 1500, "bottom_y": 1560 }
  }
  ```
  `serial`, `name`, and a non-empty `rectangle` are required.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or missing fields, `500` if the edit fails.

### `GET /devices/{serial}/reset_keyboard`

Resets keyboard keys to their detected defaults.

- **Path params:** `serial` (required).
- **Query params:**
  - `locale`
  - `upper_case` — `true` to reset the upper-case layout (any other value = false).
- **Response `200`:** `{ "buttons": [ OCRResult, ... ] }`
- **Errors:** `400` if `serial` is empty.

### `GET /devices/{serial}/delete_button`

Deletes a saved keyboard key override.

- **Path params:** `serial` (required).
- **Query params:**
  - `locale`
  - `name` — key name to delete.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` if `serial` is empty, `500` if the deletion fails.

## Session / streaming

### `POST /devices/{serial}/session`

Opens a session: starts the scrcpy server on the device and opens the
streaming/control sockets. On success the server returns the TCP ports the
client should connect to, then asynchronously begins accepting the video, CV,
and control connections on those ports.

- **Path params:** `serial` (required).
- **Response `200`:**
  ```json
  {
    "video_port": "3002",
    "cv_port": "3003",
    "control_port": "3004"
  }
  ```
  Ports are derived from the server's base socket port (`SOCKET_PORT` env,
  default `3001`). The three ports are **raw TCP sockets**, not HTTP — the
  client connects to them directly to receive the H.264 video stream, CV
  results, and to send control commands.
- **Errors:** `500` if scrcpy could not be started (the session is then
  closed server-side).

### `GET /devices/{serial}/session`

Returns the session's status.

- **Response `200`:** `{ "status": "<status>" }` where `<status>` is one of:
  - `closed` — no active session for this serial;
  - `idle` — session is open, script queue is empty;
  - `running <script> on <location>` — a queued script is executing;
  - an error text (e.g. `unable to find parameter <type> with value <value> in
    script <location>/<name>`) — the last queued script failed; the queue was
    cleared. The error stays until the next script is queued.

### `DELETE /devices/{serial}/session`

Closes the session: stops the scrcpy server and tears down the sockets.

- **Response `200`:** `{ "status": "ok" }`

## Data models

### AdbDevice

`server/pkg/models/adb_device.go`

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Serial | `serial` | string | Device serial |
| Brand | `brand` | string | |
| Device | `device` | string | |
| Locale | `locale` | string | First system locale |
| Model | `model` | string | |
| OsVersion | `os_version` | string | |
| Manufacturer | `manufacturer` | string | |
| MarketingName | `marketing_name` | string | |
| ScrcpyRunning | `scrcpy_running` | bool | Whether scrcpy is currently running |

### Rectangle

`server/pkg/models/rectangle.go`

```json
{ "left_x": 0, "right_x": 0, "top_y": 0, "bottom_y": 0 }
```

| Field | JSON | Type |
| ----- | ---- | ---- |
| LeftX | `left_x` | int |
| RightX | `right_x` | int |
| TopY | `top_y` | int |
| BottomY | `bottom_y` | int |

### Script

`server/pkg/models/script.go`

```json
{
  "name": "open_profile",
  "location": "main_screen",
  "next_location": ["profile"],
  "params": [ /* Parameter */ ],
  "events": [ /* Event */ ],
  "timeout": 15
}
```

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Name | `name` | string | Script name; with `location` it forms the script's identity |
| Location | `location` | string | The screen this script is recorded on |
| NextLocation | `next_location` | string array | omitempty; the screens this script can lead to |
| Params | `params` | [Parameter](#parameter) array | Locators matched **in order** to narrow the interactive zone |
| Events | `events` | [Event](#event) array | omitempty; replayed against the **last** matched param, or verbatim when `params` is empty |
| Timeout | `timeout` | int | Seconds to locate a param's target before failing (default `15` when omitted or `<= 0`) |

### Parameter

A parameter locates an element on screen. Parameters are matched **in order**,
progressively identifying the element to act on; the script's `events` are
applied to the last one's region.

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Type | `type` | string | One of `template`, `text`, `type_text`, `yolo_class`, `command` |
| Value | `value` | string | Template image name for `template` (matched against `<value>.png`), OCR text for `text`, YOLO class name for `yolo_class`, text to type for `type_text` |
| Locale | `locale` | string | omitempty; OCR language/locale, and the keyboard locale for `type_text` |

### Event

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Time | `time` | int64 | Timestamp / delay |
| Data | `data` | ControlBytes | Serialized as an **array of ints** (one per byte), 32 bytes (`ControlBytesSize`) |

### OCRResult

`server/internal/cv/model.go`

```json
{ "text": "Login", "rectangle": { "left_x": 40, "right_x": 220, "top_y": 900, "bottom_y": 980 } }
```

| Field | JSON | Type |
| ----- | ---- | ---- |
| Text | `text` | string |
| Rectangle | `rectangle` | [Rectangle](#rectangle) |
