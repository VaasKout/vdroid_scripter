package com.vision.scripter.streaming.impl.blocks.video.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.commandobservers.StreamingSharedEvent
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
fun VideoSharedCommandObserver(
    uiStateHolder: VideoUiStateHolder,
    sharedStateHolder: StreamingUiStateHolder,
) {
    sharedStateHolder.sharedEventsFlow.observe {
        when (it) {
            is StreamingSharedEvent.LoadData -> {
                uiStateHolder.onLoadData(it.onStart)
            }

            is StreamingSharedEvent.InitArgs -> {
                uiStateHolder.initArgs(it.serial)
            }

            is StreamingSharedEvent.EditKeyboardButton -> {

            }

            is StreamingSharedEvent.FindText -> {

            }

            is StreamingSharedEvent.KeyboardInit -> {

            }

            is StreamingSharedEvent.NextCvMode -> {

            }

            is StreamingSharedEvent.RecordCancelled -> {

            }

            is StreamingSharedEvent.SaveClicked -> {

            }

            is StreamingSharedEvent.SaveLocale -> {

            }

            is StreamingSharedEvent.SaveRecordName -> {

            }

            is StreamingSharedEvent.TimeoutSaved -> {

            }
        }
    }
}