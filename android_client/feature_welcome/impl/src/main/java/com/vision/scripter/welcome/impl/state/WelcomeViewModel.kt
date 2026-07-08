package com.vision.scripter.welcome.impl.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.welcome.impl.ui.WelcomeUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class WelcomeViewModel @Inject constructor(
    val welcomeInteractor: WelcomeInteractor,
) : ViewModel(), WelcomeUiStateHolder by welcomeInteractor {

    override fun onCleared() {
        super.onCleared()
        welcomeInteractor.clear()
    }
}
