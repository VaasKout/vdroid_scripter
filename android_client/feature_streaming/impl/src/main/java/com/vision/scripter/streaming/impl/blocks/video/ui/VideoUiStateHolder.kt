package com.vision.scripter.streaming.impl.blocks.video.ui

import android.view.MotionEvent
import android.view.Surface
import androidx.compose.runtime.Stable
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface VideoUiStateHolder {

    val uiCommandsFlow: CommandFlow<VideoUiCommand>
    val uiStateFlow: StateFlow<VideoUiState>

    fun initArgs(serial: String)
    fun onLoadData(onStart: Boolean)
    fun onVideoSurfaceCreated(surfaceWidth: Int, surfaceHeight: Int, newSurface: Surface)
    fun onVideoSurfaceDestroyed()
    fun onTouchEvent(viewWidth: Int, viewHeight: Int, event: MotionEvent?)

    fun nextCvMode(cvMode: CVMode)
    fun saveClicked(parameter: Parameter?)
    fun recordCancelled()
    fun findTextClicked(text: String, locale: String)
    fun initKeyboardClicked()
    fun editKeyboardKeyClicked(key: String)
    fun saveKeyboardLocaleClicked(locale: String)
    fun saveTimeoutClicked(timeout: Int)
    fun saveRecordNameClicked(name: String)
}
