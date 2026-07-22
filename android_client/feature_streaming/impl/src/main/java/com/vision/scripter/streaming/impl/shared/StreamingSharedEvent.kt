package com.vision.scripter.streaming.impl.shared

import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.streaming.impl.screen.main.state.CVMode

// video -> menu
sealed interface VideoToMenu : StreamingSharedEvent {
    data object TextFound : VideoToMenu
    data class SetKeyboardLoading(val isLoading: Boolean) : VideoToMenu
    data object ScriptSaved : VideoToMenu
    data class SelectKeyboardKey(val oldKey: String) : VideoToMenu
}

// screen -> video
sealed interface ScreenToVideo : StreamingSharedEvent {
    data object StartLoading : ScreenToVideo
    data class InitArgs(val serial: String) : ScreenToVideo
}

// video -> screen
sealed interface VideoToScreen : StreamingSharedEvent {
    data object ShowNetworkError : VideoToScreen
    data object ShowScriptSavedSnackbar : VideoToScreen
    data object SuccessLoading : VideoToScreen
}

// menu -> video
sealed interface MenuToVideo : StreamingSharedEvent {
    data class SaveClicked(val param: Parameter?) : MenuToVideo
    data class NextCvMode(val cvMode: CVMode) : MenuToVideo
    data object RecordCancelled : MenuToVideo
    data class FindText(val text: String, val locale: String) : MenuToVideo
    data object KeyboardInit : MenuToVideo
    data class EditKeyboardButton(val newKey: String, val oldKey: String) : MenuToVideo
    data class SaveLocale(val locale: String) : MenuToVideo
    data class SaveRecordName(val name: String) : MenuToVideo
    data class TimeoutSaved(val timeout: Int) : MenuToVideo
}

sealed interface StreamingSharedEvent