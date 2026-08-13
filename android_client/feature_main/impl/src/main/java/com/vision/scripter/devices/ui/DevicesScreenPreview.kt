package com.vision.scripter.devices.ui

import com.vision.scripter.main.impl.R
import com.vision.scripter.ui.CommandFlow
import kotlinx.collections.immutable.persistentListOf
import kotlinx.collections.immutable.persistentMapOf
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharedFlow

internal val devicesUiStateLoadingPreview = DevicesUiState()
internal val devicesUiStateNoDataPreview = DevicesUiState(isLoading = false)
internal val devicesUiStateWithDataPreview = DevicesUiState(
    isLoading = false,
    devices = persistentListOf(
        UiDevice(
            serial = "c88893qa",
            deviceParams = persistentMapOf(
                R.string.serial_key to "c9858321",
                R.string.model_key to "Samsung",
            ),
        ),
    )
)

internal class DevicesScreenUiStateHolderPreview(state: DevicesUiState) : DevicesUiStateHolder {

    override val uiStateFlow: SharedFlow<DevicesUiState> = MutableStateFlow(state)
    override val uiCommandsFlow: CommandFlow<DevicesUiCommand>
        get() = throw UnsupportedOperationException()

    override fun onLoadData(onStart: Boolean) {}

}