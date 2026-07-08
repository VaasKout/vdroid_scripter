package com.vision.scripter.main.impl.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterRepository
import com.vision.scripter.main.impl.ui.MainUiCommand
import com.vision.scripter.main.impl.ui.MainUiState
import com.vision.scripter.main.impl.ui.MainUiStateHolder
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.map
import kotlinx.coroutines.flow.shareIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import javax.inject.Inject

@ViewModelScoped
class MainInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterRepository: ScripterRepository,
    private val uiStateMapper: MainUiStateMapper,
) : MainUiStateHolder {

    private val _stateFlow = MutableStateFlow(MainState())
    private val stateFlow: StateFlow<MainState> = _stateFlow.asStateFlow()
    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("main_interactor")

    override val uiStateFlow: SharedFlow<MainUiState>
        get() = stateFlow.map(uiStateMapper::map)
            .shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<MainUiCommand> = CommandFlow(coroutineScope)

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    loadingState = if (onStart) LoadingState.LoadingOnStart
                    else LoadingState.RefreshLoading,
                )
            }
            when (val result = scripterRepository.getDevices()) {
                is ApiResponse.Success -> {
                    _stateFlow.update { it.copy(devices = result.data) }
                }

                is ApiResponse.Error -> {
                    uiCommandsFlow.tryEmit(MainUiCommand.ShowNetworkError)
                }
            }
            if (!onStart) delay(500)
            _stateFlow.update {
                it.copy(
                    loadingState = LoadingState.None
                )
            }
        }
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
