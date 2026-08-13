package com.vision.scripter.devices.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.SharedFlow

@Stable
interface DevicesUiStateHolder {
    val uiStateFlow: SharedFlow<DevicesUiState>
    val uiCommandsFlow: CommandFlow<DevicesUiCommand>

    fun onLoadData(onStart: Boolean)
}