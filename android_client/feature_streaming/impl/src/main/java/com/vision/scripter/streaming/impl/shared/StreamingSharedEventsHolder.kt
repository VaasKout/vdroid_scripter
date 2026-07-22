package com.vision.scripter.streaming.impl.shared

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.ui.CommandFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class StreamingSharedEventsHolder @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
) {
    val coroutineScope = coroutineScopeFactory.createBackgroundScope("StreamingSharedEventsHandler")
    val sharedEventsFlow: CommandFlow<StreamingSharedEvent> = CommandFlow(coroutineScope)

    fun emit(event: StreamingSharedEvent) {
        sharedEventsFlow.tryEmit(event)
    }
}