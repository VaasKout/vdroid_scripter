package com.vision.scripter.scripts.impl.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.ui.states.LoadingState

data class ScriptsState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val nodes: List<String> = listOf(),
    val selectedNode: String = "",
    val scripts: List<String> = listOf(),
    val itemToDelete: String = "",
    val bottomSheetData: BottomSheetData? = null,
) {
    data class BottomSheetData(
        val isLoading: Boolean = true,
        val selectedScript: String = "",
        val selectedDevice: String = "",
        val devices: List<AdbDevice> = listOf(),
    )
}
