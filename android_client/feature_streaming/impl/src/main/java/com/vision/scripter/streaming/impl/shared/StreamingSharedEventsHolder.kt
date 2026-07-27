package com.vision.scripter.streaming.impl.shared

import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class StreamingSharedEventsHolder @Inject constructor() {
    private val _sharedEventsFlow = MutableSharedFlow<StreamingSharedEvent>(replay = 1)
    val sharedEventsFlow: SharedFlow<StreamingSharedEvent> = _sharedEventsFlow.asSharedFlow()

    fun emit(event: StreamingSharedEvent) {
        _sharedEventsFlow.tryEmit(event)
    }

    @OptIn(ExperimentalCoroutinesApi::class)
    fun close() {
        _sharedEventsFlow.resetReplayCache()
    }
}
