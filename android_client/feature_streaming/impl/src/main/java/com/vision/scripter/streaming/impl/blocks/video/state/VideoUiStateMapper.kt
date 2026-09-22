package com.vision.scripter.streaming.impl.blocks.video.state

import androidx.compose.ui.graphics.Color
import com.vision.scripter.data.api.models.LandmarkType
import com.vision.scripter.streaming.impl.blocks.video.ui.UiRectangle
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiState
import dagger.hilt.android.scopes.ViewModelScoped
import kotlinx.collections.immutable.toPersistentList
import javax.inject.Inject

@ViewModelScoped
class VideoUiStateMapper @Inject constructor() {
    fun map(state: VideoState): VideoUiState {
        return VideoUiState(
            streamingIsLoading = state.streamingData == null,
            rectangles = state.uiRectangles().toPersistentList(),
        )
    }

    private fun VideoState.uiRectangles(): List<UiRectangle> = buildList {
        overlay.rectangles.mapTo(this) { UiRectangle(rectangle = it, color = Color.Red) }
        overlay.scan.mapTo(this) { UiRectangle(rectangle = it.rectangle, color = it.type.color()) }
        selectedRectangles.mapTo(this) { UiRectangle(rectangle = it, color = Color.Blue) }
    }

    private fun LandmarkType.color(): Color = when (this) {
        LandmarkType.TEXT -> Color.Green
        LandmarkType.YOLO -> Color.Yellow
        LandmarkType.IMAGE -> Color.Magenta
    }
}
