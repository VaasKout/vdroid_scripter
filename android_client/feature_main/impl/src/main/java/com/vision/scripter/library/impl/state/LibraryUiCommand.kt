package com.vision.scripter.library.impl.state

sealed class LibraryUiCommand {
    data object ShowNetworkError : LibraryUiCommand()
}
