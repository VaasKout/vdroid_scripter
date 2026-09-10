package com.vision.scripter.libraryitems.state

import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.libraryitems.data.Run
import com.vision.scripter.libraryitems.data.label
import com.vision.scripter.libraryitems.ui.LibraryItemsUiState
import com.vision.scripter.libraryitems.ui.UiDevicePicker
import com.vision.scripter.libraryitems.ui.UiLibraryItem
import com.vision.scripter.libraryitems.ui.UiPickerDevice
import com.vision.scripter.libraryitems.ui.UiRun
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toImmutableList
import javax.inject.Inject

@ViewModelScoped
class LibraryItemsUiStateMapper @Inject constructor() {
    fun map(state: LibraryItemsState): LibraryItemsUiState {
        val canPlay = state.type != LibraryType.IMAGES
        val runsByName = state.runs.values
            .filter { it.type == state.type }
            .associateBy { it.name }
        return LibraryItemsUiState(
            type = state.type,
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            isRefreshing = state.loadingState == LoadingState.RefreshLoading,
            items = state.names.map { name ->
                UiLibraryItem(
                    name = name,
                    canPlay = canPlay,
                    run = runsByName[name]?.toUi(),
                )
            }.toImmutableList(),
            itemToDelete = state.itemToDelete,
            picker = state.picker?.toUi(),
        )
    }

    private fun Run.toUi(): UiRun = UiRun(
        statusText = status.text,
        deviceLabel = deviceLabel,
        isError = status is SessionStatus.Error,
    )

    private fun DevicePicker.toUi(): UiDevicePicker {
        val uiDevices = devices.map {
            UiPickerDevice(
                serial = it.device.serial,
                label = it.device.label(),
                busy = it.busy,
                statusText = it.status.text,
                selected = it.device.serial == selectedSerial,
            )
        }.toImmutableList()
        return UiDevicePicker(
            isLoading = isLoading,
            devices = uiDevices,
            canPlay = uiDevices.any { it.selected && !it.busy },
        )
    }
}
