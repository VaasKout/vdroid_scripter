package com.vision.scripter.data.impl

import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.AdbDevicesResponse
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.EditKeyboardRequest
import com.vision.scripter.data.api.models.KeyboardButtons
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.SaveRectRequest
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.data.api.models.isEmpty
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.network.api.NetworkClient
import com.vision.scripter.network.api.NetworkError
import kotlinx.serialization.json.Json
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ScripterRepositoryImpl @Inject constructor(
    private val networkClient: NetworkClient,
) : ScripterRepository {

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

    override suspend fun startSockets(serial: String): ApiResponse<StreamingData> {
        return when (val result = networkClient.get("start_sockets/$serial")) {
            is ApiResponse.Success -> {
                val json = result.data
                val streamingData = if (json.isEmpty()) StreamingData()
                else Json.decodeFromString<StreamingData>(result.data)
                ApiResponse.Success(streamingData)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun saveScript(script: Script): Boolean {
        if (script.node.isEmpty() || script.name.isEmpty() || script.isEmpty()) return false
        val body = Json.encodeToString(script)
        val result = networkClient.post("save_script", body)
        return result is ApiResponse.Success
    }

    override suspend fun saveRect(
        serial: String,
        rectangle: CvRectangle?,
    ): Boolean {
        if (rectangle.isEmpty()) return false
        val saveRectRequest = SaveRectRequest(
            serial = serial,
            rectangle = rectangle,
        )
        val body = Json.encodeToString(saveRectRequest)
        val result = networkClient.post("save_rectangle", body)
        return result is ApiResponse.Success
    }

    override suspend fun findText(
        serial: String,
        text: String,
        locale: String,
    ): ApiResponse<List<RectangleWithText>> {
        return when (val result = networkClient.get(
            "/devices/$serial/find_text?text=$text&locale=$locale",
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

    override suspend fun getNodes(): ApiResponse<List<String>> {
        return when (val result = networkClient.get("nodes")) {
            is ApiResponse.Success -> {
                val json = result.data
                val nodes = if (json.isEmpty()) listOf()
                else Json.decodeFromString<List<String>>(result.data)
                ApiResponse.Success(nodes)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun getNodeScripts(node: String): ApiResponse<List<String>> {
        return when (val result = networkClient.get("nodes/$node")) {
            is ApiResponse.Success -> {
                val json = result.data
                val scripts = if (json.isEmpty()) listOf()
                else Json.decodeFromString<List<String>>(result.data)
                ApiResponse.Success(scripts)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun getScriptInfo(node: String, name: String): ApiResponse<Script> {
        return when (val result = networkClient.get("nodes/$node/$name")) {
            is ApiResponse.Success -> {
                val json = result.data
                val script = if (json.isEmpty()) Script()
                else Json.decodeFromString<Script>(result.data)
                ApiResponse.Success(script)
            }

            is ApiResponse.Error -> result
        }
    }

    override suspend fun deleteNode(node: String): Boolean {
        return when (val result = networkClient.delete("nodes/$node")) {
            is ApiResponse.Success -> result.data.isNotEmpty()
            is ApiResponse.Error -> false
        }
    }

    override suspend fun deleteScript(node: String, name: String): Boolean {
        return when (val result = networkClient.delete("nodes/$node/$name")) {
            is ApiResponse.Success -> result.data.isNotEmpty()
            is ApiResponse.Error -> false
        }
    }

    override suspend fun runScript(serial: String, node: String, name: String): Boolean {
        return when (val result = networkClient.get("nodes/$node/$name/run?serial=$serial")) {
            is ApiResponse.Success -> result.data.isNotEmpty()
            is ApiResponse.Error -> false
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
        val result = networkClient.get("/devices/$serial/delete_button?locale=$locale&name=$name")
        return result is ApiResponse.Success
    }

}