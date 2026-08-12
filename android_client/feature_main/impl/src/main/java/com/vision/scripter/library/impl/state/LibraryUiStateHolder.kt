package com.vision.scripter.library.impl.state

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.SharedFlow

@Stable
interface LibraryUiStateHolder {
    val uiStateFlow: SharedFlow<LibraryUiState>
    val uiCommandsFlow: CommandFlow<LibraryUiCommand>

    fun onLoadData(onStart: Boolean)

    fun onDeleteItem(kind: LibraryKind, name: String)
    fun onDismiss()
    fun onConfirmDelete()
}
