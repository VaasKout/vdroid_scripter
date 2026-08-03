package com.vision.scripter.editscript.impl.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import androidx.navigation.NavController
import com.vision.scripter.editscript.impl.ui.EditScriptUiCommand
import com.vision.scripter.editscript.impl.ui.EditScriptUiStateHolder
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun EditScriptCommandObserver(
    uiStateHolder: EditScriptUiStateHolder,
    navController: NavController,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is EditScriptUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }

            is EditScriptUiCommand.NavigateBack -> {
                navController.popBackStack()
            }
        }
    }
}
