package com.vision.scripter.streaming.impl.screen.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.screen.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
internal fun StreamingSharedEventObserver(
    uiStateHolder: StreamingUiStateHolder,
) {
    uiStateHolder.sharedCommandsFlow.observe {
        if (it is VideoToScreen) {
            uiStateHolder.onSharedEvent(it)
        }
    }
}