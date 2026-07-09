package com.vision.scripter.scripts.impl.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterRepository
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
internal class ScriptsInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterRepository: ScripterRepository,
    private val uiStateMapper: ScriptsUiStateMapper,
) : ScriptsUiStateHolder {

    private val _stateFlow = MutableStateFlow(ScriptsState())
    private val stateFlow: StateFlow<ScriptsState> = _stateFlow.asStateFlow()
    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("scripts_interactor")

    private val currentState: ScriptsState
        get() = _stateFlow.value

    override val uiStateFlow: SharedFlow<ScriptsUiState>
        get() = stateFlow.map(uiStateMapper::map)
            .shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<ScriptsUiCommand> = CommandFlow(coroutineScope)

    override fun onLoadData(onStart: Boolean) {
        coroutineScope.launch {
            _stateFlow.update {
                it.copy(
                    loadingState = if (onStart) LoadingState.LoadingOnStart
                    else LoadingState.RefreshLoading,
                )
            }

            when (val result = scripterRepository.getScripts()) {
                is ApiResponse.Success -> {
                    _stateFlow.update { it.copy(scripts = result.data) }
                }

                is ApiResponse.Error -> {
                    uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
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

    override fun onPlayScript(name: String) {
        _stateFlow.update {
            it.copy(
                scriptToRun = name,
                selectedSerial = "",
                devices = listOf(),
                isDevicesLoading = true,
            )
        }

        coroutineScope.launch {
            when (val result = scripterRepository.getDevices()) {
                is ApiResponse.Success -> {
                    _stateFlow.update {
                        it.copy(
                            devices = result.data,
                            selectedSerial = result.data.firstOrNull()?.serial ?: "",
                            isDevicesLoading = false,
                        )
                    }
                }

                is ApiResponse.Error -> {
                    uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
                    _stateFlow.update { it.copy(isDevicesLoading = false) }
                }
            }
        }
    }

    override fun onSelectDevice(serial: String) {
        _stateFlow.update { it.copy(selectedSerial = serial) }
    }

    override fun onConfirmRunScript() {
        val serial = currentState.selectedSerial
        val name = currentState.scriptToRun
        if (serial.isEmpty() || name.isEmpty()) return

        coroutineScope.launch {
            val started = scripterRepository.runScript(serial = serial, name = name)
            if (!started) uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
        }
        onDismissDevicePicker()
    }

    override fun onDismissDevicePicker() {
        _stateFlow.update {
            it.copy(
                scriptToRun = "",
                selectedSerial = "",
                devices = listOf(),
                isDevicesLoading = false,
            )
        }
    }

    override fun onDeleteScript(name: String) {
        _stateFlow.update {
            it.copy(
                scriptToDelete = name,
            )
        }
    }

    override fun onDismissDeleteDialog() {
        _stateFlow.update {
            it.copy(
                scriptToDelete = "",
            )
        }
    }

    override fun onConfirmDeleteScript() {
        coroutineScope.launch {
            val deleted = scripterRepository.deleteScript(name = currentState.scriptToDelete)
            if (!deleted) uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
            onDismissDeleteDialog()
            onLoadData(onStart = false)
        }
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
