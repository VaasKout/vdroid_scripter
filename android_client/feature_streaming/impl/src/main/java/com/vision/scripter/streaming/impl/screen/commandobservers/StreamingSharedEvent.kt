package com.vision.scripter.streaming.impl.screen.commandobservers

import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode

// video -> menu
sealed interface VideoToMenu : StreamingSharedEvent {
    data class SetKeyboardLoading(val isLoading: Boolean) : VideoToMenu
    data class SelectKeyboardKey(val oldKey: String) : VideoToMenu
}

// video -> screen
sealed interface VideoToScreen : StreamingSharedEvent {
    data object ShowNetworkError : VideoToScreen
    data object ShowScriptSavedSnackbar : VideoToScreen
    data object SuccessLoading : VideoToScreen
}

// screen -> video
sealed interface ScreenToVideo : StreamingSharedEvent {
    data object Refresh : ScreenToVideo
}

// menu -> video
sealed interface MenuToVideo : StreamingSharedEvent {
    data object SaveClicked : MenuToVideo
    data object CancelClicked : MenuToVideo
    data class OnCvModeClicked(val nextMode: CVMode, val recording: Boolean) : MenuToVideo
    data class OnTextModeClicked(val recording: Boolean) : MenuToVideo
    data class FindText(val text: String, val locale: String) : MenuToVideo
    data class KeyboardStateChanged(val keyboardMode: KeyboardMode) : MenuToVideo
    data class KeyboardButtonEdited(val oldKey: String, val newKey: String) : MenuToVideo
    data class KeyboardLocaleSaved(val locale: String) : MenuToVideo
    data class RecordNameSaved(val name: String) : MenuToVideo
    data class TimeoutSaved(val timeout: Int) : MenuToVideo
    data class OnRecordingClicked(val isControlRecording: Boolean) : MenuToVideo
}

sealed interface StreamingSharedEvent