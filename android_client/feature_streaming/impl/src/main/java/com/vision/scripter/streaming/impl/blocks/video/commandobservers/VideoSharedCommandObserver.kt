package com.vision.scripter.streaming.impl.blocks.video.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import com.vision.scripter.streaming.impl.screen.commandobservers.MenuToVideo
import com.vision.scripter.streaming.impl.screen.commandobservers.ScreenToVideo
import com.vision.scripter.streaming.impl.screen.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
internal fun VideoSharedCommandObserver(
    videoUiStateHolder: VideoUiStateHolder,
    streamingUiStateHolder: StreamingUiStateHolder,
) {
    videoUiStateHolder.menuCommandsFlow.observe {
        streamingUiStateHolder.sharedCommandsFlow.tryEmit(it)
    }

    videoUiStateHolder.screenCommandsFlow.observe {
        streamingUiStateHolder.sharedCommandsFlow.tryEmit(it)
    }

    streamingUiStateHolder.sharedCommandsFlow.observe {
        when (it) {
            is ScreenToVideo -> videoUiStateHolder.onSharedEvent(it)
            is MenuToVideo -> videoUiStateHolder.onSharedEvent(it)
            else -> Unit
        }
    }
}
