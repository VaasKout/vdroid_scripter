package com.vision.scripter.scripts.impl.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.RunScriptsRequest
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
    private val scripterDataSource: ScripterDataSource,
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

            val location = currentState.selectedLocation
            val result = if (location.isEmpty()) scripterDataSource.getLocations()
            else scripterDataSource.getLocationScripts(location)

            when (result) {
                is ApiResponse.Success -> _stateFlow.update {
                    if (location.isEmpty()) it.copy(locations = result.data)
                    else it.copy(scripts = result.data)
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

    override fun onLocationClick(location: String) {
        _stateFlow.update { it.copy(selectedLocation = location, scripts = listOf()) }
        onLoadData(onStart = true)
    }

    override fun onBack() {
        _stateFlow.update { it.copy(selectedLocation = "", scripts = listOf()) }
        onLoadData(onStart = true)
    }

    override fun onPlayScript(name: String) {
        _stateFlow.update {
            it.copy(
                bottomSheetData = ScriptsState.BottomSheetData(
                    selectedScript = name,
                ),
                itemToDelete = "",
            )
        }

        coroutineScope.launch {
            when (val result = scripterDataSource.getDevices()) {
                is ApiResponse.Success -> {
                    val bottomSheetState = currentState.bottomSheetData ?: return@launch
                    val selectedDevice = result.data.firstOrNull()?.serial ?: ""
                    _stateFlow.update {
                        it.copy(
                            bottomSheetData = bottomSheetState.copy(
                                isLoading = false,
                                selectedDevice = selectedDevice,
                                devices = result.data,
                            )
                        )
                    }
                }

                is ApiResponse.Error -> {
                    uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
                    val bottomSheetState = currentState.bottomSheetData ?: return@launch
                    _stateFlow.update {
                        it.copy(
                            bottomSheetData = bottomSheetState.copy(
                                isLoading = false,
                            )
                        )
                    }
                }
            }
        }
    }

    override fun onEditScript(name: String) {
        val location = currentState.selectedLocation
        if (location.isEmpty() || name.isEmpty()) return
        uiCommandsFlow.tryEmit(ScriptsUiCommand.OpenEditScript(location = location, name = name))
    }

    override fun onSelectDevice(serial: String) {
        if (serial.isEmpty()) return
        val bottomSheetData = currentState.bottomSheetData ?: return
        _stateFlow.update {
            it.copy(bottomSheetData = bottomSheetData.copy(selectedDevice = serial))
        }
    }

    override fun onConfirmRunScript() {
        val bottomSheetState = currentState.bottomSheetData ?: return
        val serial = bottomSheetState.selectedDevice
        val name = bottomSheetState.selectedScript
        val location = currentState.selectedLocation
        if (serial.isEmpty() || name.isEmpty() || location.isEmpty()) return

        coroutineScope.launch {
            val started = scripterDataSource.runScripts(
                serial = serial,
                scripts = listOf(RunScriptsRequest.ScriptRef(location = location, name = name)),
            )
            if (!started) uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
        }
        onDismiss()
    }

    override fun onDismiss() {
        _stateFlow.update {
            it.copy(
                bottomSheetData = null,
                itemToDelete = "",
            )
        }
    }

    override fun onDeleteItem(item: String) {
        _stateFlow.update {
            it.copy(itemToDelete = item)
        }
    }

    override fun onConfirmDelete() {
        val selectedLocation = currentState.selectedLocation
        val targetName = currentState.itemToDelete.ifEmpty { return }
        coroutineScope.launch {
            val deleted = if (selectedLocation.isNotEmpty()) {
                scripterDataSource.deleteScript(location = selectedLocation, name = targetName)
            } else {
                scripterDataSource.deleteLocation(location = targetName)
            }

            if (!deleted) uiCommandsFlow.tryEmit(ScriptsUiCommand.ShowNetworkError)
            onDismiss()
            onLoadData(onStart = false)
        }
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
