package com.vision.scripter.streaming.impl.blocks.menu.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.commandobservers.VideoToMenu
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
internal fun MenuSharedCommandObserver(
    menuUiStateHolder: MenuUiStateHolder,
    streamingUiStateHolder: StreamingUiStateHolder,
) {
    menuUiStateHolder.videoCommandsFlow.observe {
        streamingUiStateHolder.sharedCommandsFlow.tryEmit(it)
    }

    streamingUiStateHolder.sharedCommandsFlow.observe {
        when (it) {
            is VideoToMenu -> menuUiStateHolder.onSharedEvent(it)
            else -> Unit
        }
    }
}
