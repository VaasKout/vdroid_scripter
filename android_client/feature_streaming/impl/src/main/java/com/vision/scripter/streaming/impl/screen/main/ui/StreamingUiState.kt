package com.vision.scripter.streaming.impl.screen.main.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.screen.main.state.MenuState
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class StreamingUiState(
    val isLoading: Boolean = true,
    val dialogState: DialogState = DialogState.NONE,
    val hasConnection: Boolean = false,
    val rectangles: ImmutableList<CvRectangle> = persistentListOf(),
    val selectedRectangles: ImmutableList<CvRectangle> = persistentListOf(),
    val keyboardButtons: ImmutableList<RectangleWithText> = persistentListOf(),
    val menuState: MenuState = MenuState.Usual(),
)
