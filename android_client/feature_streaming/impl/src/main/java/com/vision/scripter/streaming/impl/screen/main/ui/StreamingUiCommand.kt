package com.vision.scripter.streaming.impl.screen.main.ui

sealed class StreamingUiCommand {
    data object ShowNetworkError : StreamingUiCommand()
    data object ShowScriptSavedSnackbar : StreamingUiCommand()
}