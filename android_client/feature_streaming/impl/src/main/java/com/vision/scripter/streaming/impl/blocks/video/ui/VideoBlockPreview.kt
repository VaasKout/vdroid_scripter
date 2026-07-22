package com.vision.scripter.streaming.impl.blocks.video.ui

import android.view.MotionEvent
import android.view.Surface
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val videoUiStatePreview = VideoUiState()

internal class VideoBlockPreviewUiStateHolder(state: VideoUiState) :
    VideoUiStateHolder {
    override val uiStateFlow: StateFlow<VideoUiState> = MutableStateFlow(state)

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
