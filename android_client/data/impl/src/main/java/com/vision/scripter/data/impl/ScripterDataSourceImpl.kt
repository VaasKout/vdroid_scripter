package com.vision.scripter.data.impl

import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.AdbDevicesResponse
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.EditKeyboardRequest
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.KeyboardButtons
import com.vision.scripter.data.api.models.Library
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.SaveActionRequest
import com.vision.scripter.data.api.models.SaveImageRequest
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.data.api.models.isEmpty
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.network.api.NetworkClient
import com.vision.scripter.network.api.NetworkError
import kotlinx.serialization.json.Json
import java.net.URLEncoder
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ScripterDataSourceImpl @Inject constructor(
    private val networkClient: NetworkClient,
) : ScripterDataSource {

    override suspend fun getDevices(): ApiResponse<List<AdbDevice>> {
        return when (val result = networkClient.get("devices")) {
            is ApiResponse.Success -> {
                val json = result.data
                val devices = if (json.isEmpty()) listOf()
                else Json.decodeFromString<AdbDevicesResponse>(result.data).devices
                ApiResponse.Success(devices)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun getDevicePreview(serial: String): ApiResponse<ByteArray> {
        return when (val result = networkClient.getMultipart("preview/$serial")) {
            is ApiResponse.Success -> {
                val bytesArray = result.data
                if (bytesArray.isEmpty()) {
                    val networkError = NetworkError.ServerError("no images")
                    ApiResponse.Error(networkError)
                } else ApiResponse.Success(result.data.first())
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun pingServer(): Boolean {
        val result = networkClient.get("ping")
        return result is ApiResponse.Success
    }

    override suspend fun startSession(serial: String): ApiResponse<StreamingData> {
        return when (val result = networkClient.post("devices/$serial/session", "")) {
            is ApiResponse.Success -> {
                val json = result.data
                val streamingData = if (json.isEmpty()) StreamingData()
                else Json.decodeFromString<StreamingData>(result.data)
                ApiResponse.Success(streamingData)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun getLibrary(): ApiResponse<Library> {
        return when (val result = networkClient.get("library")) {
            is ApiResponse.Success -> {
                val json = result.data
                val library = if (json.isEmpty()) Library()
                else Json.decodeFromString<Library>(result.data)
                ApiResponse.Success(library)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun saveImage(
        serial: String,
        rectangle: CvRectangle?,
    ): Boolean {
        if (serial.isEmpty() || rectangle.isEmpty() || rectangle?.label.isNullOrEmpty()) {
            return false
        }
        val request = SaveImageRequest(
            serial = serial,
            rectangle = rectangle,
        )
        val body = Json.encodeToString(request)
        val result = networkClient.post("save_image", body)
        return result is ApiResponse.Success
    }

    override suspend fun saveAction(
        name: String,
        screenWidth: Int,
        screenHeight: Int,
        events: List<Event>,
    ): Boolean {
        if (name.isEmpty() || events.isEmpty()) return false
        val request = SaveActionRequest(
            name = name,
            screenWidth = screenWidth,
            screenHeight = screenHeight,
            events = events,
        )
        val body = Json.encodeToString(request)
        val result = networkClient.post("save_action", body)
        return result is ApiResponse.Success
    }

    override suspend fun deleteImage(name: String): Boolean {
        if (name.isEmpty()) return false
        val result = networkClient.delete("images/${encodePath(name)}")
        return result is ApiResponse.Success
    }

    override suspend fun deleteAction(name: String): Boolean {
        if (name.isEmpty()) return false
        val result = networkClient.delete("actions/${encodePath(name)}")
        return result is ApiResponse.Success
    }

    override suspend fun findText(
        serial: String,
        text: String,
        locale: String,
    ): ApiResponse<List<RectangleWithText>> {
        return when (val result = networkClient.get(
            "/devices/$serial/find_text?text=${encodeQuery(text)}&locale=$locale",
        )) {
            is ApiResponse.Success -> {
                val json = result.data
                val ocrData = if (json.isEmpty()) listOf()
                else Json.decodeFromString<List<RectangleWithText>>(result.data)
                ApiResponse.Success(ocrData)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun resetKeyboard(
        serial: String,
        locale: String
    ): ApiResponse<List<RectangleWithText>> {
        return when (val result =
            networkClient.get("/devices/$serial/reset_keyboard?locale=$locale")) {
            is ApiResponse.Success -> {
                val json = result.data
                val keyboardButtons = if (json.isEmpty()) KeyboardButtons()
                else Json.decodeFromString<KeyboardButtons>(result.data)
                ApiResponse.Success(keyboardButtons.buttons)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun getKeyboard(
        serial: String,
        locale: String
    ): ApiResponse<List<RectangleWithText>> {
        return when (val result =
            networkClient.get("/devices/$serial/keyboard?locale=$locale")) {
            is ApiResponse.Success -> {
                val json = result.data
                val keyboardButtons = if (json.isEmpty()) KeyboardButtons()
                else Json.decodeFromString<KeyboardButtons>(result.data)
                ApiResponse.Success(keyboardButtons.buttons)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun editKeyboard(
        serial: String,
        locale: String,
        name: String,
        rectangle: CvRectangle?,
    ): Boolean {
        if (serial.isEmpty() || name.isEmpty() || rectangle.isEmpty()) return false
        val request = EditKeyboardRequest(
            serial = serial,
            locale = locale,
            name = name,
            rectangle = rectangle,
        )
        val body = Json.encodeToString(request)
        val result = networkClient.post("/devices/$serial/edit_keyboard", body)
        return result is ApiResponse.Success
    }

    override suspend fun deleteButton(serial: String, locale: String, name: String): Boolean {
        if (serial.isEmpty() || name.isEmpty()) return false
        val result = networkClient.get(
            "/devices/$serial/delete_button?locale=$locale&name=${encodeQuery(name)}",
        )
        return result is ApiResponse.Success
    }

    private fun encodeQuery(value: String): String = URLEncoder.encode(value, "UTF-8")

    private fun encodePath(value: String): String = encodeQuery(value).replace("+", "%20")
}