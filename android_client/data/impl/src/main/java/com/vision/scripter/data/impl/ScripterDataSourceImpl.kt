package com.vision.scripter.data.impl

import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.AdbDevicesResponse
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.FoundLandmark
import com.vision.scripter.data.api.models.LandmarksResponse
import com.vision.scripter.data.api.models.Library
import com.vision.scripter.data.api.models.RectanglesResponse
import com.vision.scripter.data.api.models.Route
import com.vision.scripter.data.api.models.RoutesResponse
import com.vision.scripter.data.api.models.SaveActionRequest
import com.vision.scripter.data.api.models.SaveImageRequest
import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.data.api.models.SessionStatusResponse
import com.vision.scripter.data.api.models.Step
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.data.api.models.isEmpty
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.network.api.NetworkClient
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

    override suspend fun getSessionStatus(serial: String): ApiResponse<SessionStatus> {
        return networkClient.get("devices/$serial/session").decode(SessionStatus.Closed) {
            SessionStatus.parse(Json.decodeFromString<SessionStatusResponse>(it).status)
        }
    }

    override suspend fun closeSession(serial: String): Boolean {
        if (serial.isEmpty()) return false
        val result = networkClient.delete("devices/$serial/session")
        return result is ApiResponse.Success
    }

    override suspend fun getRoutes(): ApiResponse<List<String>> {
        return networkClient.get("routes").decode(listOf()) {
            Json.decodeFromString<RoutesResponse>(it).routes
        }
    }

    override suspend fun getRoute(name: String): ApiResponse<Route> {
        return networkClient.get("routes/${encodePath(name)}").decode(Route()) {
            Json.decodeFromString<Route>(it)
        }
    }

    override suspend fun deleteRoute(name: String): Boolean {
        if (name.isEmpty()) return false
        val result = networkClient.delete("routes/${encodePath(name)}")
        return result is ApiResponse.Success
    }

    override suspend fun runRoute(serial: String, name: String): ApiResponse<Unit> {
        val path = "run_route?serial=${encodeQuery(serial)}&name=${encodeQuery(name)}"
        return networkClient.get(path).ignoreBody()
    }

    override suspend fun queueSteps(serial: String, steps: List<Step>): ApiResponse<Unit> {
        val body = Json.encodeToString(steps)
        return networkClient.post("devices/$serial/queue_steps", body).ignoreBody()
    }

    override suspend fun scan(
        serial: String,
        images: List<String>,
        locale: String,
    ): ApiResponse<List<FoundLandmark>> {
        val path = buildString {
            append("devices/$serial/scan?locale=${encodeQuery(locale)}")
            if (images.isNotEmpty()) {
                append("&images=${encodeQuery(images.joinToString(","))}")
            }
        }
        return networkClient.get(path).decode(listOf()) {
            Json.decodeFromString<LandmarksResponse>(it).landmarks
        }
    }

    override suspend fun getRectangles(serial: String): ApiResponse<List<CvRectangle>> {
        return networkClient.get("devices/$serial/rectangles").decode(listOf()) {
            Json.decodeFromString<RectanglesResponse>(it).rectangles
        }
    }

    private inline fun <T> ApiResponse<String>.decode(
        empty: T,
        parse: (String) -> T,
    ): ApiResponse<T> {
        return when (this) {
            is ApiResponse.Success -> ApiResponse.Success(if (data.isEmpty()) empty else parse(data))
            is ApiResponse.Error -> this
        }
    }

    private fun ApiResponse<String>.ignoreBody(): ApiResponse<Unit> {
        return when (this) {
            is ApiResponse.Success -> ApiResponse.Success(Unit)
            is ApiResponse.Error -> this
        }
    }

    private fun encodeQuery(value: String): String = URLEncoder.encode(value, "UTF-8")

    private fun encodePath(value: String): String = encodeQuery(value).replace("+", "%20")
}