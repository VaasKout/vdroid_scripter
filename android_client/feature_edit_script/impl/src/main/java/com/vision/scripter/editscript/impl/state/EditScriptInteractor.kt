package com.vision.scripter.editscript.impl.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.Script
import com.vision.scripter.data.api.models.isEmpty
import com.vision.scripter.editscript.impl.ui.EditScriptUiCommand
import com.vision.scripter.editscript.impl.ui.EditScriptUiState
import com.vision.scripter.editscript.impl.ui.EditScriptUiStateHolder
import com.vision.scripter.network.api.ApiResponse
import com.vision.scripter.ui.CommandFlow
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.cancel
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
class EditScriptInteractor @Inject constructor(
    coroutineScopeFactory: CoroutineScopeFactory,
    private val scripterDataSource: ScripterDataSource,
    private val uiStateMapper: EditScriptUiStateMapper,
) : EditScriptUiStateHolder {

    private val _stateFlow = MutableStateFlow(EditScriptState())
    private val stateFlow: StateFlow<EditScriptState> = _stateFlow.asStateFlow()

    private val coroutineScope: CoroutineScope =
        coroutineScopeFactory.createBackgroundScope("edit_script_interactor")

    private val currentState: EditScriptState
        get() = _stateFlow.value

    override val uiStateFlow: SharedFlow<EditScriptUiState>
        get() = stateFlow.map(uiStateMapper::map)
            .shareIn(coroutineScope, SharingStarted.WhileSubscribed(), replay = 1)

    override val uiCommandsFlow: CommandFlow<EditScriptUiCommand> = CommandFlow(coroutineScope)

    override fun init(node: String, name: String) {
        if (name.isEmpty()) return
        _stateFlow.update {
            it.copy(
                initialNode = node,
                script = Script(name = name),
            )
        }

        coroutineScope.launch {
            when (val result = scripterDataSource.getScriptInfo(node = node, name = name)) {
                is ApiResponse.Success -> _stateFlow.update {
                    it.copy(
                        loadingState = LoadingState.None,
                        script = result.data,
                    )
                }

                is ApiResponse.Error -> {
                    uiCommandsFlow.tryEmit(EditScriptUiCommand.ShowNetworkError)
                    _stateFlow.update { it.copy(loadingState = LoadingState.None) }
                }
            }
        }
    }

    override fun onNodeChanged(value: String) {
        val updatedScript = currentState.script?.copy(node = value.trim()) ?: return
        _stateFlow.update { it.copy(script = updatedScript) }
    }

    override fun onNextNodeChanged(value: String) {
        val updatedScript = currentState.script?.copy(nextNode = value.trim()) ?: return
        _stateFlow.update { it.copy(script = updatedScript) }
    }

    override fun onTimeoutChanged(value: String) {
        val script = currentState.script ?: return
        val newTimeout = value.filter(Char::isDigit).toIntOrNull() ?: 0
        val updatedScript = script.copy(timeout = newTimeout)
        _stateFlow.update { it.copy(script = updatedScript) }
    }

    override fun onDeleteParam(id: Int) {
        val script = currentState.script ?: return
        val updatedParams = script.params.filterIndexed { index, _ -> index != id }
        val updatedScript = script.copy(params = updatedParams)
        _stateFlow.update {
            it.copy(script = updatedScript)
        }
    }

    override fun onDeleteEvents() {
        val updatedScript = currentState.script?.copy(events = listOf()) ?: return
        _stateFlow.update { it.copy(script = updatedScript) }
    }

    override fun onTopbarActionClicked() {
        val state = currentState
        if (state.loadingState != LoadingState.None) return
        val script = currentState.script ?: return
        if (script.node.isEmpty()) return
        _stateFlow.update { it.copy(showDialog = true) }
    }

    override fun onConfirmDialog() {
        _stateFlow.update { it.copy(showDialog = false) }
        val script = currentState.script ?: return
        val state = currentState
        coroutineScope.launch {
            if (!script.isEmpty()) {
                val saved = scripterDataSource.saveScript(script)
                if (!saved) {
                    uiCommandsFlow.tryEmit(EditScriptUiCommand.ShowNetworkError)
                    return@launch
                }
            }

            if (script.isEmpty() || script.node != currentState.initialNode) {
                val deleted = scripterDataSource.deleteScript(
                    node = state.initialNode,
                    name = script.name,
                )
                if (!deleted) {
                    uiCommandsFlow.tryEmit(EditScriptUiCommand.ShowNetworkError)
                    return@launch
                }
            }

            uiCommandsFlow.tryEmit(EditScriptUiCommand.NavigateBack)
        }
    }

    override fun onDismissDialog() {
        _stateFlow.update { it.copy(showDialog = false) }
    }

    override fun onBackClicked() {
        uiCommandsFlow.tryEmit(EditScriptUiCommand.NavigateBack)
    }

    fun clear() {
        coroutineScope.cancel()
    }
}
