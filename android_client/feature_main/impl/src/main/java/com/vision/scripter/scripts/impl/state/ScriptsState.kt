package com.vision.scripter.scripts.impl.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.ui.states.LoadingState

data class ScriptsState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val nodes: List<String> = listOf(),
    val selectedNode: String = "",
    val scripts: List<String> = listOf(),
    val deleteTarget: String = "",
    val deleteIsNode: Boolean = false,
    val scriptToRun: String = "",
    val devices: List<AdbDevice> = listOf(),
    val isDevicesLoading: Boolean = false,
    val selectedSerial: String = "",
)
