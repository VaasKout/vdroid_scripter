package com.vision.scripter.streaming.impl.screen.main.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class StreamingViewModel @Inject constructor(
    private val streamingInteractor: StreamingInteractor,
) : ViewModel(), StreamingUiStateHolder by streamingInteractor {

    val menuInteractor = streamingInteractor.menuInteractor

    override fun onCleared() {
        super.onCleared()
        streamingInteractor.clear()
        menuInteractor.clear()
    }
}