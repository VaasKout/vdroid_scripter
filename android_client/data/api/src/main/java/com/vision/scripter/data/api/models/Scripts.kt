package com.vision.scripter.data.api.models

import android.view.MotionEvent.ACTION_DOWN
import android.view.MotionEvent.ACTION_UP
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
    @SerialName("id")
    val id: Int = 0,
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
    val data: ByteArray? = null,
)

fun List<Event>.extractPressEvent(): List<Event> {
    val pressAction = mutableListOf<Event>()
    forEachIndexed { index, event ->
        val data = event.data ?: return@forEachIndexed
        val nextData = if (index < this.size - 1) this[index + 1].data else null

        if (
            index < this.size - 1 && data.size == 32 && data[1].toInt() == ACTION_DOWN &&
            nextData != null && nextData[1].toInt() == ACTION_UP
        ) {
            pressAction.add(this[index])
            pressAction.add(this[index + 1])
            return pressAction
        }
    }
    return pressAction
}

@Serializable
data class KeyboardButtons(
    @SerialName("buttons")
    val buttons: List<RectangleWithText> = listOf(),
)
