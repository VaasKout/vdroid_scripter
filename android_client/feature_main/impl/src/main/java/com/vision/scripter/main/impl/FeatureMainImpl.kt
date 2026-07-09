package com.vision.scripter.main.impl

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.remember
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.compose.composable
import com.vision.scripter.main.api.FeatureMain
import com.vision.scripter.main.api.MainRoute
import com.vision.scripter.main.impl.commandobservers.MainUiCommandObserver
import com.vision.scripter.main.impl.state.MainViewModel
import com.vision.scripter.main.impl.ui.MainContainerScreen
import com.vision.scripter.scripts.impl.commandobservers.ScriptsCommandObserver
import com.vision.scripter.scripts.impl.state.ScriptsViewModel
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

            val mainViewModel = hiltViewModel<MainViewModel>()
            val scriptsViewModel = hiltViewModel<ScriptsViewModel>()

            MainUiCommandObserver(
                uiStateHolder = mainViewModel,
                navController = navController,
                snackbarHostState = snackbarHostState,
            )

            ScriptsCommandObserver(
                uiStateHolder = scriptsViewModel,
                snackbarHostState = snackbarHostState,
            )

            MainContainerScreen(
                mainUiStateHolder = mainViewModel,
                scriptsUiStateHolder = scriptsViewModel,
                snackbarHostState = snackbarHostState,
            )
        }
    }
}
