# Server HTTP API

This document describes every HTTP endpoint exposed by the server, defined in
[`server/internal/server`](../server/internal/server).

## Base URL & conventions

- **Base URL:** `http://<host>:8080`
  The port comes from the `SERVER_PORT` env var and defaults to `:8080`
  (`server/config/config.go`).
- **Path parameters** use Go 1.22+ `net/http` patterns, e.g. `{serial}`, `{name}`.
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
| GET | `/devices/{serial}/scripts` | List script names for a device |
| GET | `/devices/{serial}/scripts/{name}` | Get a single script |
| DELETE | `/devices/{serial}/scripts/{name}` | Delete a script |
| GET | `/devices/{serial}/scripts/{name}/run` | Run a script |
| POST | `/save_rectangle` | Save a selected zone (rectangle) |
| POST | `/save_step` | Append a step to a script |
| GET | `/devices/{serial}/find_text` | OCR: locate text on the current screen |
| GET | `/devices/{serial}/keyboard` | Detect on-screen keyboard keys |
| POST | `/devices/{serial}/edit_keyboard` | Override a keyboard key's rectangle |
| GET | `/devices/{serial}/reset_keyboard` | Reset keyboard keys to defaults |
| GET | `/devices/{serial}/delete_button` | Delete a keyboard key override |
| GET | `/start_sockets/{serial}` | Start scrcpy and open streaming sockets |

---

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

---

## Scripts

### `GET /devices/{serial}/scripts`

Lists the names of all saved scripts for the device.

- **Path params:** `serial` — device serial (required).
- **Response `200`:** array of strings.
  ```json
  ["login_flow", "open_chat"]
  ```
- **Errors:** `400` if `serial` is empty, `500` on read failure.

### `GET /devices/{serial}/scripts/{name}`

Returns a single script and its steps.

- **Path params:** `serial`, `name` (both required).
- **Response `200`:** a [`Script`](#script) object.
- **Errors:** `400` if `serial` or `name` is empty, `500` on read failure.

### `DELETE /devices/{serial}/scripts/{name}`

Deletes a script.

- **Path params:** `serial`, `name` (both required).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` if `serial` or `name` is empty.

### `GET /devices/{serial}/scripts/{name}/run`

Executes the named script on the device. The script replays its recorded
control events over the scrcpy control socket (using the server's configured
socket port).

- **Path params:** `serial`, `name`.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `500` if execution fails.

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

### `POST /save_step`

Appends a step to the named script.

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "name": "login_flow",
    "step": { /* ScriptStep, see models */ }
  }
  ```
  `serial` and `name` are required.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON or empty `serial`/`name`, `500` if not saved.

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

---

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

---

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

---

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
  "name": "login_flow",
  "steps": [ /* ScriptStep */ ]
}
```

### ScriptStep

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| ID | `id` | int | omitempty |
| Events | `events` | [Event](#event) array | omitempty |
| Flags | `flags` | int | Bit flags: `EventOnTemplate=1`, `EventOnText=2`, `TypeText=4`, `TemplateIsVisible=8`, `TextIsVisible=16` |
| Text | `text` | string | omitempty |
| Locale | `locale` | string | omitempty |
| Command | `command` | string | omitempty |

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
