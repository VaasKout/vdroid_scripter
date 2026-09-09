package com.vision.scripter.library.state

import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import javax.inject.Inject

@ViewModelScoped
class LibraryUiStateMapper @Inject constructor() {
    fun map(state: LibraryState): LibraryUiState {
        return LibraryUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            isRefreshing = state.loadingState == LoadingState.RefreshLoading,
            imagesCount = state.images.size,
            actionsCount = state.actions.size,
            routesCount = state.routes.size,
        )
    }
}
