package com.vision.scripter.streaming.impl.blocks.menu.state

sealed class MenuUiCommand {
    data object ExitCommand : MenuUiCommand()
}