package com.vision.scripter.streaming.impl.screen.main.state

import com.vision.scripter.streaming.impl.blocks.menu.state.DialogState
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiState
import com.vision.scripter.ui.states.LoadingState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toPersistentList
import javax.inject.Inject

@ViewModelScoped
class StreamingUiStateMapper @Inject constructor() {
    fun map(
        state: StreamingState,
        menuState: MenuState,
        dialogState: DialogState,
    ): StreamingUiState {
        val keyboardButtons = if (
            menuState is MenuState.Keyboard &&
            !menuState.showCvRectangles
        ) {
            state.keyboard.buttons
        } else {
            listOf()
        }

        return StreamingUiState(
            isLoading = state.loadingState == LoadingState.LoadingOnStart,
            hasConnection = !state.streamingHost.isBlank() && state.streamingData != null,
            rectangles = state.cvRectangles.toPersistentList(),
            selectedRectangles = state.selectedRectangles.toPersistentList(),
            keyboardButtons = keyboardButtons.toPersistentList(),
            menuState = menuState,
            dialogState = dialogState,
        )
    }
}
