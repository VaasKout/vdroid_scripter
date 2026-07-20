package com.vision.scripter.streaming.impl.blocks.menu.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuState

@Immutable
data class MenuUiState(
    val menuState: MenuState = MenuState.Usual(),
    val dialogState: DialogState = DialogState.None,
)