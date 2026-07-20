package com.vision.scripter.streaming.impl.blocks.video.ui

import android.view.MotionEvent
import android.view.Surface
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val videoUiStatePreview = VideoUiState()

internal class VideoBlockPreviewUiStateHolder(state: VideoUiState) :
    VideoUiStateHolder {
    override val uiStateFlow: StateFlow<VideoUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<VideoUiCommand>
        get() = throw UnsupportedOperationException()


    override fun initArgs(serial: String) {}
    override fun onLoadData(onStart: Boolean) {}

    override fun onVideoSurfaceCreated(
        surfaceWidth: Int,
        surfaceHeight: Int,
        newSurface: Surface
    ) {
    }

    override fun onVideoSurfaceDestroyed() {}

    override fun onTouchEvent(
        viewWidth: Int,
        viewHeight: Int,
        event: MotionEvent?
    ) {
    }

    override fun nextCvMode(cvMode: CVMode) {}
    override fun saveClicked(parameter: Parameter?) {}
    override fun recordCancelled() {}
    override fun findTextClicked(text: String, locale: String) {}
    override fun initKeyboardClicked() {}
    override fun editKeyboardKeyClicked(key: String) {}
    override fun saveKeyboardLocaleClicked(locale: String) {}
    override fun saveTimeoutClicked(timeout: Int) {}
    override fun saveRecordNameClicked(name: String) {}
}
