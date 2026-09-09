package com.vision.scripter.data.api.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Route(
    @SerialName("name")
    val name: String = "",
    @SerialName("prompt")
    val prompt: String = "",
    @SerialName("steps")
    val steps: List<Step> = listOf(),
)

@Serializable
data class Step(
    @SerialName("id")
    val id: Int = 0,
    @SerialName("event")
    val event: String = "",
    @SerialName("landmarks")
    val landmarks: List<Landmark> = listOf(),
    @SerialName("timeout")
    val timeout: Int = 0,
    @SerialName("delay")
    val delay: Int = 0,
)

@Serializable
data class Landmark(
    @SerialName("type")
    val type: String = "",
    @SerialName("value")
    val value: String = "",
    @SerialName("locale")
    val locale: String = "",
)

@Serializable
data class RoutesResponse(
    @SerialName("routes")
    val routes: List<String> = listOf(),
)

@Serializable
data class FoundLandmark(
    @SerialName("type")
    val type: String = "",
    @SerialName("value")
    val value: String = "",
    @SerialName("locale")
    val locale: String = "",
    @SerialName("rectangle")
    val rectangle: CvRectangle = CvRectangle(leftX = 0, rightX = 0, topY = 0, bottomY = 0),
)

@Serializable
data class LandmarksResponse(
    @SerialName("landmarks")
    val landmarks: List<FoundLandmark> = listOf(),
)

@Serializable
data class RectanglesResponse(
    @SerialName("rectangles")
    val rectangles: List<CvRectangle> = listOf(),
)

@Serializable
data class SessionStatusResponse(
    @SerialName("status")
    val status: String = "",
)

sealed interface SessionStatus {
    val text: String

    data object Closed : SessionStatus {
        override val text: String = "closed"
    }

    data object Idle : SessionStatus {
        override val text: String = "idle"
    }

    data class Running(override val text: String) : SessionStatus

    data class Error(override val text: String) : SessionStatus

    companion object {
        fun parse(raw: String): SessionStatus {
            val status = raw.trim()
            if (status == Closed.text) return Closed
            if (status == Idle.text) return Idle
            if (status.startsWith("running")) return Running(status)
            return Error(status)
        }
    }
}
