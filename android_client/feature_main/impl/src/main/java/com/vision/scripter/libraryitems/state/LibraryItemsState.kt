package com.vision.scripter.libraryitems.state

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.libraryitems.data.Run
import com.vision.scripter.ui.states.LoadingState

data class LibraryItemsState(
    val type: LibraryType = LibraryType.IMAGES,
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val names: List<String> = listOf(),
    val itemToDelete: String? = null,
    val runs: Map<String, Run> = mapOf(),
    val picker: DevicePicker? = null,
)

data class DevicePicker(
    val itemName: String,
    val isLoading: Boolean = true,
    val devices: List<PickerDevice> = listOf(),
    val lastSerial: String = "",
)

data class PickerDevice(
    val device: AdbDevice,
    val status: SessionStatus,
)
