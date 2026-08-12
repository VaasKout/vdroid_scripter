package com.vision.scripter.streaming.impl.data

import com.vision.scripter.data.api.CvStreamer
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.smallestBy
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.streaming.impl.di.StreamingScope
import com.vision.scripter.streaming.impl.screen.state.CVMode
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.update
import kotlinx.serialization.json.Json
import java.util.concurrent.atomic.AtomicReference
import javax.inject.Inject

@StreamingScope
class CvStreamerRepository @Inject constructor(
    private val cvStreamer: CvStreamer,
    private val scripterDataSource: ScripterDataSource,
) {
    private val cvMode = MutableStateFlow(CVMode.NO_CV)

    private val _cvRectanglesFlow = MutableStateFlow<List<CvRectangle>>(listOf())
    private val _yoloRectanglesFlow = MutableStateFlow<List<CvRectangle>>(listOf())
    private val _selectedRectangles = MutableStateFlow<List<CvRectangle>>(listOf())

    private val selectedRectanglesSnapshot = AtomicReference<List<CvRectangle>>(listOf())

    fun observeRectangles(): Flow<List<CvRectangle>> =
        combine(
            _cvRectanglesFlow,
            _yoloRectanglesFlow,
            cvMode,
        ) { cvRects, yoloRects, mode ->
            when (mode) {
                CVMode.NO_CV -> listOf()
                CVMode.CV_RECTS -> cvRects
                CVMode.YOLO -> yoloRects
            }
        }

    fun observeSelectedRectangles(): StateFlow<List<CvRectangle>> =
        _selectedRectangles.asStateFlow()

    suspend fun initConnection(
        host: String,
        port: Int,
    ): Boolean {
        try {
            return cvStreamer.connect(
                host = host,
                port = port,
            )
        } catch (e: Exception) {
            e.printStackTrace()
        }
        return false
    }

    suspend fun decodeRectanglesInLoop(
        screenSizes: ScreenSizes,
    ) {
        while (true) {
            if (cvMode.value == CVMode.NO_CV) {
                delay(100)
                continue
            }
            val buffer = cvStreamer.readRectangles() ?: continue
            val jsonString = buffer.decodeToString()
            val rectangles = try {
                if (jsonString.isNotEmpty()) {
                    Json.decodeFromString<List<CvRectangle>>(jsonString)
                } else {
                    listOf()
                }
            } catch (_: Exception) {
                listOf()
            }

            val adjusted = rectangles.adjustToClient(screenSizes)
            when (cvMode.value) {
                CVMode.CV_RECTS -> _cvRectanglesFlow.update { adjusted }
                CVMode.YOLO -> _yoloRectanglesFlow.update { adjusted }
                else -> {}
            }
        }
    }

    suspend fun nextCvMode(newCvMode: CVMode) {
        if (newCvMode != cvMode.value) {
            _cvRectanglesFlow.update { listOf() }
            _yoloRectanglesFlow.update { listOf() }
            cvStreamer.sendCvMode(newCvMode.value)
            cvMode.value = newCvMode
        }
    }

    suspend fun findTextRectangles(
        serial: String,
        text: String,
        locale: String,
        screenSizes: ScreenSizes,
    ): Boolean {
        val result = scripterDataSource.findText(serial = serial, text = text, locale = locale)
        if (result is ApiResponse.Success) {
            val rectangles = result.data.mapNotNull { it.rectangle }
            _selectedRectangles.update {
                rectangles.adjustToClient(screenSizes)
            }
        }

        return result is ApiResponse.Success && result.data.isNotEmpty()
    }

    fun selectRectangle(x: Int, y: Int) {
        val current = when (cvMode.value) {
            CVMode.NO_CV -> listOf()
            CVMode.CV_RECTS -> _cvRectanglesFlow.value
            CVMode.YOLO -> _yoloRectanglesFlow.value
        }
        val selected = current.smallestBy(x, y)
        setSelectedRectangle(selected)
    }

    fun setSelectedRectangle(rect: CvRectangle?) {
        _selectedRectangles.value = listOfNotNull(rect)
    }

    fun clearSelectedRectangles() {
        _selectedRectangles.update { listOf() }
    }

    fun snapshotSelectedRectangles() {
        selectedRectanglesSnapshot.set(_selectedRectangles.value)
    }

    fun restoreSelectedRectangles() {
        _selectedRectangles.value = selectedRectanglesSnapshot.getAndSet(listOf())
    }

    fun clearAllRectangles() {
        _cvRectanglesFlow.update { listOf() }
        _yoloRectanglesFlow.update { listOf() }
        _selectedRectangles.update { listOf() }
    }

    fun close() {
        cvStreamer.close()
        _cvRectanglesFlow.update { listOf() }
        _yoloRectanglesFlow.update { listOf() }
        _selectedRectangles.update { listOf() }
        cvMode.value = CVMode.NO_CV
    }
}