package com.vision.scripter.streaming.impl.screen.main.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.screen.main.state.DEFAULT_TIMEOUT
import com.vision.scripter.streaming.impl.screen.main.state.MenuState

@Immutable
data class StreamingUiState(
    val isLoading: Boolean = true,
    val dialogState: DialogState = DialogState.NONE,
    val hasConnection: Boolean = false,
    val rectangles: List<CvRectangle> = listOf(),
    val selectedRectangles: List<CvRectangle> = listOf(),
    val keyboardButtons: List<RectangleWithText> = listOf(),
    val menuState: MenuState = MenuState.Usual(),
    val recordTimeout: Int = DEFAULT_TIMEOUT,
)
