package com.vision.scripter.data.api

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.network.api.ApiResponse

interface ScripterDataSource {
    suspend fun getDevices(): ApiResponse<List<AdbDevice>>
    suspend fun getDevicePreview(serial: String): ApiResponse<ByteArray>
    suspend fun pingServer(): Boolean

    suspend fun startSession(serial: String): ApiResponse<StreamingData>

    suspend fun saveRect(
        serial: String,
        location: String,
        name: String,
        value: String,
        rectangle: CvRectangle?,
    ): Boolean

    suspend fun findText(
        serial: String,
        text: String,
        locale: String,
    ): ApiResponse<List<RectangleWithText>>

    suspend fun getLocations(): ApiResponse<List<String>>
    suspend fun getLocationScripts(location: String): ApiResponse<List<String>>
    suspend fun getScriptInfo(location: String, name: String): ApiResponse<Script>
    suspend fun deleteLocation(location: String): Boolean
    suspend fun saveScript(script: Script): Boolean
    suspend fun editScript(script: Script, prevLocation: String): Boolean
    suspend fun deleteScript(location: String, name: String): Boolean
    suspend fun runScript(serial: String, location: String, name: String): Boolean
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