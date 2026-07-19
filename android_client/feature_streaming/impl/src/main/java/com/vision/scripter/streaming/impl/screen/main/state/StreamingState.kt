package com.vision.scripter.streaming.impl.screen.main.state

import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.Event
import com.vision.scripter.data.api.models.Parameter
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.StreamingData
import com.vision.scripter.streaming.impl.blocks.video.VideoCodec
import com.vision.scripter.ui.states.LoadingState

const val ENG = "eng"
const val RUS = "rus"
const val NUMBERS = "numbers"
const val PHONE = "phone"
const val SPACE_KEY = "space"

const val DEFAULT_TIMEOUT = 15

const val NEW_SCRIPTS_NODE = "new_scripts"

const val TEMPLATE = "template"
const val YOLO_CLASS = "yolo_class"
const val TEXT = "text"
const val TYPE_TEXT = "type_text"
const val COMMAND = "command"

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
        val node: String = NEW_SCRIPTS_NODE,
        val params: List<Parameter> = listOf(),
        val events: List<Event> = listOf(),
        val locale: String = "",
        val timeout: Int = DEFAULT_TIMEOUT,
    ) {
        fun clear() = Record()
    }

    data class Keyboard(
        val buttons: List<RectangleWithText> = listOf(),
    )
}

sealed interface MenuState {
    data class Usual(
        val localCvMode: CVMode = CVMode.NO_CV,
        val textHighlighted: Boolean = false,
        val keyboardHighlighted: Boolean = false,
        val expanded: Boolean = false,
    ) : MenuState

    data class Recording(
        val controlRecording: Boolean = false,
        val localCvMode: CVMode = CVMode.NO_CV,
        val textHighlighted: Boolean = false,
        val keyboardHighlighted: Boolean = false,
        val customTimeout: Boolean = false,
        val recordTimeout: Int = DEFAULT_TIMEOUT,
    ) : MenuState

    data class SelectingCV(
        val cvMode: CVMode = CVMode.CV_RECTS,
    ) : MenuState

    data class SelectingText(
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
    YOLO(2),
}

fun CVMode.increment(): CVMode {
    val newValue = this.value + 1
    val newMode = CVMode.entries.firstOrNull { it.value == newValue }
    return newMode ?: CVMode.NO_CV
}

fun CVMode.toggleDetection(): CVMode = when (this) {
    CVMode.CV_RECTS -> CVMode.YOLO
    else -> CVMode.CV_RECTS
}

fun CVMode.toType(): String = when (this) {
    CVMode.CV_RECTS -> TEMPLATE
    CVMode.YOLO -> YOLO_CLASS
    else -> ""
}
