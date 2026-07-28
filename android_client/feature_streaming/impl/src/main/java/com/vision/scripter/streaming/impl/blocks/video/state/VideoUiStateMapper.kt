package com.vision.scripter.streaming.impl.blocks.video.state

import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toPersistentList
import javax.inject.Inject

@ViewModelScoped
class VideoUiStateMapper @Inject constructor() {
    fun map(state: VideoState): VideoUiState {
        return VideoUiState(
            streamingIsLoading = state.streamingData == null,
            rectangles = state.cvRectangles.toPersistentList(),
            selectedRectangles = state.selectedRectangles.toPersistentList(),
            keyboardButtons = state.keyboard.buttons.toPersistentList(),
        )
    }
}
