package com.vision.scripter.streaming.impl.data

import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.streaming.impl.di.StreamingScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@StreamingScope
class RecordRepository @Inject constructor(
    private val scripterDataSource: ScripterDataSource,
) {

    private val _stateFlow = MutableStateFlow(Record())
    fun observeRecord(): StateFlow<Record> = _stateFlow.asStateFlow()

    private val currentState: Record
        get() = _stateFlow.value

    private var startRecordingTime = 0L

    fun initData(name: String, itemType: ItemType) {
        startRecordingTime = 0L
        _stateFlow.update {
            Record(name = name, type = itemType)
        }
    }

    fun switchRecording() {
        _stateFlow.update {
            it.copy(recording = !it.recording)
        }
    }

    @OptIn(ExperimentalUnsignedTypes::class)
    fun recordBytes(bytesArray: UByteArray?) {
        if (bytesArray == null) return
        if (!currentState.recording) return

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

    suspend fun saveImage(serial: String, rectangle: CvRectangle?): Boolean {
        val record = currentState
        if (record.name.isEmpty()) return false
        return scripterDataSource.saveImage(
            serial = serial,
            rectangle = rectangle?.copy(label = record.name),
        )
    }

    suspend fun saveAction(screenSizes: ScreenSizes): Boolean {
        val record = currentState
        if (record.name.isEmpty() || record.events.isEmpty()) return false
        return scripterDataSource.saveAction(
            name = record.name,
            screenWidth = screenSizes.remoteWidth,
            screenHeight = screenSizes.remoteHeight,
            events = record.events,
        )
    }

    fun clear() {
        startRecordingTime = 0L
        _stateFlow.update { Record() }
    }
}

enum class ItemType {
    NONE,
    IMAGE,
    ACTION,
}

data class Record(
    val name: String = "",
    val type: ItemType = ItemType.NONE,
    val recording: Boolean = false,
    val events: List<Event> = listOf(),
)
