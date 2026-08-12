package com.vision.scripter.data.api.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Library(
    @SerialName("images")
    val images: List<String> = listOf(),
    @SerialName("actions")
    val actions: List<String> = listOf(),
)

@Serializable
data class SaveImageRequest(
    @SerialName("serial")
    val serial: String = "",
    @SerialName("name")
    val name: String = "",
    @SerialName("rectangle")
    val rectangle: CvRectangle? = null,
)

@Serializable
data class SaveActionRequest(
    @SerialName("name")
    val name: String = "",
    @SerialName("screen_width")
    val screenWidth: Int = 0,
    @SerialName("screen_height")
    val screenHeight: Int = 0,
    @SerialName("events")
    val events: List<Event> = listOf(),
)

@Serializable
data class Event(
    @SerialName("time")
    val time: Long = 0L,
    @SerialName("data")
    val data: UByteArray? = null,
)

@Serializable
data class KeyboardButtons(
    @SerialName("buttons")
    val buttons: List<RectangleWithText> = listOf(),
)
