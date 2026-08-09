package com.vision.scripter.scripts.impl.state

import androidx.compose.runtime.Immutable
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class ScriptsUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val selectedLocation: String = "",
    val locations: ImmutableList<String> = persistentListOf(),
    val scripts: ImmutableList<String> = persistentListOf(),
    val itemToDelete: String = "",
    val bottomSheetUiData: BottomSheetUiData? = null,
)

@Immutable
data class BottomSheetUiData(
    val isLoading: Boolean = false,
    val devices: ImmutableList<DeviceToPick> = persistentListOf(),
)

@Immutable
data class DeviceToPick(
    val serial: String = "",
    val title: String = "",
    val selected: Boolean = false,
)
