package com.vision.scripter.streaming.impl.blocks.menu.state

sealed interface MenuEvent {
    data class SaveTemplate(val flags: Int) : MenuEvent
    data class SaveText(val text: String, val locale: String, val flags: Int) : MenuEvent
    data class SaveTyping(val text: String, val flags: Int) : MenuEvent
    data object SaveStep : MenuEvent
    data object RecordCancelled : MenuEvent
    data class FindText(val text: String, val locale: String) : MenuEvent
    data object KeyboardInit : MenuEvent
    data class EditKeyboardButton(val name: String) : MenuEvent
    data class SaveLocale(val locale: String) : MenuEvent
    data class SaveRecordName(val name: String) : MenuEvent
    data class TimeoutSaved(val timeout: Int) : MenuEvent
}
