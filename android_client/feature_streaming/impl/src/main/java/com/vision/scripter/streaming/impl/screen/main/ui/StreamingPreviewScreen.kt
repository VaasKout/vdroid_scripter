package com.vision.scripter.streaming.impl.screen.main.ui

import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

internal val streamingPreviewState = StreamingUiState()

internal class StreamingPreviewScreenUiStateHolder(state: StreamingUiState) :
    StreamingUiStateHolder {
    override val uiStateFlow: StateFlow<StreamingUiState?>
        get() = throw UnsupportedOperationException()
    override val uiCommandsFlow: CommandFlow<StreamingUiCommand>
        get() = throw UnsupportedOperationException()

    override fun onLoadData(serial: String) = Unit
}
