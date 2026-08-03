package com.vision.scripter.editscript.impl.ui

import com.vision.scripter.ui.CommandFlow
import kotlinx.collections.immutable.persistentListOf
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow

internal val editScriptUiStatePreview = EditScriptUiState(
    isLoading = false,
    name = "login_tap",
    node = "main_screen",
    nextNode = "profile",
    timeout = "15",
    params = persistentListOf(
        ParamUiData(id = 0, title = "template: login_button", locale = ""),
        ParamUiData(id = 1, title = "text: Sign in", locale = "eng"),
    ),
    eventsCount = 24,
)

internal class EditScriptUiStateHolderPreview(state: EditScriptUiState) : EditScriptUiStateHolder {

    override val uiStateFlow: SharedFlow<EditScriptUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<EditScriptUiCommand>
        get() = throw UnsupportedOperationException()

    override fun init(node: String, name: String) {}
    override fun onNodeChanged(value: String) {}
    override fun onNextNodeChanged(value: String) {}
    override fun onTimeoutChanged(value: String) {}
    override fun onDeleteParam(id: Int) {}
    override fun onDeleteEvents() {}
    override fun onSaveClicked() {}
    override fun onConfirmSave() {}
    override fun onDismissDialog() {}
    override fun onBackClicked() {}
}
