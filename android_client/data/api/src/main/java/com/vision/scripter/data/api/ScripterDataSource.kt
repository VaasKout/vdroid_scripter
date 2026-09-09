package com.vision.scripter.data.api

import com.vision.scripter.data.api.models.AdbDevice
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.FoundLandmark
import com.vision.scripter.data.api.models.Library
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.Route
import com.vision.scripter.data.api.models.SessionStatus
import com.vision.scripter.data.api.models.Step
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.network.api.ApiResponse

interface ScripterDataSource {
    suspend fun getDevices(): ApiResponse<List<AdbDevice>>
    suspend fun pingServer(): Boolean

    suspend fun startSession(serial: String): ApiResponse<StreamingData>
    suspend fun getSessionStatus(serial: String): ApiResponse<SessionStatus>
    suspend fun closeSession(serial: String): Boolean

    suspend fun getLibrary(): ApiResponse<Library>

    suspend fun saveImage(
        serial: String,
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

    suspend fun getRoutes(): ApiResponse<List<String>>
    suspend fun getRoute(name: String): ApiResponse<Route>
    suspend fun deleteRoute(name: String): Boolean
    suspend fun runRoute(serial: String, name: String): ApiResponse<Unit>
    suspend fun queueSteps(serial: String, steps: List<Step>): ApiResponse<Unit>

    suspend fun scan(
        serial: String,
        images: List<String>,
        locale: String,
    ): ApiResponse<List<FoundLandmark>>

    suspend fun getRectangles(serial: String): ApiResponse<List<CvRectangle>>

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
