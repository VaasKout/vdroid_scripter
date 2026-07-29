package com.vision.scripter.streaming.impl.screen.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.streaming.impl.screen.commandobservers.StreamingSharedEvent
import com.vision.scripter.streaming.impl.screen.commandobservers.VideoToScreen
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.CommandFlow2
import kotlinx.coroutines.flow.StateFlow

@Stable
interface StreamingUiStateHolder {
    val uiStateFlow: StateFlow<StreamingUiState>
    val uiCommandsFlow: CommandFlow<StreamingUiCommand>
    val sharedCommandsFlow: CommandFlow2<StreamingSharedEvent>

    fun onSharedEvent(event: VideoToScreen)
    fun onRefresh()
}
