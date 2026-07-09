package com.vision.scripter.scripts.impl.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.vision.scripter.scripts.impl.state.ScriptsUiCommand
import com.vision.scripter.scripts.impl.state.ScriptsUiStateHolder
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun ScriptsCommandObserver(
    uiStateHolder: ScriptsUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is ScriptsUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }
        }
    }
}