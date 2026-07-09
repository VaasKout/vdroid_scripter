package com.vision.scripter.scripts.impl.ui

import com.vision.scripter.scripts.impl.state.ScriptsUiCommand
import com.vision.scripter.scripts.impl.state.ScriptsUiState
import com.vision.scripter.scripts.impl.state.ScriptsUiStateHolder
import com.vision.scripter.ui.CommandFlow
import kotlinx.collections.immutable.persistentListOf
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow

internal val scriptsUiStatePreview = ScriptsUiState(
    scripts = persistentListOf(
        "test_1",
        "test_2",
    )
)

internal class ScriptsScreenUiStateHolderPreview(state: ScriptsUiState) : ScriptsUiStateHolder {

    override val uiStateFlow: SharedFlow<ScriptsUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<ScriptsUiCommand>
        get() = throw UnsupportedOperationException()

    override fun onLoadData(onStart: Boolean) {}
    override fun onPlayScript(name: String) {}
    override fun onDeleteScript(name: String) {}
    override fun onDismissDeleteDialog() {}
    override fun onConfirmDeleteScript() {}
}