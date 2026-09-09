package com.vision.scripter.libraryitems.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import com.vision.scripter.libraryitems.ui.LibraryItemsUiCommand
import com.vision.scripter.libraryitems.ui.LibraryItemsUiStateHolder
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun LibraryItemsCommandObserver(
    uiStateHolder: LibraryItemsUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is LibraryItemsUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }

            is LibraryItemsUiCommand.ShowError -> {
                snackbarHostState.showSnackbar(it.text.ifEmpty { commonNetworkError })
            }
        }
    }
}
