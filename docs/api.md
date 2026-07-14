# Server HTTP API

This document describes every HTTP endpoint exposed by the server, defined in
[`server/internal/server`](../server/internal/server).

## Base URL & conventions

- **Base URL:** `http://<host>:8080`
  The port comes from the `SERVER_PORT` env var and defaults to `:8080`
  (`server/config/config.go`).
- **Path parameters** use Go 1.22+ `net/http` patterns, e.g. `{serial}`, `{node}`,
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
| GET | `/scripts` | List script node (screen) names |
| GET | `/scripts/{node}/{name}` | Get a single script |
| DELETE | `/scripts/{node}/{name}` | Delete a script |
| GET | `/devices/{serial}/scripts/{node}/{name}/run` | Run a script on a device |
| POST | `/save_rectangle` | Save a selected zone (rectangle) |
| POST | `/save_script` | Create or replace a script |
| GET | `/devices/{serial}/find_text` | OCR: locate text on the current screen |
| GET | `/devices/{serial}/keyboard` | Detect on-screen keyboard keys |
| POST | `/devices/{serial}/edit_keyboard` | Override a keyboard key's rectangle |
| GET | `/devices/{serial}/reset_keyboard` | Reset keyboard keys to defaults |
| GET | `/devices/{serial}/delete_button` | Delete a keyboard key override |
| GET | `/start_sockets/{serial}` | Start scrcpy and open streaming sockets |

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

## Scripts

A script belongs to a **node** (the screen it is recorded on) and is identified by
the `node` + `name` pair. On disk it lives at
`scripts/<node>/<name>/run.json`, with template images stored next to it as
`<param.id>.png`.

Scripts are stored server-wide: listing, fetching, and deleting are
device-independent. Only running a script targets a specific device.

### `GET /scripts`

Lists the names of the top-level script directories — i.e. the **node (screen)
names** that have scripts saved under them.

- **Response `200`:** array of strings.
  ```json
  ["main_screen", "profile"]
  ```
- **Errors:** `500` on read failure.

### `GET /scripts/{node}/{name}`

Returns a single script.

- **Path params:** `node`, `name` (both required).
- **Response `200`:** a [`Script`](#script) object.
- **Errors:** `400` if `node` or `name` is empty, `500` on read failure.

### `DELETE /scripts/{node}/{name}`

Deletes a script (removes its whole directory, including template images).

- **Path params:** `node`, `name` (both required).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` if `node` or `name` is empty.

### `GET /devices/{serial}/scripts/{node}/{name}/run`

Executes the named script on the given device.

The script's `params` are matched **in order** to narrow down the interactive
zone (e.g. first locate a text label, then the checkbox nearest to it). The
script's `events` are then replayed over the scrcpy control socket, offset to the
**last** matched parameter's region. If `params` is empty and `events` is not,
the events are replayed exactly as recorded.

- **Path params:** `serial`, `node`, `name` (all required).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `500` if the script is not found/empty or execution fails.

### `POST /save_rectangle`

Saves a selected screen zone (rectangle) for the device.

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "rectangle": { "left_x": 100, "right_x": 300, "top_y": 200, "bottom_y": 260 }
  }
  ```
  `serial` is required.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or empty `serial`, `500` if not saved.

### `POST /save_script`

Creates a new script or replaces an existing one with the same `node` + `name`.
The body is a whole [`Script`](#script) object.

- **Request body:**
  ```json
  {
    "name": "open_profile",
    "node": "main_screen",
    "next_node": "profile",
    "params": [
      { "type": "text", "value": "Profile" },
      { "type": "template", "value": "" }
    ],
    "events": [ /* Event, see models */ ],
    "timeout": 15
  }
  ```
  `name` and `node` are required. `params[].id` is assigned by the server
  (sequentially, starting at `1`) — any value sent is overwritten. `timeout` is
  optional (seconds to keep locating a parameter's target before failing;
  defaults to `15` when omitted or `<= 0`).
- **Template images:** a pending zone saved via
  [`POST /save_rectangle`](#post-save_rectangle) is committed to the **last**
  `template` parameter as `<param.id>.png` in the script's directory.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or empty `node`/`name`, `500` if not saved.

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

## Sockets / streaming

### `GET /start_sockets/{serial}`

Starts the scrcpy server on the device and opens the streaming/control sockets.
On success the server returns the TCP ports the client should connect to, then
asynchronously begins accepting the video, CV, and control connections on those
ports.

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
- **Errors:** `500` if scrcpy could not be started (the connection is then
  closed server-side).

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
  "node": "main_screen",
  "next_node": "profile",
  "params": [ /* Parameter */ ],
  "events": [ /* Event */ ],
  "timeout": 15
}
```

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Name | `name` | string | Script name; with `node` it forms the script's identity |
| Node | `node` | string | The screen this script is recorded on |
| NextNode | `next_node` | string | omitempty; the screen this script leads to |
| Params | `params` | [Parameter](#parameter) array | Locators matched **in order** to narrow the interactive zone |
| Events | `events` | [Event](#event) array | omitempty; replayed against the **last** matched param, or verbatim when `params` is empty |
| Timeout | `timeout` | int | Seconds to locate a param's target before failing (default `15` when omitted or `<= 0`) |

### Parameter

A parameter locates an element on screen. Parameters are matched in order, each
narrowing the zone — later params resolve to the candidate nearest the previous
match (e.g. "the checkbox next to *this* label").

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| ID | `id` | int | Assigned by the server (from `1`); template images are named `<id>.png` |
| Type | `type` | string | One of `template`, `text`, `type_text`, `yolo_class`, `command` |
| Value | `value` | string | OCR text for `text`, YOLO class name for `yolo_class`; unused for `template` (matched by `<id>.png`) |
| Locale | `locale` | string | omitempty; OCR language/locale |

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
