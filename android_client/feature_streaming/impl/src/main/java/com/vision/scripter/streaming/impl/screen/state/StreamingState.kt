package com.vision.scripter.streaming.impl.screen.state

const val ENG = "eng"
const val RUS = "rus"
const val NUMBERS = "numbers"
const val PHONE = "phone"
const val SPACE_KEY = "space"

const val DEFAULT_TIMEOUT = 15

const val NEW_SCRIPTS_LOCATION = "new_scripts"

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

enum class KeyboardMode {
    TYPING,
    EDIT,
    ADD_NEW,
}

fun KeyboardMode.increment(): KeyboardMode = when (this) {
    KeyboardMode.TYPING -> KeyboardMode.EDIT
    KeyboardMode.EDIT -> KeyboardMode.ADD_NEW
    else -> KeyboardMode.TYPING
}

data class StreamingState(
    val loading: Boolean = true,
    val isError: Boolean = false,
)
