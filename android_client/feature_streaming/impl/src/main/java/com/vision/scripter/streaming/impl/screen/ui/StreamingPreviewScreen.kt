package com.vision.scripter.streaming.impl.screen.ui

import com.vision.scripter.streaming.impl.screen.commandobservers.StreamingSharedEvent
import com.vision.scripter.streaming.impl.screen.commandobservers.VideoToScreen
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.CommandFlow2
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val streamingPreviewState = StreamingUiState()

internal class StreamingScreenUiStateHolderPreview(state: StreamingUiState) :
    StreamingUiStateHolder {
    override val uiStateFlow: StateFlow<StreamingUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<StreamingUiCommand>
        get() = throw UnsupportedOperationException()
    override val sharedCommandsFlow: CommandFlow2<StreamingSharedEvent>
        get() = throw UnsupportedOperationException()

    override fun onSharedEvent(event: VideoToScreen) = Unit
    override fun onRefresh() = Unit
}
