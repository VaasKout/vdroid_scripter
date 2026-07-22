package com.vision.scripter.streaming.impl.blocks.video.ui

import android.view.MotionEvent
import android.view.Surface
import androidx.compose.runtime.Stable
import kotlinx.coroutines.flow.StateFlow

@Stable
interface VideoUiStateHolder {

    val uiStateFlow: StateFlow<VideoUiState>

    fun onVideoSurfaceCreated(surfaceWidth: Int, surfaceHeight: Int, newSurface: Surface)
    fun onVideoSurfaceDestroyed()
    fun onTouchEvent(viewWidth: Int, viewHeight: Int, event: MotionEvent?)
}
