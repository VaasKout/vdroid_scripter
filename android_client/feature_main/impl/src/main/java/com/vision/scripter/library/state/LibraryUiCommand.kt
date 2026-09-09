package com.vision.scripter.library.state

sealed class LibraryUiCommand {
    data object ShowNetworkError : LibraryUiCommand()
    data class OpenList(val type: LibraryType) : LibraryUiCommand()
}
