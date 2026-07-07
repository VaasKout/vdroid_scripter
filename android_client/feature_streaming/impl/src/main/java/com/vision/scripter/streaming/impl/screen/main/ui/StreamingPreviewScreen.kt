package com.vision.scripter.streaming.impl.screen.main.ui

import android.view.MotionEvent
import android.view.Surface
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.SharedFlow

internal val streamingPreviewState = StreamingUiState()

internal class StreamingPreviewScreenUiStateHolder(state: StreamingUiState) :
    StreamingUiStateHolder {
    override val uiStateFlow: SharedFlow<StreamingUiState>
        get() = throw UnsupportedOperationException()
    override val uiCommandsFlow: CommandFlow<StreamingUiCommand>
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

    override fun onScriptModeClicked() {}

    override fun onCvModeClicked() {}
    override fun onTextModeClicked() {}
    override fun onKeyboardClicked() {}
    override fun onKeyboardInitClicked() {}
    override fun onKeyboardEdited(addNew: Boolean) {}
    override fun onEditKeyboardButtonSaved(name: String) {}
    override fun onTryToFindText(text: String, locale: String) {}

    override fun onRecordingClicked() {}
    override fun onSaveClicked() {}
    override fun onExpandClicked() {}
    override fun onCancelClicked() {}
    override fun exit() {}
    override fun onTimeoutClicked() {}
    override fun onTimeoutSaved(timeout: Int) {}
    override fun onSavedRecordName(name: String) {}
    override fun onSaveLocale(locale: String) {}
    override fun onDialogDismissed() {}
}