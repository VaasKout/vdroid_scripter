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
| GET | `/devices/{serial}/screenshot` | Take an ADB screenshot, optionally with detected UI rectangles |
| POST | `/run_steps` | Queue one or more steps to run in order on a device |
| GET | `/scripts` | List saved script names |
| GET | `/scripts/{name}` | Read a script's steps |
| POST | `/scripts` | Save (overwrite) a named script |
| DELETE | `/scripts/{name}` | Delete a script |
| GET | `/run_script` | Queue a saved script's steps on a device |
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

### `GET /devices/{serial}/screenshot`

Takes a screenshot over ADB and returns it as multipart form data.

- **Path params:** `serial` (required).
- **Query params:** `rectangles` — `true` to also run UI rectangle detection
  (`FindAllRectangles`) on the screenshot.
- **Response `200`:** multipart form with an `image` file field (the
  screenshot) and, when `rectangles=true`, a `rectangles` field holding a JSON
  array of [`Rectangle`](#rectangle) (`[]` when nothing was detected).
- **Errors:** `400` if `serial` is empty, `500` when the screenshot could not
  be taken.

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

## Steps

A **step** is the unit of execution: an **event** applied to a chain of
CV-located **anchors**. Steps compose the [library](#library) at runtime —
each anchor (`type` + `value`) locates a region on the live screen (a library
image, OCR text, or a YOLO class) and the event acts on the **last** anchor's
region (generated taps, CV keyboard typing, or a recorded gesture from the
library).

The chain resolves on a single video frame: the first anchor picks its best
match on screen, every following anchor picks the candidate of its value
**nearest** (by center distance) to the previous anchor, and the event applies
to the last anchor. One anchor is the normal case; a preceding unique anchor
disambiguates duplicates — "the toggle next to *Show refresh rate*" is
`anchors: [{text "Show refresh rate"}, {image "toggle"}]`.

| `event` | Behavior |
| ------- | -------- |
| *(empty)* | Visibility check of the anchor chain — no touch. Anchors required. |
| `tap` / `long_tap` | Generated tap pair placed at a random point inside the last anchor's region. Anchors required. |
| `type_text` | The **last anchor's `value` is the text to type**, typed via the CV keyboard (its `locale` = keyboard locale). Nothing is located on screen. |
| any other name | The library event with that name is replayed: **anchored** when anchors are given (first touch moved into the last anchor's region, relative shape preserved), **verbatim** without them. |

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
      { "anchors": [{ "type": "yolo", "value": "home" }] },
      { "event": "tap", "anchors": [{ "type": "image", "value": "catalog_cart_icon" }] },
      { "event": "swipe_catalog_1" },
      { "event": "tap", "anchors": [{ "type": "text", "value": "Corn", "locale": "eng" }], "timeout": 10 },
      { "event": "tap", "anchors": [{ "type": "text", "value": "Show refresh rate" }, { "type": "image", "value": "toggle" }] },
      { "event": "type_text", "anchors": [{ "type": "text", "value": "hello", "locale": "eng" }] }
    ]
  }
  ```
  `serial` and a non-empty `steps` array are required; every entry must be a
  valid [`Step`](#step).
- **Response `200`:** `{ "status": "ok" }` — the steps were queued. Track
  execution via [`GET /devices/{serial}/session`](#get-devicesserialsession).
- **Errors:** `400` on invalid JSON or an invalid step, `500` when a referenced
  library image/action does not exist or the session could not be started.

## Scripts

A **script** is a cached `/run_steps` body: a named, reusable step list stored
as `scripts/<name>/run.json` (a JSON array of [`Step`](#step)s). Scripts are
pure caching — saving one never executes anything, and running one queues its
steps exactly as if they had been sent to `/run_steps`. Saving under an
existing name overwrites the whole script.

### `GET /scripts`

Lists all saved script names (the folder names under the scripts dir).

- **Response `200`:**
  ```json
  { "scripts": ["open_main_page", "x5_add_corn_to_cart"] }
  ```
  `scripts` is `[]` when none are saved.

### `GET /scripts/{name}`

Returns the script's steps (the content of its `run.json`).

- **Path params:** `name` (required).
- **Response `200`:** a JSON array of [`Step`](#step)s.
  ```json
  [
    { "event": "tap", "anchors": [{ "type": "text", "value": "Pyaterochka", "locale": "eng" }] },
    { "event": "swipe_x5_catalog_1" }
  ]
  ```
- **Errors:** `404` when the script does not exist.

### `POST /scripts`

Saves a script. The body is the same as `/run_steps` with `name` instead of
`serial`; steps are validated the same way (referenced library images and
events must exist). Saving under an existing name **rewrites the whole
script**.

- **Request body:**
  ```json
  {
    "name": "open_main_page",
    "steps": [
      { "event": "tap", "anchors": [{ "type": "text", "value": "Pyaterochka", "locale": "eng" }] },
      { "event": "tap", "anchors": [{ "type": "yolo", "value": "home" }] }
    ]
  }
  ```
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON, a missing/invalid `name`, or invalid
  steps; `500` when a referenced library image/action does not exist or the
  file could not be written.

### `DELETE /scripts/{name}`

Deletes the script (its whole folder).

- **Path params:** `name` (required).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `404` when the script does not exist.

### `GET /run_script`

Loads a saved script and queues its steps, exactly like `/run_steps` (steps
re-validated, session opened automatically when none exists).

- **Query params:** `serial`, `name` (both required).
- **Response `200`:** `{ "status": "ok" }` — the script's steps were queued.
  Track execution via [`GET /devices/{serial}/session`](#get-devicesserialsession).
- **Errors:** `400` on missing params, `500` when the script does not exist,
  a referenced library asset is missing, or the session could not be started.

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
  "anchors": [
    { "type": "text", "value": "Show refresh rate", "locale": "eng" },
    { "type": "image", "value": "toggle" }
  ],
  "timeout": 15
}
```

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Event | `event` | string | `tap`, `long_tap`, `type_text`, the name of a library event, or **empty for a visibility check** |
| Anchors | `anchors` | []Anchor | omitempty; the target chain — resolved in order on one frame, each anchor found nearest to the previous one, the event applies to the **last** one. Empty only for a target-less library event replay. |
| Timeout | `timeout` | int | omitempty; seconds to locate the anchors before failing (default `15` when omitted or `<= 0`). The runner grabs the latest video frame and retries roughly once per second until the deadline. |

### Anchor

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Type | `type` | string | `image` (template match against `images/<value>.png`), `text` (OCR), or `yolo` (detection class) |
| Value | `value` | string | Library image name, OCR text, or YOLO class name — for `type_text`'s last anchor, the text to type |
| Locale | `locale` | string | omitempty; OCR language for `text` anchors, keyboard locale for `type_text` |

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
