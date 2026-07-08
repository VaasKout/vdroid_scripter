package com.vision.scripter.streaming.impl.screen.main.ui

import android.view.MotionEvent
import android.view.Surface
import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface StreamingUiStateHolder {
    val uiStateFlow: StateFlow<StreamingUiState?>
    val uiCommandsFlow: CommandFlow<StreamingUiCommand>

    fun initArgs(serial: String)
    fun onLoadData(onStart: Boolean)
    fun onVideoSurfaceCreated(surfaceWidth: Int, surfaceHeight: Int, newSurface: Surface)
    fun onVideoSurfaceDestroyed()
    fun onTouchEvent(viewWidth: Int, viewHeight: Int, event: MotionEvent?)
}
