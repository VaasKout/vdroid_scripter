package com.vision.scripter.data.api.models

import android.view.MotionEvent.ACTION_DOWN
import android.view.MotionEvent.ACTION_UP
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class SaveStepRequest(
    @SerialName("serial")
    val serial: String,
    @SerialName("name")
    val name: String,
    @SerialName("step")
    val scriptStep: ScriptStep,
)

@Serializable
data class Script(
    @SerialName("name")
    val name: String = "",
    @SerialName("steps")
    val steps: List<ScriptStep> = listOf(),
)

const val EVENT_ON_TEMPLATE = 1 shl 0
const val EVENT_ON_TEXT = 1 shl 1
const val TYPE_TEXT = 1 shl 2
const val TEMPLATE_IS_VISIBLE = 1 shl 3
const val TEXT_IS_VISIBLE = 1 shl 4

@Serializable
data class ScriptStep(
    @SerialName("events")
    val events: List<StepEvent> = listOf(),
    @SerialName("flags")
    val flags: Int = 0,
    @SerialName("text")
    val text: String = "",
    @SerialName("locale")
    val locale: String = "",
    @SerialName("command")
    val command: String = "",
)

fun ScriptStep.isEmpty(): Boolean {
    return events.isEmpty() && flags == 0 && text.isEmpty() && command.isEmpty()
}

@Serializable
data class StepEvent(
    @SerialName("time")
    val time: Long = 0L,
    @SerialName("data")
    val data: ByteArray? = null,
)

fun List<StepEvent>.extractPressEvent(): List<StepEvent> {
    val pressAction = mutableListOf<StepEvent>()
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