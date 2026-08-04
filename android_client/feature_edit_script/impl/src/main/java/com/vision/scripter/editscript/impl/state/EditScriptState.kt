package com.vision.scripter.editscript.impl.state

import com.vision.scripter.data.api.models.Script
import com.vision.scripter.ui.states.LoadingState

data class EditScriptState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val initialNode: String = "",
    val script: Script? = null,
    val showDialog: Boolean = false,
)
