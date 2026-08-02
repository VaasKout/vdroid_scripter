package com.vision.scripter.scripts.impl.state

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.SharedFlow

@Stable
interface ScriptsUiStateHolder {
    val uiStateFlow: SharedFlow<ScriptsUiState>
    val uiCommandsFlow: CommandFlow<ScriptsUiCommand>

    fun onLoadData(onStart: Boolean)

    fun onNodeClick(node: String)
    fun onBack()

    fun onPlayScript(name: String)
    fun onSelectDevice(serial: String)
    fun onConfirmRunScript()

    fun onDismiss()
    fun onDeleteItem(item: String)
    fun onConfirmDelete()
}
