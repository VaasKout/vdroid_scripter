package com.vision.scripter.streaming.impl.blocks.video.ui

import android.view.MotionEvent
import android.view.Surface
import com.vision.scripter.streaming.impl.screen.main.commandobservers.MenuToVideo
import com.vision.scripter.streaming.impl.screen.main.commandobservers.ScreenToVideo
import com.vision.scripter.streaming.impl.screen.main.commandobservers.VideoToMenu
import com.vision.scripter.streaming.impl.screen.main.commandobservers.VideoToScreen
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val videoUiStatePreview = VideoUiState()

internal class VideoBlockPreviewUiStateHolder(state: VideoUiState) :
    VideoUiStateHolder {
    override val uiStateFlow: StateFlow<VideoUiState> = MutableStateFlow(state)
    override val menuCommandsFlow: CommandFlow<VideoToMenu>
        get() = throw UnsupportedOperationException()
    override val screenCommandsFlow: CommandFlow<VideoToScreen>
        get() = throw UnsupportedOperationException()

    override fun init(serial: String) = Unit
    override fun onSharedEvent(event: ScreenToVideo) = Unit
    override fun onSharedEvent(event: MenuToVideo) = Unit
    override fun onVideoSurfaceCreated(
        surfaceWidth: Int,
        surfaceHeight: Int,
        newSurface: Surface
    ) = Unit

    override fun onVideoSurfaceDestroyed() = Unit
    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?
    ) = Unit
}
