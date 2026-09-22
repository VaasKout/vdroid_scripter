package com.vision.scripter.data.api.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class AdbDevicesResponse(
    @SerialName("devices")
    val devices: List<AdbDevice>,
)

@Serializable
data class AdbDevice(
    @SerialName("serial")
    val serial: String = "",
    @SerialName("brand")
    val brand: String = "",
    @SerialName("device")
    val device: String = "",
    @SerialName("locale")
    val locale: String = "",
    @SerialName("model")
    val model: String = "",
    @SerialName("os_version")
    val osVersion: String = "",
    @SerialName("manufacturer")
    val manufacturer: String = "",
    @SerialName("marketing_name")
    val marketingName: String = "",
    @SerialName("scrcpy_running")
    val scrCpyConnection: Boolean = false,
)

@Serializable
data class StreamingData(
    @SerialName("video_port")
    val videoPort: String = "",
    @SerialName("control_port")
    val controlPort: String = "",
)

data class ScreenSizes(
    val surfaceWidth: Int = 0,
    val surfaceHeight: Int = 0,
    val remoteWidth: Int = 0,
    val remoteHeight: Int = 0,
)