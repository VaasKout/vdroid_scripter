package com.vision.scripter.streaming.impl.blocks.menu.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.streaming.impl.data.ItemType
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface MenuUiStateHolder {

    val uiCommandsFlow: CommandFlow<MenuUiCommand>
    val uiStateFlow: StateFlow<MenuUiState>

    fun init(serial: String)

    fun onAddClicked()
    fun onAddItemConfirmed(name: String, itemType: ItemType)

    fun onRectanglesClicked()
    fun onRefreshRectanglesClicked()

    fun onScanClicked()
    fun onScanConfirmed(locale: String, includeImages: Boolean)

    fun onRecordingClicked()
    fun onSaveClicked()
    fun onExpandClicked()
    fun onCancelClicked()
    fun onExitClicked()

    fun onDialogDismissed()
}
