package com.vision.scripter.streaming.impl.blocks.menu.commandobservers

import androidx.compose.runtime.Composable
import androidx.navigation.NavController
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiCommand
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.ui.observe

@Composable
fun MenuCommandObserver(
    uiStateHolder: MenuUiStateHolder,
    navController: NavController,
) {
    uiStateHolder.uiCommandsFlow.observe {
        when (it) {
            is MenuUiCommand.ExitCommand -> {
                navController.popBackStack()
            }
        }
    }
}