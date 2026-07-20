package com.vision.scripter.streaming.impl.blocks.menu.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiCommand
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
fun MenuCommandObserver(
    uiStateHolder: MenuUiStateHolder,
    sharedStateHolder: StreamingUiStateHolder,
) {
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is MenuUiCommand.ExitCommand -> {
                sharedStateHolder.uiCommandsFlow.tryEmit(StreamingUiCommand.Exit)
            }

            is MenuUiCommand.NextCvMode -> {

            }

            is MenuUiCommand.FindText -> {

            }

            is MenuUiCommand.KeyboardInit -> {

            }

            is MenuUiCommand.EditKeyboardButton -> {

            }

            is MenuUiCommand.SaveParameter -> {

            }

            is MenuUiCommand.CancelRecording -> {

            }

            is MenuUiCommand.SaveTimeout -> {

            }

            is MenuUiCommand.SaveRecordName -> {

            }

            is MenuUiCommand.SaveLocale -> {

            }
        }
    }
}