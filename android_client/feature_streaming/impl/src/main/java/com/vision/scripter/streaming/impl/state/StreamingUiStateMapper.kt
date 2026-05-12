package com.vision.scripter.streaming.impl.state

import com.vision.scripter.streaming.impl.ui.StreamingUiState
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import javax.inject.Inject

@ViewModelScoped
class StreamingUiStateMapper @Inject constructor() {
    fun map(state: StreamingState): StreamingUiState {
        val keyboardButtons = if (state.menuState is MenuState.TypingText) {
            state.keyboard.buttons
        } else {
            listOf()
        }

        return StreamingUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            hasConnection = !state.streamingHost.isBlank() && state.streamingData != null,
            rectangles = state.cvRectangles,
            keyboardButtons = keyboardButtons,
            menuState = state.menuState,
            showTextDialog = state.showTextDialog,
            showRecordDialog = state.showRecordDialog,
            showKeyboardDialog = state.showKeyboardDialog,
        )
    }
}