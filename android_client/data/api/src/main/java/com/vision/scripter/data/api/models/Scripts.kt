package com.vision.scripter.data.api.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class Script(
    @SerialName("name")
    val name: String = "",
    @SerialName("node")
    val node: String = "",
    @SerialName("next_node")
    val nextNode: String = "",
    @SerialName("params")
    val params: List<Parameter> = listOf(),
    @SerialName("events")
    val events: List<Event> = listOf(),
    @SerialName("timeout")
    val timeout: Int = 0,
)

@Serializable
data class Parameter(
    @SerialName("type")
    val type: String = "",
    @SerialName("value")
    val value: String = "",
    @SerialName("locale")
    val locale: String = "",
)

fun Script.isEmpty(): Boolean {
    return params.isEmpty() && events.isEmpty()
}

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
