package com.vision.scripter.scripts.impl.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.scripts.impl.state.ScriptsState.BottomSheetData
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
            selectedNode = state.selectedNode,
            nodes = state.nodes.toImmutableList(),
            scripts = state.scripts.toImmutableList(),
            itemToDelete = state.itemToDelete,
            bottomSheetUiData = state.bottomSheetData?.toDevicePickerData(),
        )
    }
}

private fun BottomSheetData.toDevicePickerData(): BottomSheetUiData? {
    if (selectedScript.isEmpty()) return null
    return BottomSheetUiData(
        isLoading = isLoading,
        devices = devices.map {
            DeviceToPick(
                serial = it.serial,
                title = it.bottomSheetTitle(),
                selected = it.serial == selectedDevice,
            )
        }.toImmutableList(),
    )
}

private fun AdbDevice.bottomSheetTitle(): String {
    val name = listOf(brand, model).filter { it.isNotBlank() }.joinToString(" ")
    if (name.isBlank()) return serial
    return "$name ($serial)"
}
