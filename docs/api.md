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
| POST | `/run_steps` | Queue one or more steps to run in order on a device |
| GET | `/devices/{serial}/find_text` | OCR: locate text on the current screen |
| GET | `/library` | List image and action names in the library |
| POST | `/save_image` | Crop and save a named template image into the library |
| POST | `/save_action` | Save a named recorded gesture into the library |
| DELETE | `/images/{name}` | Delete a library image |
| DELETE | `/actions/{name}` | Delete a library action |
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

## Steps

A **step** is the unit of execution: an **event** applied to a CV-located
target. Steps compose the [library](#library) at runtime — the target
(`type` + `value`) locates a region on the live screen (a library image, OCR
text, or a YOLO class) and the event acts on it (generated taps, CV keyboard
typing, or a recorded gesture from the library).

| `event` | Behavior |
| ------- | -------- |
| *(empty)* | Visibility check of the target — no touch. Target required. |
| `tap` / `long_tap` | Generated tap pair placed at a random point inside the found region. Target required. |
| `type_text` | `value` **is the text to type**, typed via the CV keyboard (`locale` = keyboard locale). No target. |
| any other name | The library event with that name is replayed: **anchored** when a target is given (first touch moved into the found region, relative shape preserved), **verbatim** without one. |

### `POST /run_steps`

Queues one or more steps to run in order on a device. If the device has no open
session, one is opened automatically (scrcpy is started). All steps are
validated up front — library images and events referenced by the steps must
exist — then appended to the session's queue. A per-session worker executes
steps sequentially, updating the session status; a step failure sets the error
status and clears the remaining queue.

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "steps": [
      { "type": "yolo", "value": "home" },
      { "event": "tap", "type": "image", "value": "catalog_cart_icon" },
      { "event": "swipe_catalog_1" },
      { "event": "tap", "type": "text", "value": "Corn", "locale": "eng", "timeout": 10 },
      { "event": "type_text", "value": "hello", "locale": "eng" }
    ]
  }
  ```
  `serial` and a non-empty `steps` array are required; every entry must be a
  valid [`Step`](#step).
- **Response `200`:** `{ "status": "ok" }` — the steps were queued. Track
  execution via [`GET /devices/{serial}/session`](#get-devicesserialsession).
- **Errors:** `400` on invalid JSON or an invalid step, `500` when a referenced
  library image/action does not exist or the session could not be started.

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

## Library

The library holds the reusable, uniquely named building blocks that steps
compose: **images** (template crops, stored as `images/<name>.png`) and
**actions** (recorded gesture event streams, stored as
`actions/<name>.json`). Names are flat — pick descriptive ones like
`catalog_cart_icon` or `swipe_catalog_1`, since the name is the only
context an item carries. Saving under an existing name overwrites the item.

### `GET /library`

Lists everything in the library.

- **Response `200`:**
  ```json
  {
    "images": ["tg_chat_send_button", "catalog_cart_icon"],
    "actions": ["swipe_catalog_1", "swipe_catalog_2"]
  }
  ```
  Both arrays are sorted and `[]` when empty.

### `POST /save_image`

Takes a screenshot of the device, crops the given rectangle out of it, and
saves it as `images/<name>.png` — the template that an `image` step target
with the same `value` is matched against.

- **Request body:**
  ```json
  {
    "serial": "ABCD1234",
    "rectangle": { "left_x": 100, "right_x": 300, "top_y": 200, "bottom_y": 260, "label": "catalog_cart_icon" }
  }
  ```
  The image name comes from `rectangle.label` — required, no path separators.
  `serial` and a non-empty `rectangle` are required. An existing image with the
  same name is overwritten.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON, a bad `label`, or a missing field, `500` if
  the screenshot or crop fails.

### `POST /save_action`

Saves a recorded gesture as `actions/<name>.json`. The body is a whole
[`Action`](#action) object.

- **Request body:**
  ```json
  {
    "name": "swipe_catalog_1",
    "screen_width": 1080,
    "screen_height": 2400,
    "events": [ /* Event, see models */ ]
  }
  ```
  `name` (trimmed, no path separators) and a non-empty `events` array are
  required. An existing action with the same name is overwritten.
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON, a bad `name`, or empty `events`, `500` if
  not saved.

### `DELETE /images/{name}`

Deletes `images/<name>.png`.

- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `404` if no image with that name exists.

### `DELETE /actions/{name}`

Deletes `actions/<name>.json`.

- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `404` if no action with that name exists.

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
  - `idle` — session is open, step queue is empty;
  - `running <step>` (e.g. `running tap on image catalog_cart_icon`) — a
    queued step is executing;
  - an error text (e.g. `unable to find <target type> <value> on screen`) —
    the last queued step failed; the queue was cleared. The error stays until
    the next step is queued.

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
{ "left_x": 0, "right_x": 0, "top_y": 0, "bottom_y": 0, "label": "" }
```

| Field | JSON | Type |
| ----- | ---- | ---- |
| LeftX | `left_x` | int |
| RightX | `right_x` | int |
| TopY | `top_y` | int |
| BottomY | `bottom_y` | int |
| Label | `label` | string (omitempty; YOLO class in detections, the image name in `save_image`) |

### Step

`server/pkg/models/step.go`

```json
{
  "event": "tap",
  "type": "image",
  "value": "catalog_cart_icon",
  "locale": "",
  "timeout": 15
}
```

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Event | `event` | string | `tap`, `long_tap`, `type_text`, the name of a library event, or **empty for a visibility check** |
| Type | `type` | string | Target type: `image` (template match against `images/<value>.png`), `text` (OCR), `yolo` (detection class); empty for a target-less library event and for `type_text` |
| Value | `value` | string | Target value: library image name, OCR text, or YOLO class name — except for `type_text`, where it is the text to type |
| Locale | `locale` | string | omitempty; OCR language for `text` targets, keyboard locale for `type_text` |
| Timeout | `timeout` | int | omitempty; seconds to locate the target before failing (default `15` when omitted or `<= 0`). The runner grabs the latest video frame and retries roughly once per second until the deadline. |

### Event

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Time | `time` | int64 | Timestamp / delay |
| Data | `data` | ControlBytes | Serialized as an **array of ints** (one per byte), 32 bytes (`ControlBytesSize`) |

### Action

`server/pkg/models/action.go`

```json
{
  "name": "swipe_catalog_1",
  "screen_width": 1080,
  "screen_height": 2400,
  "events": [ /* Event */ ]
}
```

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Name | `name` | string | Unique library name |
| ScreenWidth | `screen_width` | int | Screen width of the recording device |
| ScreenHeight | `screen_height` | int | Screen height of the recording device |
| Events | `events` | [Event](#event) array | The recorded gesture |

### OCRResult

`server/internal/cv/model.go`

```json
{ "text": "Login", "rectangle": { "left_x": 40, "right_x": 220, "top_y": 900, "bottom_y": 980 } }
```

| Field | JSON | Type |
| ----- | ---- | ---- |
| Text | `text` | string |
| Rectangle | `rectangle` | [Rectangle](#rectangle) |
