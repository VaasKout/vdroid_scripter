package com.vision.scripter.streaming.impl.data

import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.ScreenSizes
import com.vision.scripter.data.api.models.adjustToClient
import com.vision.scripter.data.api.models.smallestBy
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.streaming.impl.di.StreamingScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import javax.inject.Inject

@StreamingScope
class CvRepository @Inject constructor(
    private val scripterDataSource: ScripterDataSource,
) {
    private val _overlayFlow = MutableStateFlow(CvOverlay())
    fun observeOverlay(): StateFlow<CvOverlay> = _overlayFlow.asStateFlow()

    private val _selectedRectangles = MutableStateFlow<List<CvRectangle>>(listOf())
    fun observeSelectedRectangles(): StateFlow<List<CvRectangle>> =
        _selectedRectangles.asStateFlow()

    suspend fun refreshRectangles(
        serial: String,
        screenSizes: ScreenSizes,
    ): Boolean {
        val result = scripterDataSource.getRectangles(serial)
        if (result !is ApiResponse.Success) return false
        _overlayFlow.value = CvOverlay(rectangles = result.data.adjustToClient(screenSizes))
        return true
    }

    suspend fun scan(
        serial: String,
        images: List<String>,
        locale: String,
        screenSizes: ScreenSizes,
    ): Boolean {
        val result = scripterDataSource.scan(serial = serial, images = images, locale = locale)
        if (result !is ApiResponse.Success) return false
        val labelled = result.data.map {
            it.rectangle.copy(label = "${it.type} ${it.value}")
        }
        _overlayFlow.value = CvOverlay(scan = labelled.adjustToClient(screenSizes))
        return true
    }

    fun clearOverlay() {
        _overlayFlow.value = CvOverlay()
    }

    fun selectRectangle(x: Int, y: Int) {
        setSelectedRectangle(_overlayFlow.value.rectangles.smallestBy(x, y))
    }

    fun setSelectedRectangle(rect: CvRectangle?) {
        _selectedRectangles.value = listOfNotNull(rect)
    }

    fun clearSelectedRectangles() {
        _selectedRectangles.value = listOf()
    }

    fun close() {
        clearOverlay()
        clearSelectedRectangles()
    }
}

data class CvOverlay(
    val rectangles: List<CvRectangle> = listOf(),
    val scan: List<CvRectangle> = listOf(),
)
