package com.vision.scripter.streaming.impl.screen.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface StreamingUiStateHolder {
    val uiStateFlow: StateFlow<StreamingUiState>
    val uiCommandsFlow: CommandFlow<StreamingUiCommand>

    fun onRefresh()
}
