package com.vision.scripter.streaming.impl

import androidx.activity.compose.LocalActivity
import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.lifecycle.viewmodel.compose.hiltViewModel
import androidx.navigation.NavController
import androidx.navigation.NavGraphBuilder
import androidx.navigation.NavType
import androidx.navigation.compose.composable
import androidx.navigation.navArgument
import com.vision.scripter.streaming.api.FeatureStreaming
import com.vision.scripter.streaming.api.StreamingRouteWithArgs
import com.vision.scripter.streaming.api.StreamingSerialArg
import com.vision.scripter.streaming.impl.screen.main.commandobservers.StreamingCommandObserver
import com.vision.scripter.streaming.impl.screen.main.state.StreamingViewModel
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingScreen
import dagger.hilt.android.scopes.ActivityScoped
import javax.inject.Inject

@ActivityScoped
class FeatureStreamingImpl @Inject constructor() : FeatureStreaming {

    override fun register(
        navGraphBuilder: NavGraphBuilder,
        navController: NavController
    ) {
        navGraphBuilder.composable(
            route = StreamingRouteWithArgs,
            arguments = listOf(navArgument(StreamingSerialArg) { type = NavType.StringType })
        ) { backStackEntry ->
            val activity = LocalActivity.current
            val view = LocalView.current

            DisposableEffect(Unit) {
                activity ?: return@DisposableEffect onDispose { }

                val controller = WindowInsetsControllerCompat(activity.window, view)
                controller.hide(
                    WindowInsetsCompat.Type.statusBars() or
                            WindowInsetsCompat.Type.navigationBars()
                )

                controller.systemBarsBehavior =
                    WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE

                onDispose {
                    controller.show(
                        WindowInsetsCompat.Type.statusBars() or
                                WindowInsetsCompat.Type.navigationBars()
                    )
                }
            }

            val serial = backStackEntry.arguments?.getString(StreamingSerialArg).orEmpty()
            val snackbarHostState = remember { SnackbarHostState() }
            val streamingViewModel = hiltViewModel<StreamingViewModel>()
            StreamingCommandObserver(
                uiStateHolder = streamingViewModel,
                navController = navController,
                snackbarHostState = snackbarHostState,
            )

            StreamingScreen(
                serial = serial,
                uiStateHolder = streamingViewModel,
                snackbarHostState = snackbarHostState,
            )
        }
    }
}