package com.vision.scripter.streaming.impl.blocks.menu.state

sealed interface DialogState {
    data object None : DialogState
    data object AddItem : DialogState
    data object Text : DialogState
    data object Keyboard : DialogState
    data class EditKeyboard(val oldKey: String) : DialogState
}
