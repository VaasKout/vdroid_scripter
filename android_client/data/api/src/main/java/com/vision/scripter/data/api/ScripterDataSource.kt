package com.vision.scripter.data.api

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Library
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.network.api.ApiResponse

interface ScripterDataSource {
    suspend fun getDevices(): ApiResponse<List<AdbDevice>>
    suspend fun getDevicePreview(serial: String): ApiResponse<ByteArray>
    suspend fun pingServer(): Boolean

    suspend fun startSession(serial: String): ApiResponse<StreamingData>

    suspend fun getLibrary(): ApiResponse<Library>

    suspend fun saveImage(
        serial: String,
        name: String,
        rectangle: CvRectangle?,
    ): Boolean

    suspend fun saveAction(
        name: String,
        screenWidth: Int,
        screenHeight: Int,
        events: List<Event>,
    ): Boolean

    suspend fun deleteImage(name: String): Boolean
    suspend fun deleteAction(name: String): Boolean

    suspend fun findText(
        serial: String,
        text: String,
        locale: String,
    ): ApiResponse<List<RectangleWithText>>

    suspend fun resetKeyboard(serial: String, locale: String): ApiResponse<List<RectangleWithText>>
    suspend fun getKeyboard(serial: String, locale: String): ApiResponse<List<RectangleWithText>>

    suspend fun editKeyboard(
        serial: String,
        locale: String,
        name: String,
        rectangle: CvRectangle?,
    ): Boolean

    suspend fun deleteButton(
        serial: String,
        locale: String,
        name: String,
    ): Boolean
}
