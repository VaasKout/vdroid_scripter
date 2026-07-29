package com.vision.scripter.streaming.impl.screen.main.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.streaming.impl.screen.main.commandobservers.ScreenToVideo
import com.vision.scripter.streaming.impl.screen.main.commandobservers.StreamingSharedEvent
import com.vision.scripter.streaming.impl.screen.main.commandobservers.VideoToScreen
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiCommand
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.CommandFlow2
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import javax.inject.Inject

@ViewModelScoped
class StreamingInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val uiStateMapper: StreamingUiStateMapper,
) : StreamingUiStateHolder {

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("streaming_interactor")

    private val _stateFlow = MutableStateFlow(StreamingState())
    private val stateFlow: SharedFlow<StreamingState> = _stateFlow.asStateFlow()

    override val uiStateFlow: StateFlow<StreamingUiState> =
        stateFlow.map(uiStateMapper::map).stateIn(
            coroutineScope, SharingStarted.Eagerly, StreamingUiState(),
        )

    override val uiCommandsFlow: CommandFlow<StreamingUiCommand> = CommandFlow(coroutineScope)
    override val sharedCommandsFlow: CommandFlow2<StreamingSharedEvent> = CommandFlow2()

    override fun onSharedEvent(event: VideoToScreen) {
        when (event) {
            is VideoToScreen.ShowNetworkError -> showNetworkError()
            is VideoToScreen.ShowScriptSavedSnackbar -> showStepSavedSnackbar()
            is VideoToScreen.SuccessLoading -> doneLoading()
        }
    }

    override fun onRefresh() {
        _stateFlow.update { it.copy(loading = true, isError = false) }
        sharedCommandsFlow.tryEmit(ScreenToVideo.Refresh)
    }

    private fun showStepSavedSnackbar() {
        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowScriptSavedSnackbar)
    }

    private fun showNetworkError() {
        _stateFlow.update { it.copy(loading = false, isError = true) }
        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
    }

    private fun doneLoading() {
        _stateFlow.update { it.copy(loading = false, isError = false) }
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
