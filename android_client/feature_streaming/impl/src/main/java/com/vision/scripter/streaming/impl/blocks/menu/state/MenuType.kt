package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.streaming.impl.screen.state.CVMode
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

    data class SelectingCV(
        val localCvMode: CVMode = CVMode.CV_RECTS,
    ) : MenuType

    data class CustomAction(
        val recording: Boolean = false,
    ) : MenuType

    data class Keyboard(
        val isLoading: Boolean = true,
        val mode: KeyboardMode = KeyboardMode.EDIT,
    ) : MenuType
}
