package com.vision.scripter.streaming.impl.blocks.menu.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuType

@Immutable
data class MenuUiState(
    val menuType: MenuType = MenuType.Usual(),
    val dialogState: DialogState = DialogState.None,
)