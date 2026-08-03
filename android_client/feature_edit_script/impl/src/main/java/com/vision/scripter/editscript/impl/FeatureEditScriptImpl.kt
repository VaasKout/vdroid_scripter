package com.vision.scripter.editscript.impl

import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.vision.scripter.editscript.api.EditScriptNameArg
import com.vision.scripter.editscript.api.EditScriptNodeArg
import com.vision.scripter.editscript.api.EditScriptRouteWithArgs
import com.vision.scripter.editscript.api.FeatureEditScript
import com.vision.scripter.editscript.impl.commandobservers.EditScriptCommandObserver
import com.vision.scripter.editscript.impl.state.EditScriptViewModel
import com.vision.scripter.editscript.impl.ui.EditScriptScreen
import dagger.hilt.android.scopes.ActivityScoped
import javax.inject.Inject

@ActivityScoped
class FeatureEditScriptImpl @Inject constructor() : FeatureEditScript {

    override fun register(
        navGraphBuilder: NavGraphBuilder,
        navController: NavController,
    ) {
        navGraphBuilder.composable(
            route = EditScriptRouteWithArgs,
            arguments = listOf(
                navArgument(EditScriptNodeArg) { type = NavType.StringType },
                navArgument(EditScriptNameArg) { type = NavType.StringType },
            ),
        ) { backStackEntry ->
            val node = backStackEntry.arguments?.getString(EditScriptNodeArg).orEmpty()
            val name = backStackEntry.arguments?.getString(EditScriptNameArg).orEmpty()
            val snackbarHostState = remember { SnackbarHostState() }
            val editScriptViewModel = hiltViewModel<EditScriptViewModel>()

            LaunchedEffect(Unit) {
                editScriptViewModel.init(node = node, name = name)
            }

            EditScriptCommandObserver(
                uiStateHolder = editScriptViewModel,
                navController = navController,
                snackbarHostState = snackbarHostState,
            )

            EditScriptScreen(
                uiStateHolder = editScriptViewModel,
                snackbarHostState = snackbarHostState,
            )
        }
    }
}
