package com.vision.scripter.streaming.impl.ui

sealed class StreamingUiCommand {
    data object ShowNetworkError : StreamingUiCommand()
    data object ShowStepSavedSnackbar: StreamingUiCommand()
    data object ExitCommand : StreamingUiCommand()
}