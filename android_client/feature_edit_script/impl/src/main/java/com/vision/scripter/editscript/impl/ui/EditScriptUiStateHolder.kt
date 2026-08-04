package com.vision.scripter.editscript.impl.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.SharedFlow

@Stable
interface EditScriptUiStateHolder {
    val uiStateFlow: SharedFlow<EditScriptUiState>
    val uiCommandsFlow: CommandFlow<EditScriptUiCommand>

    fun init(node: String, name: String)

    fun onNodeChanged(value: String)
    fun onNextNodeChanged(value: String)
    fun onTimeoutChanged(value: String)

    fun onDeleteParam(id: Int)
    fun onDeleteEvents()

    fun onTopbarActionClicked()
    fun onConfirmDialog()
    fun onDismissDialog()
    fun onBackClicked()
}
