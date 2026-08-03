package com.vision.scripter.editscript.impl.state

import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.ui.states.LoadingState

data class EditScriptState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val initialNode: String = "",
    val scriptName: String = "",
    val node: String = "",
    val nextNode: String = "",
    val timeout: String = "",
    val params: List<Parameter> = listOf(),
    val events: List<Event> = listOf(),
    val showDialog: Boolean = false,
)
