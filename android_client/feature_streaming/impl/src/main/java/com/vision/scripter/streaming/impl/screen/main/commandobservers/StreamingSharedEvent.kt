package com.vision.scripter.streaming.impl.screen.main.commandobservers

import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.streaming.impl.screen.main.state.CVMode

sealed interface StreamingSharedEvent {
    data class LoadData(val onStart: Boolean) : StreamingSharedEvent
    data class InitArgs(val serial: String) : StreamingSharedEvent

    data class SaveClicked(val param: Parameter?) : StreamingSharedEvent
    data class NextCvMode(val cvMode: CVMode) : StreamingSharedEvent
    data object RecordCancelled : StreamingSharedEvent
    data class FindText(val text: String, val locale: String) : StreamingSharedEvent
    data object KeyboardInit : StreamingSharedEvent
    data class EditKeyboardButton(val newKey: String, val oldKey: String) : StreamingSharedEvent
    data class SaveLocale(val locale: String) : StreamingSharedEvent
    data class SaveRecordName(val name: String) : StreamingSharedEvent
    data class TimeoutSaved(val timeout: Int) : StreamingSharedEvent
}