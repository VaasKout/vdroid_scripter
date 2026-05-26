package com.vision.scripter.streaming.impl.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.streaming.impl.state.DialogState
import com.vision.scripter.streaming.impl.state.MenuState

@Immutable
data class StreamingUiState(
    val isLoading: Boolean = true,
    val dialogState: DialogState = DialogState.NONE,
    val hasConnection: Boolean = false,
    val rectangles: List<CvRectangle> = listOf(),
    val selectedRectangles: List<CvRectangle> = listOf(),
    val keyboardButtons: List<RectangleWithText> = listOf(),
    val menuState: MenuState = MenuState.Usual(),
)
