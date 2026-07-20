package com.vision.scripter.streaming.impl.blocks.menu.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class MenuViewModel @Inject constructor(
    private val menuInteractor: MenuInteractor,
) : ViewModel(), MenuUiStateHolder by menuInteractor {
    override fun onCleared() {
        super.onCleared()
        menuInteractor.clear()
    }
}
