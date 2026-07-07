package com.vision.scripter.streaming.impl.screen.main.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuInteractor
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class StreamingViewModel @Inject constructor(
    private val streamingInteractor: StreamingInteractor,
    menuInteractor: MenuInteractor,
) : ViewModel(), StreamingUiStateHolder by streamingInteractor {

    val menuUiStateHolder: MenuUiStateHolder = menuInteractor

    override fun onCleared() {
        super.onCleared()
        streamingInteractor.closeStreams()
    }
}