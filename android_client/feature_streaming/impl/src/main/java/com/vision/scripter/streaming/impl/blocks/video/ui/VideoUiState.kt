package com.vision.scripter.streaming.impl.blocks.video.ui

import androidx.compose.runtime.Immutable
import com.vision.scripter.data.api.models.CvRectangle
import com.vision.scripter.data.api.models.RectangleWithText
import kotlinx.collections.immutable.ImmutableList
import kotlinx.collections.immutable.persistentListOf

@Immutable
data class VideoUiState(
    val streamingIsLoading: Boolean = true,
    val rectangles: ImmutableList<CvRectangle> = persistentListOf(),
    val selectedRectangles: ImmutableList<CvRectangle> = persistentListOf(),
    val keyboardButtons: ImmutableList<RectangleWithText> = persistentListOf(),
)
