package com.vision.scripter.library.commandobservers

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.res.stringResource
import androidx.navigation.NavController
import com.vision.scripter.library.state.LibraryUiCommand
import com.vision.scripter.library.state.LibraryUiStateHolder
import com.vision.scripter.libraryitems.libraryItemsRoute
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun LibraryCommandObserver(
    uiStateHolder: LibraryUiStateHolder,
    navController: NavController,
    snackbarHostState: SnackbarHostState,
) {
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is LibraryUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }

            is LibraryUiCommand.OpenList -> {
                navController.navigate(libraryItemsRoute(it.type))
            }
        }
    }
}
