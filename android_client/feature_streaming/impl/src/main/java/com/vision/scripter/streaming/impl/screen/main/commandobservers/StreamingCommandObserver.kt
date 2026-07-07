package com.vision.scripter.streaming.impl.screen.main.commandobservers

import android.widget.Toast
import androidx.compose.material3.SnackbarHostState
import androidx.compose.runtime.Composable
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import com.vision.scripter.streaming.impl.R
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiCommand
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe
import com.vision.scripter.ui.R as CommonR

@Composable
internal fun StreamingCommandObserver(
    uiStateHolder: StreamingUiStateHolder,
    snackbarHostState: SnackbarHostState,
) {
    val ctx = LocalContext.current
    val commonNetworkError = stringResource(CommonR.string.common_network_error)
    val stepSavedMessage = stringResource(R.string.step_saved)
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is StreamingUiCommand.ShowNetworkError -> {
                snackbarHostState.showSnackbar(commonNetworkError)
            }

            is StreamingUiCommand.ShowStepSavedSnackbar -> {
                Toast.makeText(
                    ctx,
                    stepSavedMessage,
                    Toast.LENGTH_SHORT
                ).show()
            }
        }
    }
}