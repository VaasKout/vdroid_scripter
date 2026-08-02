package com.vision.scripter.scripts.impl.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
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

            val node = currentState.selectedNode
            val result = if (node.isEmpty()) scripterDataSource.getNodes()
            else scripterDataSource.getNodeScripts(node)

            when (result) {
                is ApiResponse.Success -> _stateFlow.update {
                    if (node.isEmpty()) it.copy(nodes = result.data)
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

    override fun onNodeClick(node: String) {
        _stateFlow.update { it.copy(selectedNode = node, scripts = listOf()) }
        onLoadData(onStart = true)
    }

    override fun onBack() {
        _stateFlow.update { it.copy(selectedNode = "", scripts = listOf()) }
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
        val node = currentState.selectedNode
        if (serial.isEmpty() || name.isEmpty() || node.isEmpty()) return

        coroutineScope.launch {
            val started = scripterDataSource.runScript(serial = serial, node = node, name = name)
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
        val selectedNode = currentState.selectedNode
        val targetName = currentState.itemToDelete.ifEmpty { return }
        coroutineScope.launch {
            val deleted = if (selectedNode.isNotEmpty()) {
                scripterDataSource.deleteScript(node = selectedNode, name = targetName)
            } else {
                scripterDataSource.deleteNode(node = targetName)
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
