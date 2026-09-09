package com.vision.scripter.main

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.remember
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.vision.scripter.devices.commandobservers.DevicesUiCommandObserver
import com.vision.scripter.devices.state.DevicesViewModel
import com.vision.scripter.library.commandobservers.LibraryCommandObserver
import com.vision.scripter.library.state.LibraryType
import com.vision.scripter.library.state.LibraryViewModel
import com.vision.scripter.libraryitems.LibraryItemsRouteWithArgs
import com.vision.scripter.libraryitems.LibraryTypeArg
import com.vision.scripter.libraryitems.commandobservers.LibraryItemsCommandObserver
import com.vision.scripter.libraryitems.state.LibraryItemsViewModel
import com.vision.scripter.libraryitems.ui.LibraryItemsScreen
import com.vision.scripter.main.api.FeatureMain
import com.vision.scripter.main.api.MainRoute
import com.vision.scripter.main.ui.MainContainerScreen
import dagger.hilt.android.scopes.ActivityScoped
import javax.inject.Inject

@ActivityScoped
class FeatureMainImpl @Inject constructor() : FeatureMain {
    override fun register(
        navGraphBuilder: NavGraphBuilder,
        navController: NavController,
    ) {
        navGraphBuilder.composable(route = MainRoute) {
            val snackbarHostState = remember { SnackbarHostState() }

            val devicesViewModel = hiltViewModel<DevicesViewModel>()
            val libraryViewModel = hiltViewModel<LibraryViewModel>()

            DevicesUiCommandObserver(
                uiStateHolder = devicesViewModel,
                navController = navController,
                snackbarHostState = snackbarHostState,
            )

            LibraryCommandObserver(
                uiStateHolder = libraryViewModel,
                navController = navController,
                snackbarHostState = snackbarHostState,
            )

            MainContainerScreen(
                devicesUiStateHolder = devicesViewModel,
                libraryUiStateHolder = libraryViewModel,
                snackbarHostState = snackbarHostState,
            )
        }

        navGraphBuilder.composable(
            route = LibraryItemsRouteWithArgs,
            arguments = listOf(navArgument(LibraryTypeArg) { type = NavType.StringType }),
        ) { backStackEntry ->
            val typeName = backStackEntry.arguments?.getString(LibraryTypeArg).orEmpty()
            val type = LibraryType.entries.firstOrNull { it.name == typeName } ?: LibraryType.IMAGES
            val snackbarHostState = remember { SnackbarHostState() }
            val libraryItemsViewModel = hiltViewModel<LibraryItemsViewModel>()

            LibraryItemsCommandObserver(
                uiStateHolder = libraryItemsViewModel,
                snackbarHostState = snackbarHostState,
            )

            LibraryItemsScreen(
                uiStateHolder = libraryItemsViewModel,
                type = type,
                snackbarHostState = snackbarHostState,
                onBack = { navController.popBackStack() },
            )
        }
    }
}
