package com.vision.scripter.scripts.impl.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toImmutableList
import javax.inject.Inject

@ViewModelScoped
class ScriptsUiStateMapper @Inject constructor() {
    fun map(state: ScriptsState): ScriptsUiState {
        return ScriptsUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            isRefreshing = state.loadingState == LoadingState.RefreshLoading,
            showScripts = state.selectedNode.isNotEmpty(),
            selectedNode = state.selectedNode,
            nodes = state.nodes.toImmutableList(),
            scripts = state.scripts.toImmutableList(),
            deleteDialog = DeleteDialogData(
                show = state.deleteTarget.isNotEmpty(),
                name = state.deleteTarget,
                isNode = state.deleteIsNode,
            ),
            devicePickerData = DevicePickerBottomSheetData(
                showDevicePicker = state.scriptToRun.isNotEmpty(),
                isDevicesLoading = state.isDevicesLoading,
                devices = state.devices.map {
                    UiScriptDevice(
                        serial = it.serial,
                        title = it.pickerTitle(),
                    )
                }.toImmutableList(),
                selectedSerial = state.selectedSerial,
            ),
        )
    }
}

private fun AdbDevice.pickerTitle(): String {
    val name = listOf(brand, model).filter { it.isNotBlank() }.joinToString(" ")
    if (name.isBlank()) return serial
    return "$name ($serial)"
}
