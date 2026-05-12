package com.vision.scripter.streaming.impl.ui

import android.view.MotionEvent
import android.view.Surface
import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.SharedFlow

@Stable
interface StreamingUiStateHolder {
    val uiStateFlow: SharedFlow<StreamingUiState>
    val uiCommandsFlow: CommandFlow<StreamingUiCommand>

    fun initArgs(serial: String)
    fun onLoadData(onStart: Boolean)
    fun onVideoSurfaceCreated(surfaceWidth: Int, surfaceHeight: Int, newSurface: Surface)
    fun onVideoSurfaceDestroyed()
    fun onTouchEvent(viewWidth: Int, viewHeight: Int, event: MotionEvent?)

    fun onScriptModeClicked()
    fun onCvModeClicked()

    fun onTextModeClicked()
    fun onTryToFindText(text: String)

    fun onKeyboardClicked()
    fun onKeyboardInitClicked()

    fun onRecordingClicked()
    fun onSaveClicked()
    fun onExpandClicked()
    fun onCancelClicked()

    fun onSavedRecordName(name: String)
    fun onSaveLocale(locale: String)
    fun onDialogDismissed()
}