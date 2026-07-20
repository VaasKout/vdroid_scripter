package com.vision.scripter.streaming.impl.blocks.video.ui

sealed class VideoUiCommand {
    data object ShowNetworkError : VideoUiCommand()
    data object ShowScriptSavedSnackbar : VideoUiCommand()

    data object TextFound : VideoUiCommand()
    data class SetKeyboardLoading(val isLoading: Boolean) : VideoUiCommand()
    data object ScriptSaved : VideoUiCommand()
    data class SelectKeyboardKey(val oldKey: String) : VideoUiCommand()
}
