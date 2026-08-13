package com.vision.scripter.devices.ui

sealed class DevicesUiCommand {
    data object ShowNetworkError : DevicesUiCommand()
    data class NavigateToStreaming(val serial: String) : DevicesUiCommand()
}