package com.vision.scripter.streaming.impl.state

import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.data.api.models.StepEvent
import com.vision.scripter.data.api.models.StreamingData
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
        val templateSelectMode: CvSelectMode = CvSelectMode.NONE,
        val textSelectMode: CvSelectMode = CvSelectMode.NONE,
        val typeText: Boolean = false,
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
        val templateSelectMode: CvSelectMode = CvSelectMode.NONE,
        val textSelectMode: CvSelectMode = CvSelectMode.NONE,
        val typeText: Boolean = false,
    ) : MenuState

    data class SelectingCV(
        val selectMode: CvSelectMode = CvSelectMode.APPLY_EVENT,
    ) : MenuState

    data class SelectingText(
        val selectMode: CvSelectMode = CvSelectMode.APPLY_EVENT,
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

enum class CvSelectMode {
    NONE,
    APPLY_EVENT,
    VISIBLE,
}

fun CvSelectMode.increment(): CvSelectMode {
    return when (this) {
        CvSelectMode.NONE -> CvSelectMode.APPLY_EVENT
        CvSelectMode.APPLY_EVENT -> CvSelectMode.VISIBLE
        else -> CvSelectMode.NONE
    }
}

fun CvSelectMode.incrementOnlyActive(): CvSelectMode {
    return when (this) {
        CvSelectMode.APPLY_EVENT -> CvSelectMode.VISIBLE
        CvSelectMode.VISIBLE -> CvSelectMode.APPLY_EVENT
        else -> CvSelectMode.APPLY_EVENT
    }
}
