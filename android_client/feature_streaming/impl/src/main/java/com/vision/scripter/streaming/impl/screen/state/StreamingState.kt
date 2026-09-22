package com.vision.scripter.streaming.impl.screen.state

const val ENG = "eng"
const val RUS = "rus"

val locales = listOf(
    ENG,
    RUS,
)

data class StreamingState(
    val loading: Boolean = true,
    val isError: Boolean = false,
)
