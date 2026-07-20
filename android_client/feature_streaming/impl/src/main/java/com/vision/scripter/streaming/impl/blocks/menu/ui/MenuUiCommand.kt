package com.vision.scripter.streaming.impl.blocks.menu.ui

import com.vision.scripter.streaming.impl.screen.main.state.CVMode

sealed class MenuUiCommand {
    data object ExitCommand : MenuUiCommand()

    data class NextCvMode(val cvMode: CVMode) : MenuUiCommand()
    data class FindText(val text: String, val locale: String) : MenuUiCommand()
    data object KeyboardInit : MenuUiCommand()
    data class EditKeyboardButton(val key: String) : MenuUiCommand()
    data object SaveParameter : MenuUiCommand()
    data object CancelRecording : MenuUiCommand()
    data class SaveTimeout(val timeout: Int) : MenuUiCommand()
    data class SaveRecordName(val name: String) : MenuUiCommand()
    data class SaveLocale(val locale: String) : MenuUiCommand()
}