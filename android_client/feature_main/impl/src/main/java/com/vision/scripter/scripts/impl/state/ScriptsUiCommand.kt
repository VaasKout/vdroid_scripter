package com.vision.scripter.scripts.impl.state

sealed class ScriptsUiCommand {
    data object ShowNetworkError : ScriptsUiCommand()
    data class OpenEditScript(val node: String, val name: String) : ScriptsUiCommand()
}