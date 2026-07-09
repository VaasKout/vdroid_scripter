package com.vision.scripter.scripts.impl.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.ui.states.LoadingState

data class ScriptsState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val scripts: List<String> = listOf(),
    val scriptToDelete: String = "",
    val scriptToRun: String = "",
    val devices: List<AdbDevice> = listOf(),
    val isDevicesLoading: Boolean = false,
    val selectedSerial: String = "",
)
