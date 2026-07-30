package com.vision.scripter.streaming.impl.screen

import com.vision.scripter.streaming.impl.di.StreamingScope
import com.vision.scripter.ui.CommandFlow2
import javax.inject.Inject

@StreamingScope
class StreamingEventsHolder @Inject constructor() {

    val eventsFlow: CommandFlow2<StreamingEvent> = CommandFlow2()

    fun sendEvent(event: StreamingEvent) {
        eventsFlow.tryEmit(event)
    }
}

sealed interface StreamingEvent {
    data object ShowNetworkError : StreamingEvent
    data object ShowScriptSavedSnackbar : StreamingEvent
    data object SuccessLoading : StreamingEvent
    data class SelectKeyboardKey(val oldKey: String) : StreamingEvent
}
