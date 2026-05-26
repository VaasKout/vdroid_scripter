package com.vision.scripter.streaming.impl.state

import com.vision.scripter.streaming.impl.ui.StreamingUiState
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import javax.inject.Inject

@ViewModelScoped
class StreamingUiStateMapper @Inject constructor() {
    fun map(
        state: StreamingState,
        menuState: MenuState,
        dialogState: DialogState,
    ): StreamingUiState {
        val keyboardButtons = if (
            menuState is MenuState.TypingText ||
            menuState is MenuState.Usual && menuState.showKeyboardButtons
        ) {
            state.keyboard.buttons
        } else {
            listOf()
        }

        return StreamingUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            hasConnection = !state.streamingHost.isBlank() && state.streamingData != null,
            rectangles = state.cvRectangles,
            keyboardButtons = keyboardButtons,
            menuState = menuState,
            dialogState = dialogState,
        )
    }
}
