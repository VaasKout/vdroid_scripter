package com.vision.scripter.streaming.impl.state

import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.EVENT_ON_TEMPLATE
import com.vision.scripter.data.api.models.EVENT_ON_TEXT
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.StepEvent
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.data.api.models.TEMPLATE_IS_VISIBLE
import com.vision.scripter.data.api.models.TEXT_IS_VISIBLE
import com.vision.scripter.streaming.impl.video.VideoCodec
import com.vision.scripter.ui.states.LoadingState

const val ENG = "eng"
const val RUS = "rus"
const val NUMBERS = "numbers"
const val PHONE = "phone"
const val SPACE_KEY = "space"

val locales = listOf(
    ENG,
    RUS,
)

val keyboardLocales = listOf(
    ENG,
    RUS,
    NUMBERS,
    PHONE,
)

data class StreamingState(
    val serial: String = "",
    val loadingState: LoadingState = LoadingState.LoadingOnStart,

    val connectionEstablished: Boolean = false,
    val streamingHost: String = "",
    val videoCodec: VideoCodec = VideoCodec.H264,
    val streamingData: StreamingData? = null,
    val cvRectangles: List<CvRectangle> = listOf(),
    val selectedRectangles: List<CvRectangle> = listOf(),

    val record: Record = Record(),
    val keyboard: Keyboard = Keyboard(),
) {
    data class Record(
        val recordName: String = "",
        val stepEvents: List<StepEvent> = listOf(),
        val text: String = "",
        val locale: String = "",
        val flags: Int = 0,
    ) {
        fun clearStep() = Record(recordName = recordName)
    }

    data class Keyboard(
        val buttons: List<RectangleWithText> = listOf(),
    )
}

sealed interface MenuState {
    data class Usual(
        val cvMode: CVMode = CVMode.NO_CV,
        val textHighlighted: Boolean = false,
        val expanded: Boolean = false,
    ) : MenuState

    data class Recording(
        val controlRecording: Boolean = false,
        val flags: Int = 0,
    ) : MenuState

    data class SelectingCV(
        val flags: Int = EVENT_ON_TEMPLATE,
    ) : MenuState

    data class SelectingText(
        val flags: Int = EVENT_ON_TEXT,
        val text: String = "",
        val locale: String = "",
    ) : MenuState

    data class Keyboard(
        val isLoadingKeyboard: Boolean = true,
        val recordingKeyboard: Boolean = false,
        val typeText: String = "",
        val fromUsual: Boolean = false,
        val oldKey: String = "",
        val editing: Boolean = false,
        val showCvRectangles: Boolean = false,
    ) : MenuState
}

enum class CVMode(val value: Int) {
    NO_CV(0),
    CV_RECTS(1),
}

fun CVMode.increment(): CVMode {
    val newValue = this.value + 1
    val newMode = CVMode.entries.firstOrNull { it.value == newValue }
    return newMode ?: CVMode.NO_CV
}

fun Int.hasFlag(flag: Int): Boolean = this and flag != 0

fun Int.withFlag(flag: Int, enabled: Boolean): Int =
    if (enabled) this or flag else this and flag.inv()

fun Int.templateFlag(): Int = this and (EVENT_ON_TEMPLATE or TEMPLATE_IS_VISIBLE)

fun Int.textFlag(): Int = this and (EVENT_ON_TEXT or TEXT_IS_VISIBLE)

fun Int.combineTemplate(template: Int): Int =
    (this and (EVENT_ON_TEMPLATE or TEMPLATE_IS_VISIBLE).inv()) or template

fun Int.combineText(text: Int): Int =
    (this and (EVENT_ON_TEXT or TEXT_IS_VISIBLE).inv()) or text

fun Int.nextTemplate(): Int = when (this) {
    EVENT_ON_TEMPLATE -> TEMPLATE_IS_VISIBLE
    TEMPLATE_IS_VISIBLE -> 0
    else -> EVENT_ON_TEMPLATE
}

fun Int.nextTemplateActive(): Int = when (this) {
    EVENT_ON_TEMPLATE -> TEMPLATE_IS_VISIBLE
    TEMPLATE_IS_VISIBLE -> EVENT_ON_TEMPLATE
    else -> EVENT_ON_TEMPLATE
}

fun Int.nextText(): Int = when (this) {
    EVENT_ON_TEXT -> TEXT_IS_VISIBLE
    TEXT_IS_VISIBLE -> 0
    else -> EVENT_ON_TEXT
}
