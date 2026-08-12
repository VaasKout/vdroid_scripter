package com.vision.scripter.library.impl.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.vision.scripter.library.impl.state.LibraryUiCommand
import com.vision.scripter.library.impl.state.LibraryUiStateHolder
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun LibraryCommandObserver(
    uiStateHolder: LibraryUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is LibraryUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }
        }
    }
}
