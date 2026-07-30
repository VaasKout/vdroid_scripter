package com.vision.scripter.streaming.impl.screen.ui

import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val streamingPreviewState = StreamingUiState()

internal class StreamingScreenUiStateHolderPreview(state: StreamingUiState) :
    StreamingUiStateHolder {
    override val uiStateFlow: StateFlow<StreamingUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<StreamingUiCommand>
        get() = throw UnsupportedOperationException()

    override fun onRefresh() = Unit
}
