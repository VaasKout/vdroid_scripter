package com.vision.scripter.streaming.impl.blocks.menu.ui

sealed class MenuUiCommand {
    data object ExitCommand : MenuUiCommand()
}