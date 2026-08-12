package com.vision.scripter.streaming.impl.screen.state

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

enum class KeyboardMode {
    EDIT,
    ADD_NEW,
}

fun KeyboardMode.increment(): KeyboardMode = when (this) {
    KeyboardMode.EDIT -> KeyboardMode.ADD_NEW
    else -> KeyboardMode.EDIT
}

data class StreamingState(
    val loading: Boolean = true,
    val isError: Boolean = false,
)
