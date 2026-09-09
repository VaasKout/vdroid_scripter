package com.vision.scripter.libraryitems.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface LibraryItemsUiStateHolder {
    val uiStateFlow: StateFlow<LibraryItemsUiState>
    val uiCommandsFlow: CommandFlow<LibraryItemsUiCommand>

    fun init(type: LibraryType)
    fun onLoadData(onStart: Boolean)

    fun onDeleteItem(name: String)
    fun onDismissDelete()
    fun onConfirmDelete()

    fun onPlayClicked(name: String)
    fun onDeviceChosen(serial: String)
    fun onPickerDismissed()
    fun onRunDismissed(name: String)
}
