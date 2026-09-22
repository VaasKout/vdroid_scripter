package com.vision.scripter.streaming.impl.blocks.video.ui

import androidx.compose.runtime.Immutable
import androidx.compose.ui.graphics.Color
import com.vision.scripter.data.api.models.CvRectangle
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class VideoUiState(
    val streamingIsLoading: Boolean = true,
    val rectangles: ImmutableList<UiRectangle> = persistentListOf(),
)

@Immutable
data class UiRectangle(
    val rectangle: CvRectangle,
    val color: Color,
)
