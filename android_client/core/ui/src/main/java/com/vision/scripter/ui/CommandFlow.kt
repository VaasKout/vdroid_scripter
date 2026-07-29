package com.vision.scripter.ui

import androidx.compose.runtime.Stable
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.Job
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.AbstractFlow
import kotlinx.coroutines.flow.FlowCollector
import kotlinx.coroutines.flow.consumeAsFlow
import kotlinx.coroutines.flow.emitAll
import kotlinx.coroutines.flow.receiveAsFlow
import java.util.concurrent.CopyOnWriteArrayList

@Stable
@OptIn(ExperimentalCoroutinesApi::class)
class CommandFlow<T>(scope: CoroutineScope) : AbstractFlow<T>() {

    private val channel = Channel<T>(Channel.UNLIMITED)

    init {
        scope.coroutineContext[Job]?.invokeOnCompletion {
            channel.cancel()
        }
    }

    fun tryEmit(value: T): Boolean {
        return channel.trySend(value).isSuccess
    }

    override suspend fun collectSafely(collector: FlowCollector<T>) {
        collector.emitAll(channel.receiveAsFlow())
    }
}

infix fun <T> CommandFlow<T>.emit(value: T) {
    this.tryEmit(value)
}

@Stable
@OptIn(
    ExperimentalCoroutinesApi::class
)
class CommandFlow2<T> : AbstractFlow<T>() {

    private val channels = CopyOnWriteArrayList<Channel<T>>()

    fun tryEmit(value: T) {
        channels.forEach {
            it.trySend(value)
        }
    }

    override suspend fun collectSafely(collector: FlowCollector<T>) {
        val channel = Channel<T>(Channel.UNLIMITED)
        channels.add(channel)
        try {
            collector.emitAll(channel.consumeAsFlow())
        } finally {
            channels.remove(channel)
        }
    }
}

@Suppress("NOTHING_TO_INLINE")
inline infix fun <T> CommandFlow2<T>.emit(value: T) {
    tryEmit(value)
}