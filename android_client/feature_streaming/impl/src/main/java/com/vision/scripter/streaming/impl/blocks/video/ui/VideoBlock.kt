package com.vision.scripter.streaming.impl.blocks.video.ui

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.vision.scripter.streaming.impl.blocks.video.state.VideoViewModel
import com.vision.scripter.streaming.impl.blocks.video.ui.items.RectanglesCanvas
import com.vision.scripter.streaming.impl.blocks.video.ui.items.VideoSurface

@Composable
fun VideoBlock(
    modifier: Modifier = Modifier,
    serialArg: String,
) {
    val uiStateHolder = hiltViewModel<VideoViewModel>()
    val state = uiStateHolder.uiStateFlow.collectAsStateWithLifecycle().value

    if (!state.connectionEstablished) return
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