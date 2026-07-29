package com.vision.scripter.streaming.impl.screen.state

import com.vision.scripter.streaming.impl.screen.ui.StreamingUiState
import dagger.hilt.android.scopes.ViewModelScoped
import javax.inject.Inject

@ViewModelScoped
class StreamingUiStateMapper @Inject constructor() {
    fun map(state: StreamingState): StreamingUiState {
        return StreamingUiState(
            isLoading = state.loading,
            isError = state.isError,
        )
    }
}
