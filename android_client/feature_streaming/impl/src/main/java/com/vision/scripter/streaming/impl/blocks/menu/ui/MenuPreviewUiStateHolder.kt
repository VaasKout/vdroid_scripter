package com.vision.scripter.streaming.impl.blocks.menu.ui

import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow

internal val usualMenuPreviewUiState = MenuUiState()

internal class MenuPreviewUiStateHolder(state: MenuUiState) : MenuUiStateHolder {
    override val uiCommandsFlow: CommandFlow<MenuUiCommand>
        get() = throw UnsupportedOperationException()
    override val uiStateFlow: StateFlow<MenuUiState> = MutableStateFlow(state)

    override fun onScriptModeClicked() {}
    override fun onCvModeClicked() {}
    override fun onTextModeClicked() {}
    override fun onTryToFindText(text: String, locale: String) {}
    override fun onKeyboardClicked() {}
    override fun onKeyboardModeClicked() {}
    override fun onEditKeyboardButtonSaved(oldKey: String, newKey: String) {}
    override fun onRecordingClicked() {}
    override fun onSaveClicked() {}
    override fun onExpandClicked() {}
    override fun onCancelClicked() {}
    override fun onExitClicked() {}
    override fun onTimeoutClicked() {}
    override fun onTimeoutSaved(timeout: Int) {}
    override fun onSavedRecordName(name: String) {}
    override fun onSaveLocale(locale: String) {}
    override fun onDialogDismissed() {}
}