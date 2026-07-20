package com.vision.scripter.streaming.impl.blocks.video.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiCommand
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
fun VideoCommandObserver(
    uiStateHolder: VideoUiStateHolder,
    sharedStateHolder: StreamingUiStateHolder,
) {
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is VideoUiCommand.ShowNetworkError -> {
                sharedStateHolder.showNetworkError()
            }

            is VideoUiCommand.ShowScriptSavedSnackbar -> {
                sharedStateHolder.showStepSavedSnackbar()
            }

            is VideoUiCommand.TextFound -> {

            }

            is VideoUiCommand.SetKeyboardLoading -> {

            }

            is VideoUiCommand.ScriptSaved -> {

            }

            is VideoUiCommand.SelectKeyboardKey -> {

            }
        }
    }
}