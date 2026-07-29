package com.vision.scripter.streaming.impl.blocks.video.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.streaming.impl.blocks.video.ui.VideoUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class VideoViewModel @Inject constructor(
    private val videoInteractor: VideoInteractor,
) : ViewModel(), VideoUiStateHolder by videoInteractor {

    override fun onCleared() {
        super.onCleared()
        videoInteractor.closeStreams()
        videoInteractor.clear()
    }
}
