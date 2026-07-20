package com.vision.scripter.streaming.impl.blocks.menu.commandobservers

import androidx.compose.runtime.Composable
import com.vision.scripter.streaming.impl.blocks.menu.ui.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.ui.StreamingUiStateHolder
import com.vision.scripter.ui.observe

@Composable
fun MenuSharedCommandObserver(
    uiStateHolder: MenuUiStateHolder,
    sharedStateHolder: StreamingUiStateHolder,
) {
    sharedStateHolder.sharedEventsFlow.observe {
        when (it) {
            else -> {}
        }
    }
}