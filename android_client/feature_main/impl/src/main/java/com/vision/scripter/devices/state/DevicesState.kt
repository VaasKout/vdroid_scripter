package com.vision.scripter.devices.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.ui.states.LoadingState

data class DevicesState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val devices: List<AdbDevice> = listOf()
)
