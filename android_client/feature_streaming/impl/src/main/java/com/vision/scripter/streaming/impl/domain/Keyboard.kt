package com.vision.scripter.streaming.impl.domain

import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.streaming.impl.screen.state.KeyboardState

data class Keyboard(
    val locale: String = "",
    val buttons: List<RectangleWithText> = listOf(),
    val mode: KeyboardState = KeyboardState.TYPING,
    val typedText: String = "",
)