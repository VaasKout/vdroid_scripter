package com.vision.scripter.streaming.impl.blocks.menu.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface MenuUiStateHolder {

    val uiCommandsFlow: CommandFlow<MenuUiCommand>
    val uiStateFlow: StateFlow<MenuUiState>

    fun onScriptModeClicked()
    fun onCvModeClicked()

    fun onTextModeClicked()
    fun onTryToFindText(text: String, locale: String)

    fun onKeyboardClicked()
    fun onKeyboardModeClicked()
    fun onEditKeyboardButtonSaved(oldKey: String, newKey: String)

    fun onRecordingClicked()
    fun onSaveClicked()
    fun onExpandClicked()
    fun onCancelClicked()
    fun onExitClicked()

    fun onTimeoutClicked()
    fun onTimeoutSaved(timeout: Int)

    fun onSavedRecordName(name: String)
    fun onSaveLocale(locale: String)
    fun onDialogDismissed()
}