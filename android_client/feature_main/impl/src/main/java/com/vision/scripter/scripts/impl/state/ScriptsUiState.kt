package com.vision.scripter.scripts.impl.state

import androidx.compose.runtime.Immutable
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class ScriptsUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val showScripts: Boolean = false,
    val selectedNode: String = "",
    val nodes: ImmutableList<String> = persistentListOf(),
    val scripts: ImmutableList<String> = persistentListOf(),
    val deleteDialog: DeleteDialogData = DeleteDialogData(),
    val devicePickerData: DevicePickerBottomSheetData = DevicePickerBottomSheetData(),
)

@Immutable
data class DeleteDialogData(
    val show: Boolean = false,
    val name: String = "",
    val isNode: Boolean = false,
)

@Immutable
data class DevicePickerBottomSheetData(
    val showDevicePicker: Boolean = false,
    val isDevicesLoading: Boolean = false,
    val devices: ImmutableList<UiScriptDevice> = persistentListOf(),
    val selectedSerial: String = "",
)

@Immutable
data class UiScriptDevice(
    val serial: String = "",
    val title: String = "",
)
