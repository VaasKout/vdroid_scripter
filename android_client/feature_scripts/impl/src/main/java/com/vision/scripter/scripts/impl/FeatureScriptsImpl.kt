package com.vision.scripter.scripts.impl

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.remember
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.compose.composable
import com.vision.scripter.scripts.api.FeatureScripts
import com.vision.scripter.scripts.api.ScriptsRoute
import com.vision.scripter.scripts.impl.commandobservers.ScriptsCommandObserver
import com.vision.scripter.scripts.impl.state.ScriptsViewModel
import com.vision.scripter.scripts.impl.ui.ScriptsScreen
import dagger.hilt.android.scopes.ActivityScoped
import javax.inject.Inject

@ActivityScoped
class FeatureScriptsImpl @Inject constructor() : FeatureScripts {

    override fun register(
        navGraphBuilder: NavGraphBuilder,
        navController: NavController
    ) {
        navGraphBuilder.composable(route = ScriptsRoute) {
            val snackbarHostState = remember { SnackbarHostState() }
            val scriptsViewModel = hiltViewModel<ScriptsViewModel>()
            ScriptsCommandObserver(
                uiStateHolder = scriptsViewModel,
                navController = navController,
                snackbarHostState = snackbarHostState,
            )

            ScriptsScreen(
                uiStateHolder = scriptsViewModel,
                snackbarHostState = snackbarHostState,
            )
        }
    }
}