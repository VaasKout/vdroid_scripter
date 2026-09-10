package com.vision.scripter.libraryitems.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.library.state.LibraryType
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class LibraryItemsUiState(
    val type: LibraryType = LibraryType.IMAGES,
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val items: ImmutableList<UiLibraryItem> = persistentListOf(),
    val itemToDelete: String? = null,
    val picker: UiDevicePicker? = null,
)

@Immutable
data class UiLibraryItem(
    val name: String,
    val canPlay: Boolean,
    val run: UiRun? = null,
)

@Immutable
data class UiRun(
    val statusText: String,
    val deviceLabel: String,
    val isError: Boolean,
)

@Immutable
data class UiDevicePicker(
    val isLoading: Boolean,
    val devices: ImmutableList<UiPickerDevice>,
    val canPlay: Boolean,
)

@Immutable
data class UiPickerDevice(
    val serial: String,
    val label: String,
    val busy: Boolean,
    val statusText: String,
    val selected: Boolean,
)
