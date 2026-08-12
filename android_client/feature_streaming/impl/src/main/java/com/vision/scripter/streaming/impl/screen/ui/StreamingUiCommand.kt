package com.vision.scripter.streaming.impl.screen.ui

sealed class StreamingUiCommand {
    data object ShowNetworkError : StreamingUiCommand()
    data object ShowItemSavedSnackbar : StreamingUiCommand()
}