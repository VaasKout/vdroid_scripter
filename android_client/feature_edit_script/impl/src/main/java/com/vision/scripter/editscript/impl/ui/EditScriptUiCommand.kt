package com.vision.scripter.editscript.impl.ui

sealed class EditScriptUiCommand {
    data object ShowNetworkError : EditScriptUiCommand()
    data object NavigateBack : EditScriptUiCommand()
}
