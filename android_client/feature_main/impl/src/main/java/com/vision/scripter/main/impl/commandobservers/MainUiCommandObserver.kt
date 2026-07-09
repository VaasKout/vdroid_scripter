package com.vision.scripter.main.impl.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import androidx.navigation.NavController
import com.vision.scripter.main.impl.ui.MainUiCommand
import com.vision.scripter.main.impl.ui.MainUiStateHolder
import com.vision.scripter.streaming.api.StreamingRoute
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun MainUiCommandObserver(
    uiStateHolder: MainUiStateHolder,
    navController: NavController,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is MainUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }

            is MainUiCommand.NavigateToStreaming -> {
                navController.navigate(StreamingRoute + "/${it.serial}")
            }
        }
    }
}