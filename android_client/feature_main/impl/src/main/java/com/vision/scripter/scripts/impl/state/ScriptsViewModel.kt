package com.vision.scripter.scripts.impl.state

import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
internal class ScriptsViewModel @Inject constructor(
    private val scriptsInteractor: ScriptsInteractor,
) : ViewModel(), ScriptsUiStateHolder by scriptsInteractor {

    override fun onCleared() {
        super.onCleared()
        scriptsInteractor.clear()
    }
}
