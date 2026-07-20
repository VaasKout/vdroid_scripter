package com.vision.scripter.streaming.impl.screen.main.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.streaming.impl.screen.main.commandobservers.StreamingSharedEvent
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiCommand
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.CommandFlow
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

    private val currentState: StreamingState
        get() = _stateFlow.value

    override val uiStateFlow: StateFlow<StreamingUiState> =
        stateFlow.map(uiStateMapper::map).stateIn(
            coroutineScope, SharingStarted.Eagerly, StreamingUiState(),
        )

    override val uiCommandsFlow: CommandFlow<StreamingUiCommand> = CommandFlow(coroutineScope)
    override val sharedEventsFlow: CommandFlow<StreamingSharedEvent> = CommandFlow(coroutineScope)

    override fun initArgs(serial: String) {
        sharedEventsFlow.tryEmit(StreamingSharedEvent.InitArgs(serial))
    }

    override fun onLoadData(onStart: Boolean) {
        sharedEventsFlow.tryEmit(StreamingSharedEvent.LoadData(onStart))
    }

    override fun showStepSavedSnackbar() {
        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowScriptSavedSnackbar)
    }

    override fun showNetworkError() {
        uiCommandsFlow.tryEmit(StreamingUiCommand.ShowNetworkError)
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
