package com.vision.scripter.data.api

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.data.api.models.ScriptStep
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.network.api.ApiResponse

interface ScripterRepository {
    suspend fun getDevices(): ApiResponse<List<AdbDevice>>
    suspend fun getDevicePreview(serial: String): ApiResponse<ByteArray>
    suspend fun pingServer(): Boolean

    suspend fun startSockets(serial: String): ApiResponse<StreamingData>

    suspend fun saveScriptStep(
        serial: String,
        name: String,
        step: ScriptStep,
    ): Boolean

    suspend fun saveRect(
        serial: String,
        rectangle: CvRectangle?,
    ): Boolean

    suspend fun findText(
        serial: String,
        text: String,
        locale: String,
    ): ApiResponse<List<RectangleWithText>>

    suspend fun getScripts(): ApiResponse<List<String>>
    suspend fun getScriptInfo(name: String): ApiResponse<Script>
    suspend fun deleteScript(name: String): Boolean
    suspend fun runScript(serial: String, name: String): Boolean
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