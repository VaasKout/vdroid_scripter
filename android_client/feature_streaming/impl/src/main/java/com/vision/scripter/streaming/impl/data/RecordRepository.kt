package com.vision.scripter.streaming.impl.data

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.streaming.impl.di.StreamingScope
import com.vision.scripter.streaming.impl.screen.state.DEFAULT_TIMEOUT
import com.vision.scripter.streaming.impl.screen.state.NEW_SCRIPTS_NODE
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.firstOrNull
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@StreamingScope
class RecordRepository @Inject constructor(
    private val scripterDataSource: ScripterDataSource,
    private val keyboardRepository: KeyboardRepository,
    coroutineScopeFactory: CoroutineScopeFactory,
) {

    private val _stateFlow = MutableStateFlow(Record())
    fun observeRecord(): StateFlow<Record> = _stateFlow.asStateFlow()

    private val coroutineScope = coroutineScopeFactory.createBackgroundScope("RecordRepository")

    private val currentState: Record
        get() = _stateFlow.value

    private var startRecordingTime = 0L

    fun setControlRecording(isControlRecording: Boolean) {
        _stateFlow.update {
            it.copy(controlRecording = isControlRecording)
        }
    }

    fun updateName(name: String) {
        _stateFlow.update {
            it.copy(name = name)
        }
    }

    fun updateTimeout(timeout: Int) {
        _stateFlow.update {
            it.copy(timeout = timeout)
        }
    }

    fun addParam(param: Parameter) {
        _stateFlow.update {
            it.copy(params = it.params + param)
        }
    }

    suspend fun recordBytes(bytesArray: ByteArray?) {
        if (bytesArray == null) return
        val keyboardButtons = keyboardRepository.observeKeyboardButtons().firstOrNull().orEmpty()
        val record = currentState
        val keyboardRecording = keyboardButtons.isNotEmpty() && record.name.isNotEmpty()
        if (!record.controlRecording && !keyboardRecording) return

        val elapsedMs = if (startRecordingTime == 0L) {
            startRecordingTime = System.nanoTime()
            0L
        } else {
            (System.nanoTime() - startRecordingTime) / 1_000_000L
        }

        val newEvent = Event(
            time = elapsedMs,
            data = bytesArray,
        )

        _stateFlow.update {
            it.copy(events = it.events + newEvent)
        }
    }

    suspend fun saveScript(): Boolean {
        val record = currentState
        return scripterDataSource.saveScript(
            Script(
                name = record.name,
                node = record.node,
                params = record.params,
                events = record.events,
                timeout = record.timeout,
            ),
        )
    }

    fun clear(saveName: Boolean = false) {
        startRecordingTime = 0L
        _stateFlow.update { Record(name = if (saveName) it.name else "") }
    }

    fun cancel() {
        coroutineScope.cancel()
    }
}

data class Record(
    val controlRecording: Boolean = false,
    val name: String = "",
    val node: String = NEW_SCRIPTS_NODE,
    val params: List<Parameter> = listOf(),
    val events: List<Event> = listOf(),
    val timeout: Int = DEFAULT_TIMEOUT,
)
