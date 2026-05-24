package com.vision.scripter.streaming.impl.usecases

import com.vision.scripter.data.api.CvStreamer
import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.adjustToServer
import com.vision.scripter.data.api.models.isEmpty
import com.vision.scripter.data.api.models.smallestBy
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.streaming.impl.state.CVMode
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.serialization.json.Json
import javax.inject.Inject

@ViewModelScoped
class CvUseCase @Inject constructor(
    private val cvStreamer: CvStreamer,
    private val scripterRepository: ScripterRepository,
) {
    private val cvMode = MutableStateFlow(CVMode.NO_CV)
    private val _rectanglesFlow = MutableStateFlow<List<CvRectangle>>(listOf())

    fun observe(
        coroutineScope: CoroutineScope,
    ): StateFlow<List<CvRectangle>> =
        combine(
            _rectanglesFlow,
            cvMode
        ) { rects, cvMode ->
            if (cvMode == CVMode.NO_CV) rects.filter { rect -> rect.isSelected }
            else rects
        }.stateIn(
            scope = coroutineScope,
            started = SharingStarted.WhileSubscribed(),
            initialValue = listOf(),
        )

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
            if (cvMode.value != CVMode.CV_RECTS) {
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

            _rectanglesFlow.update {
                rectangles.adjustToClient(screenSizes)
            }
        }
    }

    suspend fun nextCvMode(newCvMode: CVMode) {
        if (newCvMode != cvMode.value) {
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
        val result = scripterRepository.findText(serial = serial, text = text, locale = locale)
        if (result is ApiResponse.Success) {
            val rectangles = result.data.mapNotNull { it.rectangle }
            _rectanglesFlow.update {
                rectangles.adjustToClient(screenSizes).map { it.copy(isSelected = true) }
            }
        }

        return result is ApiResponse.Success
    }

    suspend fun saveSelectedRectangle(
        serial: String,
        screenSizes: ScreenSizes,
    ) {
        val tmpZone = _rectanglesFlow.value.firstOrNull { it.isSelected }
        if (!tmpZone.isEmpty()) {
            scripterRepository.saveRect(
                serial = serial,
                rectangle = tmpZone?.adjustToServer(screenSizes),
            )
        }
    }

    fun selectRectangle(x: Int, y: Int) {
        _rectanglesFlow.update {
            val selectedZone = it.smallestBy(x, y)
            it.map { rectangle ->
                rectangle.copy(isSelected = rectangle == selectedZone)
            }
        }
    }

    fun disableSelection() {
        _rectanglesFlow.update {
            it.map { rect -> rect.copy(isSelected = false) }
        }
    }

    fun selectAll() {
        _rectanglesFlow.update {
            it.map { rect -> rect.copy(isSelected = true) }
        }
    }


    fun clearAllRectangles() {
        _rectanglesFlow.update { listOf() }
    }

    fun close() {
        cvStreamer.close()
        _rectanglesFlow.update { listOf() }
        cvMode.value = CVMode.NO_CV
    }
}