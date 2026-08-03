package com.vision.scripter.editscript.impl.state

import com.vision.scripter.editscript.impl.ui.EditScriptUiState
import com.vision.scripter.editscript.impl.ui.ParamUiData
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toImmutableList
import javax.inject.Inject

@ViewModelScoped
class EditScriptUiStateMapper @Inject constructor() {
    fun map(state: EditScriptState): EditScriptUiState {
        return EditScriptUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            name = state.scriptName,
            node = state.node,
            nextNode = state.nextNode,
            timeout = state.timeout,
            params = state.params.mapIndexed { index, param ->
                ParamUiData(
                    id = index,
                    title = "${param.type}: ${param.value}",
                    locale = param.locale,
                )
            }.toImmutableList(),
            eventsCount = state.events.size,
            showSaveDialog = state.showDialog,
        )
    }
}
