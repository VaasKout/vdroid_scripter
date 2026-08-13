package com.vision.scripter.library.state

sealed class LibraryUiCommand {
    data object ShowNetworkError : LibraryUiCommand()
}
