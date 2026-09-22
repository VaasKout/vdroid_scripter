package com.vision.scripter.streaming.impl.blocks.menu.state

data class MenuState(
    val serial: String = "",
    val type: MenuType = MenuType.Usual(),
)

sealed interface MenuType {
    data class Usual(
        val rectsShown: Boolean = false,
        val rectsAreLoading: Boolean = false,
        val scanShown: Boolean = false,
        val scanning: Boolean = false,
        val expanded: Boolean = false,
    ) : MenuType

    data class SelectingCV(
        val rectsLoading: Boolean = false,
    ) : MenuType

    data class CustomAction(
        val recording: Boolean = false,
    ) : MenuType
}
