package com.vision.scripter.devices.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.devices.ui.DevicesUiCommand
import com.vision.scripter.devices.ui.DevicesUiState
import com.vision.scripter.devices.ui.DevicesUiStateHolder
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
import kotlin.time.Duration.Companion.milliseconds

@ViewModelScoped
class DevicesInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterDataSource: ScripterDataSource,
    private val uiStateMapper: DevicesUiStateMapper,
) : DevicesUiStateHolder {

    private val _stateFlow = MutableStateFlow(DevicesState())
    private val stateFlow: StateFlow<DevicesState> = _stateFlow.asStateFlow()
    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("main_interactor")

    override val uiStateFlow: SharedFlow<DevicesUiState>
        get() = stateFlow.map(uiStateMapper::map)
            .shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<DevicesUiCommand> = CommandFlow(coroutineScope)

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    loadingState = if (onStart) LoadingState.LoadingOnStart
                    else LoadingState.RefreshLoading,
                )
            }
            when (val result = scripterDataSource.getDevices()) {
                is ApiResponse.Success -> {
                    _stateFlow.update { it.copy(devices = result.data) }
                }

                is ApiResponse.Error -> {
                    uiCommandsFlow.tryEmit(DevicesUiCommand.ShowNetworkError)
                }
            }
            if (!onStart) delay(500.milliseconds)
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
