package com.vision.scripter.devices.ui

import androidx.compose.runtime.Immutable
import androidx.compose.ui.graphics.ImageBitmap
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.ImmutableMap
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.persistentMapOf

@Immutable
data class DevicesUiState(
    val isLoading: Boolean = true,
    val isRefreshing: Boolean = false,
    val devices: ImmutableList<UiDevice> = persistentListOf(),
)

@Immutable
data class UiDevice(
    val serial: String = "",
    val deviceParams: ImmutableMap<Int, String> = persistentMapOf(),
    val preview: ImageBitmap? = null,
)