package com.vision.scripter.devices.state

import androidx.lifecycle.ViewModel
import com.vision.scripter.devices.ui.DevicesUiStateHolder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject

@HiltViewModel
class DevicesViewModel @Inject constructor(
    private val devicesInteractor: DevicesInteractor,
) : ViewModel(), DevicesUiStateHolder by devicesInteractor {

    override fun onCleared() {
        super.onCleared()
        devicesInteractor.clear()
    }
}
