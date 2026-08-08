package com.vision.scripter.editscript.impl.state

import com.vision.scripter.data.api.models.isEmpty
import com.vision.scripter.editscript.impl.ui.EditScriptUiState
import com.vision.scripter.editscript.impl.ui.ParamUiData
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toImmutableList
import javax.inject.Inject

@ViewModelScoped
class EditScriptUiStateMapper @Inject constructor() {
    fun map(state: EditScriptState): EditScriptUiState {
        val script = state.script ?: return EditScriptUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
        )
        return EditScriptUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            name = script.name,
            node = script.node,
            nextNode = script.nextNode,
            timeout = script.timeout.toString(),
            params = script.params.mapIndexed { index, param ->
                ParamUiData(
                    id = index,
                    title = "${param.type}: ${param.value}",
                    locale = param.locale,
                )
            }.toImmutableList(),
            eventsCount = script.events.size,
            canSkip = script.canSkip,
            deleteMode = script.isEmpty() && state.loadingState == LoadingState.None,
            showDialog = state.showDialog,
        )
    }
}
