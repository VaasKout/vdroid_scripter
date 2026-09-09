### Install Android Client

You can download the prebuilt Android application with Android Studio or download it from the releases page:

👉 [Link](https://github.com/VaasKout/android_vision_scripter/releases)

Then install it on your device:

- Enable installation from unknown sources (if required)
- Open the downloaded APK file
- Install and launch the app

## Connecting the Client to the Server

The Android client connects to the server using an HTTP address provided by the user.

In the Android application, enter the following:

- **Server IP address** (local network IP of the machine running the server)
- **Port number** (default: `8080`)

### Finding the Server IP Address

The server must be running on a machine within the same local network as the Android device.

#### Linux
Use the following command:

```bash
ip a
```
Look for your active network interface (e.g., wlan0 or eth0) and find the inet field:

```bash
inet 192.168.x.x/24
```

The IP address (e.g., 192.168.x.x) is what you should use.
#### macOS

Use:

```bash
ifconfig
```
Look for your active interface (usually en0 for Wi-Fi) and find:

```bash
inet 192.168.x.x
```

That value is your local server IP.

![Example](app_example.gif)

## Using the app

- **Devices tab** — every device the server sees over ADB, with a Streaming
  button that opens the live screen. Streaming a device restarts its session
  and cancels whatever route is running on it — this is the emergency stop.
- **Library tab** — three cards: Images, Actions and Routes, each with a
  count. A card opens its list. Every row can be deleted; action and route
  rows have a play button.
- **Play** — pick a device in the bottom sheet (the last one used is
  preselected). A device that is already running something is grayed out
  with its status. An action runs as a single verbatim gesture; a route runs
  from its first step. The row then shows the live status and the device;
  on an error the server's text stays on the row until you tap it.
- **Streaming screen** — the plus icon crops a template image or records a
  gesture, the eye shows the rectangles the server detects (fetched once per
  tap), the magnifying glass scans the screen for text, YOLO classes and,
  optionally, your library images, the keyboard icon edits key zones.

## Device Compatibility Notice

⚠️ Not all Android devices are supported for wireless operation.

Some devices only work in **USB (OTG / ADB over USB) mode** with `scrcpy`. In this mode, **video streaming is not available** for some devices, which means this application will **not function properly**, since it relies on screen capture for computer vision processing.

Only devices that support stable **screen streaming** are fully compatible with this project.

Compatibility may depend on:
- Android version
- OEM restrictions (Samsung, Xiaomi, etc.)
- USB-only debugging limitations
