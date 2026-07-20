package com.vision.scripter.streaming.impl.blocks.video.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.streaming.impl.blocks.video.commandobservers.VideoCommandObserver
import com.vision.scripter.streaming.impl.blocks.video.commandobservers.VideoSharedCommandObserver
import com.vision.scripter.streaming.impl.blocks.video.state.VideoViewModel
import com.vision.scripter.streaming.impl.blocks.video.ui.items.RectanglesCanvas
import com.vision.scripter.streaming.impl.blocks.video.ui.items.VideoSurface
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder

@Composable
fun VideoBlock(
    modifier: Modifier = Modifier,
    sharedStateHolder: StreamingUiStateHolder,
) {
    val uiStateHolder = hiltViewModel<VideoViewModel>()
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle().value

    VideoCommandObserver(
        uiStateHolder = uiStateHolder,
        sharedStateHolder = sharedStateHolder,
    )

    VideoSharedCommandObserver(
        uiStateHolder = uiStateHolder,
        sharedStateHolder = sharedStateHolder,
    )

    VideoSurface(
        modifier = modifier,
        onSurfaceCreated = uiStateHolder::onVideoSurfaceCreated,
        onSurfaceDestroyed = uiStateHolder::onVideoSurfaceDestroyed,
        onTouch = uiStateHolder::onTouchEvent,
    )

    RectanglesCanvas(
        modifier = modifier,
        cvRectangles = state.rectangles,
        selectedRectangles = state.selectedRectangles,
        keyboardButtons = state.keyboardButtons,
    )
}

// TODO preview
