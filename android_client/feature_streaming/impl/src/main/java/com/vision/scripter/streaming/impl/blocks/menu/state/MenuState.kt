package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.streaming.impl.screen.state.CVMode
import com.vision.scripter.streaming.impl.screen.state.DEFAULT_TIMEOUT
import com.vision.scripter.streaming.impl.screen.state.KeyboardState

sealed interface MenuState {
    data class Usual(
        val localCvMode: CVMode = CVMode.NO_CV,
        val textHighlighted: Boolean = false,
        val keyboardHighlighted: Boolean = false,
        val expanded: Boolean = false,
    ) : MenuState

    data class Recording(
        val controlRecording: Boolean = false,
        val recordTimeout: Int = DEFAULT_TIMEOUT,
    ) : MenuState

    data class SelectingCV(
        val localCvMode: CVMode = CVMode.CV_RECTS,
    ) : MenuState

    data object SelectingText : MenuState

    data class Keyboard(
        val isLoading: Boolean = true,
        val mode: KeyboardState = KeyboardState.TYPING,
        val fromUsual: Boolean = false,
    ) : MenuState
}

