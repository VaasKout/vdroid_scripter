package com.vision.scripter.editscript.impl.state

import com.vision.scripter.coroutines.api.CoroutineScopeFactory
import com.vision.scripter.data.api.ScripterDataSource
import com.vision.scripter.data.api.models.Script
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
        if (currentState.scriptName.isNotEmpty()) return

        _stateFlow.update {
            it.copy(
                initialNode = node,
                scriptName = name,
            )
        }

        coroutineScope.launch {
            when (val result = scripterDataSource.getScriptInfo(node = node, name = name)) {
                is ApiResponse.Success -> _stateFlow.update {
                    it.copy(
                        loadingState = LoadingState.None,
                        node = result.data.node,
                        nextNode = result.data.nextNode,
                        timeout = result.data.timeout.toString(),
                        params = result.data.params,
                        events = result.data.events,
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
        _stateFlow.update { it.copy(node = value) }
    }

    override fun onNextNodeChanged(value: String) {
        _stateFlow.update { it.copy(nextNode = value) }
    }

    override fun onTimeoutChanged(value: String) {
        _stateFlow.update { it.copy(timeout = value.filter(Char::isDigit)) }
    }

    override fun onDeleteParam(id: Int) {
        _stateFlow.update {
            it.copy(params = it.params.filterIndexed { index, _ -> index != id })
        }
    }

    override fun onDeleteEvents() {
        _stateFlow.update { it.copy(events = listOf()) }
    }

    override fun onSaveClicked() {
        if (currentState.node.isBlank()) return
        _stateFlow.update { it.copy(showDialog = true) }
    }

    override fun onConfirmSave() {
        val state = currentState
        _stateFlow.update { it.copy(showDialog = false) }

        coroutineScope.launch {
            val script = Script(
                name = state.scriptName,
                node = state.node.trim(),
                nextNode = state.nextNode.trim(),
                params = state.params,
                events = state.events,
                timeout = state.timeout.toIntOrNull() ?: 0,
            )

            val saved = scripterDataSource.saveScript(script)
            if (!saved) {
                uiCommandsFlow.tryEmit(EditScriptUiCommand.ShowNetworkError)
                return@launch
            }

            if (script.node != state.initialNode) {
                val deleted = scripterDataSource.deleteScript(
                    node = state.initialNode,
                    name = state.scriptName,
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
