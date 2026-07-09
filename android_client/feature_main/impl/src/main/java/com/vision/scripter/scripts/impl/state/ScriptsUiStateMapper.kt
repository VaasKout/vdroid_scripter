package com.vision.scripter.scripts.impl.state

import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toImmutableList
import javax.inject.Inject

@ViewModelScoped
class ScriptsUiStateMapper @Inject constructor() {
    fun map(state: ScriptsState): ScriptsUiState {
        return ScriptsUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            isRefreshing = state.loadingState == LoadingState.RefreshLoading,
            scripts = state.scripts.toImmutableList(),
            scriptNameToDelete = state.scriptNameToDelete,
        )
    }
}