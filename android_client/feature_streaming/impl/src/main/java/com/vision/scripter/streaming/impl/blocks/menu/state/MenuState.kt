package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.streaming.impl.screen.main.state.DEFAULT_TIMEOUT

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

    data object SelectingText: MenuState

    data class Keyboard(
        val isLoadingKeyboard: Boolean = true,
        val recordingKeyboard: Boolean = false,
        val fromUsual: Boolean = false,
        val editing: Boolean = false,
        val showCvRectangles: Boolean = false,
    ) : MenuState
}

