package com.vision.scripter.streaming.impl.screen.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.streaming.impl.StreamingComponentManager
import com.vision.scripter.streaming.impl.screen.ui.StreamingUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class StreamingViewModel @Inject constructor(
    private val streamingInteractor: StreamingInteractor,
    private val streamingComponentManager: StreamingComponentManager,
) : ViewModel(), StreamingUiStateHolder by streamingInteractor {

    override fun onCleared() {
        super.onCleared()
        streamingInteractor.clear()
        streamingComponentManager.destroyComponent()
    }
}