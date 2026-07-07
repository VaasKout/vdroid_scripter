package com.vision.scripter.streaming.impl.blocks.menu.state

import androidx.compose.runtime.Stable
import com.vision.scripter.ui.CommandFlow

@Stable
interface MenuUiStateHolder {

    val uiCommandsFlow: CommandFlow<MenuUiCommand>

    fun onScriptModeClicked()
    fun onCvModeClicked()

    fun onTextModeClicked()
    fun onTryToFindText(text: String, locale: String)

    fun onKeyboardClicked()
    fun onKeyboardInitClicked()
    fun onKeyboardEdited(addNew: Boolean)
    fun onEditKeyboardButtonSaved(name: String)

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
