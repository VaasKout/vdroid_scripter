package com.vision.scripter.scripts.impl.state

import com.vision.scripter.ui.states.LoadingState

data class ScriptsState(
    val loadingState: LoadingState = LoadingState.LoadingOnStart,
    val scripts: List<String> = listOf(),
    val scriptNameToDelete: String = "",
)