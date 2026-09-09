package com.vision.scripter.libraryitems.ui

sealed class LibraryItemsUiCommand {
    data object ShowNetworkError : LibraryItemsUiCommand()
    data class ShowError(val text: String) : LibraryItemsUiCommand()
}
