### Dependencies

To run the server, you must install the following dependencies:

* **Go**
* **ADB**
* **FFmpeg**
* **OpenCV**
* **Tesseract OCR**

### Installation

#### Arch Linux

```bash
sudo pacman -S go android-tools ffmpeg opencv tesseract $(pacman -Sl extra | grep tesseract-data | awk '{print $2}')
```

#### macOS

```bash
brew install go android-platform-tools ffmpeg opencv tesseract tesseract-lang
sudo curl -L https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata -o /opt/homebrew/share/tessdata/eng.traineddata # Legacy tesseract support
```

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

Navigate to the server directory and run:

```bash
go mod tidy
go run cmd/main.go
```

### Network & Security Notice

⚠️ **Important:** The communication between the server and client is **not encrypted (no TLS or encryption layer)**.

This project is intended to be used in a **trusted local network environment only**, such as:

- Personal home networks
- Development environments
- Isolated testing setups

It is **not recommended** to expose the server to the public internet or untrusted networks, as all transmitted data (including screen frames and control commands) can be intercepted.

For production or internet-facing usage, additional security layers (such as VPN, SSH tunneling, or TLS encryption) must be implemented manually.
