package com.vision.scripter.streaming.impl.screen.main.ui

import androidx.compose.runtime.Stable
import com.vision.scripter.streaming.impl.screen.main.commandobservers.StreamingSharedEvent
import com.vision.scripter.ui.CommandFlow
import kotlinx.coroutines.flow.StateFlow

@Stable
interface StreamingUiStateHolder {
    val uiStateFlow: StateFlow<StreamingUiState?>
    val uiCommandsFlow: CommandFlow<StreamingUiCommand>
    val sharedEventsFlow: CommandFlow<StreamingSharedEvent>

    fun initArgs(serial: String)
    fun onLoadData(onStart: Boolean)

    fun showNetworkError()
    fun showStepSavedSnackbar()
}
