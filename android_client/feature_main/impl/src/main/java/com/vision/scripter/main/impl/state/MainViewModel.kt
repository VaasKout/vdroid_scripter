package com.vision.scripter.main.impl.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.main.impl.ui.MainUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class MainViewModel @Inject constructor(
    private val mainInteractor: MainInteractor,
) : ViewModel(), MainUiStateHolder by mainInteractor {

    override fun onCleared() {
        super.onCleared()
        mainInteractor.clear()
    }
}
