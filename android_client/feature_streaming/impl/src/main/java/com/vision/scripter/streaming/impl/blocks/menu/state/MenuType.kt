package com.vision.scripter.streaming.impl.blocks.menu.state

import com.vision.scripter.streaming.impl.screen.state.KeyboardMode

data class MenuState(
    val serial: String = "",
    val type: MenuType = MenuType.Usual(),
)

sealed interface MenuType {
    data class Usual(
        val rectsShown: Boolean = false,
        val scanShown: Boolean = false,
        val scanning: Boolean = false,
        val expanded: Boolean = false,
    ) : MenuType

    data object SelectingCV : MenuType

    data class CustomAction(
        val recording: Boolean = false,
    ) : MenuType

    data class Keyboard(
        val isLoading: Boolean = true,
        val mode: KeyboardMode = KeyboardMode.EDIT,
    ) : MenuType
}
