# VDroid Scripter

**VDroid Scripter** stands for "Visual Android Scripter".

This is a tool for remotely controlling Android devices and creating automation scripts using computer vision without requiring users to manually write code or commands.

The project consists of two main components:

* **Server**
  A Golang-based microservice responsible for receiving video streams and sending control inputs. It leverages [scrcpy-server](https://github.com/Genymobile/scrcpy/tree/master/server) for efficient screen streaming and interaction.

* **Client**
  An example Android application that connects to the server. The client can be **any application capable of decoding H.264 streams and handling input events (e.g., clicks, gestures)**.


## Getting Started
* [server](docs/server.md)
* [android-client](docs/android_client.md)

## Project Status

This project is currently in **active development** and may contain bugs.
If you have any problems, feel free to open an [issue](https://github.com/VaasKout/android_vision_scripter/issues)

## Acknowledgments

This project builds upon and would not be possible without the following open-source projects:

* [scrcpy](https://github.com/Genymobile/scrcpy) — for Android screen streaming and control
* [go-astiav](https://github.com/asticode/go-astiav) — for FFmpeg bindings in Go
* [gocv](https://github.com/hybridgroup/gocv) — for OpenCV integration in Go

Huge thanks to the authors and contributors of these projects.

## Contact

If you have any questions, ideas, or would like to contribute to development, feel free to reach out:

**Email:** *VasyaKotov1@gmail.com*  
