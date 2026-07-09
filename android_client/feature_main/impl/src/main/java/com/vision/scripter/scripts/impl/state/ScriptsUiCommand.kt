package com.vision.scripter.scripts.impl.state

sealed class ScriptsUiCommand {
    data object ShowNetworkError : ScriptsUiCommand()
}