package com.vision.scripter.main

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.remember
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.compose.composable
import com.vision.scripter.devices.commandobservers.DevicesUiCommandObserver
import com.vision.scripter.devices.state.DevicesViewModel
import com.vision.scripter.library.commandobservers.LibraryCommandObserver
import com.vision.scripter.library.state.LibraryViewModel
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
                snackbarHostState = snackbarHostState,
            )

            MainContainerScreen(
                devicesUiStateHolder = devicesViewModel,
                libraryUiStateHolder = libraryViewModel,
                snackbarHostState = snackbarHostState,
            )
        }
    }
}
