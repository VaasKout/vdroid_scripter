package com.vision.scripter.main.impl.ui

sealed class MainUiCommand {
    data object ShowNetworkError : MainUiCommand()
    data object NavigateToScripts : MainUiCommand()
    data class NavigateToStreaming(val serial: String) : MainUiCommand()
}