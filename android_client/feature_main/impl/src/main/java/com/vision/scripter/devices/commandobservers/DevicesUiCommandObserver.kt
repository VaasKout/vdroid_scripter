package com.vision.scripter.devices.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import androidx.navigation.NavController
import com.vision.scripter.devices.ui.DevicesUiCommand
import com.vision.scripter.devices.ui.DevicesUiStateHolder
import com.vision.scripter.streaming.api.StreamingRoute
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun DevicesUiCommandObserver(
    uiStateHolder: DevicesUiStateHolder,
    navController: NavController,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is DevicesUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }

            is DevicesUiCommand.NavigateToStreaming -> {
                navController.navigate(StreamingRoute + "/${it.serial}")
            }
        }
    }
}