package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.DEFAULT_TIMEOUT
import com.vision.scripter.streaming.impl.screen.state.KeyboardMode

data class MenuState(
    val serial: String = "",
    val type: MenuType = MenuType.Usual(),
)

sealed interface MenuType {
    data class Usual(
        val localCvMode: CVMode = CVMode.NO_CV,
        val textHighlighted: Boolean = false,
        val keyboardHighlighted: Boolean = false,
        val expanded: Boolean = false,
    ) : MenuType

    data class Recording(
        val controlRecording: Boolean = false,
        val recordTimeout: Int = DEFAULT_TIMEOUT,
    ) : MenuType

    data class SelectingCV(
        val localCvMode: CVMode = CVMode.CV_RECTS,
    ) : MenuType

    data object SelectingText : MenuType

    data class Keyboard(
        val isLoading: Boolean = true,
        val mode: KeyboardMode = KeyboardMode.TYPING,
        val fromUsual: Boolean = false,
    ) : MenuType
}

