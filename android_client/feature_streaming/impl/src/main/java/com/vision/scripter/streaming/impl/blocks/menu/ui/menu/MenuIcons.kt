package com.vision.scripter.streaming.impl.blocks.menu.ui.menu

import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Camera
import androidx.compose.material.icons.filled.Keyboard
import androidx.compose.material.icons.filled.TextFields
import androidx.compose.material.icons.filled.Visibility
import androidx.compose.material.icons.filled.VisibilityOff
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.dp
import com.vision.scripter.streaming.impl.blocks.menu.state.MenuUiStateHolder
import com.vision.scripter.streaming.impl.screen.main.state.CVMode
import com.vision.scripter.ui.customClickable

@Composable
fun BaseMenuIcons(
    cvMode: CVMode,
    textHighlighted: Boolean,
    keyboardHighlighted: Boolean,
    uiStateHolder: MenuUiStateHolder,
) {
    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = uiStateHolder::onCvModeClicked),
        imageVector = detectionIcon(cvMode),
        tint = if (cvMode != CVMode.NO_CV) Color.Red else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = uiStateHolder::onTextModeClicked),
        imageVector = Icons.Filled.TextFields,
        tint = if (textHighlighted) Color.Red else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )

    Icon(
        modifier = Modifier
            .size(32.dp)
            .customClickable(onClick = uiStateHolder::onKeyboardClicked),
        imageVector = Icons.Filled.Keyboard,
        tint = if (keyboardHighlighted) Color.Red else MaterialTheme.colorScheme.onSurface,
        contentDescription = "",
    )
}

fun detectionIcon(cvMode: CVMode): ImageVector = when (cvMode) {
    CVMode.YOLO -> Icons.Filled.Camera
    CVMode.CV_RECTS -> Icons.Filled.Visibility
    else -> Icons.Filled.VisibilityOff
}
