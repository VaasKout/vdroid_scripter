package com.vision.scripter.data.api

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.network.api.ApiResponse

interface ScripterRepository {
    suspend fun getDevices(): ApiResponse<List<AdbDevice>>
    suspend fun getDevicePreview(serial: String): ApiResponse<ByteArray>
    suspend fun pingServer(): Boolean

    suspend fun startSockets(serial: String): ApiResponse<StreamingData>

    suspend fun saveRect(
        serial: String,
        node: String,
        name: String,
        value: String,
        rectangle: CvRectangle?,
    ): Boolean

    suspend fun findText(
        serial: String,
        text: String,
        locale: String,
    ): ApiResponse<List<RectangleWithText>>

    suspend fun getNodes(): ApiResponse<List<String>>
    suspend fun getNodeScripts(node: String): ApiResponse<List<String>>
    suspend fun getScriptInfo(node: String, name: String): ApiResponse<Script>
    suspend fun deleteNode(node: String): Boolean
    suspend fun saveScript(script: Script): Boolean
    suspend fun deleteScript(node: String, name: String): Boolean
    suspend fun runScript(serial: String, node: String, name: String): Boolean
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