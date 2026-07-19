package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.data.api.models.Parameter

sealed interface MenuEvent {
    data class SaveClicked(val param: Parameter?) : MenuEvent
    data object RecordCancelled : MenuEvent
    data class FindText(val text: String, val locale: String) : MenuEvent
    data object KeyboardInit : MenuEvent
    data class EditKeyboardButton(val name: String) : MenuEvent
    data class SaveLocale(val locale: String) : MenuEvent
    data class SaveRecordName(val name: String) : MenuEvent
    data class TimeoutSaved(val timeout: Int) : MenuEvent
}
