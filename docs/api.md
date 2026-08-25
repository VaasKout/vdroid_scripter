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
| GET | `/devices/{serial}/find_text` | OCR: locate text on the current screen |
| GET | `/library` | List image and action names in the library |
| POST | `/save_image` | Crop and save a named template image into the library |
| POST | `/save_action` | Save a named recorded gesture into the library |
| DELETE | `/images/{name}` | Delete a library image |
| DELETE | `/actions/{name}` | Delete a library action |
| GET | `/map` | List map node names |
| GET | `/map/{name}` | Get one map node |
| POST | `/map` | Create or merge-update a map node |
| DELETE | `/map/{name}` | Delete a map node |
| POST | `/follow_route` | Navigate from one node to another via BFS over edges |
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
CV-located **landmarks**. Steps compose the [library](#library) at runtime —
each landmark (`type` + `value`) locates a region on the live screen (a library
image, OCR text, or a YOLO class) and the event acts on the **last** landmark's
region (generated taps, CV keyboard typing, or a recorded gesture from the
library).

The chain resolves on a single video frame: the first landmark picks its best
match on screen, every following landmark picks the candidate of its value
**nearest** (by center distance) to the previous landmark, and the event applies
to the last landmark. One landmark is the normal case; a preceding unique
landmark disambiguates duplicates — "the toggle next to *Show refresh rate*"
is `landmarks: [{text "Show refresh rate"}, {image "toggle"}]`. Candidate
lists are ordered top-to-bottom, left-to-right (reading order), so when a
value matches several places and no disambiguating landmark precedes it, a
bare landmark deterministically takes the first candidate on screen.

| `event` | Behavior |
| ------- | -------- |
| *(empty)* | Visibility check of the landmark chain — no touch. Landmarks required. |
| `tap` / `long_tap` | Generated tap pair placed at a random point inside the last landmark's region. Landmarks required. |
| `type_text` | The **last landmark's `value` is the text to type**, typed via the CV keyboard (its `locale` = keyboard locale). Nothing is located on screen. |
| any other name | The library event with that name is replayed: **offset into the found region** when landmarks are given (first touch moved into the last landmark's region, relative shape preserved), **verbatim** without them. |

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
      { "landmarks": [{ "type": "yolo", "value": "home" }] },
      { "event": "tap", "landmarks": [{ "type": "image", "value": "catalog_cart_icon" }] },
      { "event": "swipe_catalog_1" },
      { "event": "tap", "landmarks": [{ "type": "text", "value": "Corn", "locale": "eng" }], "timeout": 10 },
      { "event": "tap", "landmarks": [{ "type": "text", "value": "Show refresh rate" }, { "type": "image", "value": "toggle" }] },
      { "event": "type_text", "landmarks": [{ "type": "text", "value": "hello", "locale": "eng" }] }
    ]
  }
  ```
  `serial` and a non-empty `steps` array are required; every entry must be a
  valid [`Step`](#step).
- **Response `200`:** `{ "status": "ok" }` — the steps were queued. Track
  execution via [`GET /devices/{serial}/session`](#get-devicesserialsession).
- **Errors:** `400` on invalid JSON or an invalid step, `500` when a referenced
  library image/action does not exist or the session could not be started.

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

## Map

The **map** is a graph of curated screens stored as `map/<name>/node.json`.
A **node** describes one screen: its identifying `landmarks`, its `edges`
(actions with the nodes they can lead to), and an `occupancy_grid` — a
content-blind structural fingerprint the server captures on its own during
route execution. The map is purely declarative: creating and updating nodes
never touches a device, and `POST /run_steps` never reads or writes the map.

### `GET /map`

Lists all node names.

- **Response `200`:** `{ "nodes": ["main_screen", "x5_catalog"] }` — sorted,
  `[]` when the map is empty.

### `GET /map/{name}`

Returns one node.

- **Response `200`:** the [`Node`](#node) JSON.
- **Errors:** `404` if no node with that name exists.

### `POST /map`

Creates a node, or **merge-updates** an existing one (matched by `name`).

- **Request body:** a [`Node`](#node):
  ```json
  {
    "name": "main_screen",
    "landmarks": [{ "type": "image", "value": "catalog_icon" }],
    "edges": [
      {
        "action": { "event": "tap", "landmarks": [{ "type": "text", "value": "catalog" }] },
        "next_nodes": ["x5_catalog"]
      },
      { "action": { "event": "buy_button_tap" }, "next_nodes": [] }
    ],
    "occupancy_grid": { "cols": 18, "rows": 32, "cells": ["001100...", "..."] }
  }
  ```
  `name` is required (library naming rules). An edge's `action` is a full
  [`Step`](#step); `next_nodes` may be **empty** (a terminal action on a
  final page) and may reference nodes that don't exist yet (dangling names
  are legal and simply unroutable until created). `occupancy_grid` is
  optional.
- **Merge semantics** when the name exists: landmarks are appended with
  dedup by `type`+`value`+`locale`; an edge whose `action` equals an existing
  edge's action (same `event` and landmark sequence, `timeout` ignored)
  unions its `next_nodes`, otherwise the edge is appended; a **non-empty**
  `occupancy_grid` in the body replaces the stored one, an absent or empty
  grid keeps it. Nothing is ever auto-deleted.
- **Validation:** image landmarks and library events referenced by edge
  actions must exist (same rule as `/run_steps`).
- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `400` on invalid JSON, an invalid node, or a missing asset.

### `DELETE /map/{name}`

Deletes the node's folder. References to the deleted name in other nodes'
`next_nodes` become dangling.

- **Response `200`:** `{ "status": "ok" }`
- **Errors:** `404` if no node with that name exists.

### `POST /follow_route`

Navigates the device from one node to another. The server plans the shortest
path over edges (unweighted BFS, edges in stored order → deterministic),
queues one **route job** through the session queue (a session opens
automatically, like `/run_steps`), and executes it hop by hop: verify the
current node, run the edge's action, verify the expected next node.

Verification per node: all `landmarks` found on one video frame; on landmark
success with an unset grid the server captures the occupancy grid from that
same frame and saves it into `node.json`; when landmarks fail, the grid is
the fallback (Jaccard similarity over occupied cells, match at `>= 0.8`);
a node with neither landmarks nor grid is executed on trust. If an edge
declared several `next_nodes` and the expected one fails, the other declared
outcomes are checked (short probe) and the route re-plans from wherever it
actually landed (bounded — persistent no-progress ends the route). When
nothing confirms, the session status becomes
`lost on route <from> -> <to>: node <name> not confirmed` and the remaining
queue is cleared.

- **Request body:** `{ "serial": "ABCD1234", "from": "main_screen", "to": "x5_catalog" }`
  — all required; both nodes must exist; `from == to` is legal and means
  "verify I'm on this node".
- **Response `200`:** `{ "status": "ok" }` — the route was queued. Track the
  outcome via [`GET /devices/{serial}/session`](#get-devicesserialsession):
  `idle` = arrived, `lost ...` = localization lost.
- **Errors:** `400` on invalid JSON or a missing field, `500` when a node
  doesn't exist, no route exists, or the session couldn't be started.

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
  - `running route <from> -> <to>: <phase>` — a queued route is executing;
    the phase names the current hop or verification;
  - `lost on route <from> -> <to>: node <name> not confirmed` — a route lost
    localization at that node (neither landmarks nor occupancy grid matched);
    the queue was cleared;
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
  "landmarks": [
    { "type": "text", "value": "Show refresh rate", "locale": "eng" },
    { "type": "image", "value": "toggle" }
  ],
  "timeout": 15
}
```

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Event | `event` | string | `tap`, `long_tap`, `type_text`, the name of a library event, or **empty for a visibility check** |
| Landmarks | `landmarks` | []Landmark | omitempty; the target chain — resolved in order on one frame, each landmark found nearest to the previous one, the event applies to the **last** one. Empty only for a target-less library event replay. |
| Timeout | `timeout` | int | omitempty; seconds to locate the landmarks before failing (default `15` when omitted or `<= 0`). The runner grabs the latest video frame and retries roughly once per second until the deadline. |

### Landmark

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Type | `type` | string | `image` (template match against `images/<value>.png`), `text` (OCR), or `yolo` (detection class) |
| Value | `value` | string | Library image name, OCR text, or YOLO class name — for `type_text`'s last landmark, the text to type |
| Locale | `locale` | string | omitempty; OCR language for `text` landmarks, keyboard locale for `type_text` |

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

### Node

`server/pkg/models/node.go`

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Name | `name` | string | Unique node name (library naming rules) |
| Landmarks | `landmarks` | [][Landmark](#landmark) | The screen's identity check — **all** must be found on one frame for a landmark verification to pass; may be empty |
| Edges | `edges` | [][Edge](#edge) | Actions available on this screen |
| OccupancyGrid | `occupancy_grid` | [OccupancyGrid](#occupancygrid) | omitempty; unset until captured or posted |

### Edge

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Action | `action` | [Step](#step) | A full step (generated tap, `type_text`, or library event); check events (empty `event`) are not valid edge actions |
| NextNodes | `next_nodes` | []string | Nodes this action can lead to; empty = terminal action; names may dangle |

### OccupancyGrid

A fixed 18×32 grid normalized over the video frame; each cell is `'1'` when a
structural rectangle covers it. Compared with Jaccard similarity
(intersection ÷ union of occupied cells), match threshold `0.8`. A grid with
no `'1'` cells counts as unset.

| Field | JSON | Type | Notes |
| ----- | ---- | ---- | ----- |
| Cols | `cols` | int | 18 |
| Rows | `rows` | int | 32 |
| Cells | `cells` | []string | One string per row, `'0'`/`'1'` per cell, each exactly `cols` long |

### OCRResult

`server/internal/cv/model.go`

```json
{ "text": "Login", "rectangle": { "left_x": 40, "right_x": 220, "top_y": 900, "bottom_y": 980 } }
```

| Field | JSON | Type |
| ----- | ---- | ---- |
| Text | `text` | string |
| Rectangle | `rectangle` | [Rectangle](#rectangle) |
